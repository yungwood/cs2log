package stream

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/yungwood/cs2log"
)

const (
	// DefaultMaxBufferedBlockBytes is the default maximum raw bytes kept for one
	// multiline block.
	DefaultMaxBufferedBlockBytes = 1 << 20
	// DefaultMaxBufferedBlockLines is the default maximum physical lines kept
	// for one multiline block.
	DefaultMaxBufferedBlockLines = 10000

	// BlockComplete means a multiline block was closed.
	BlockComplete BlockStatus = "complete"
	// BlockTruncated means a block ended before its closing marker, usually
	// because the timestamp changed or the stream ended.
	BlockTruncated BlockStatus = "truncated"

	// JSONBlockComplete means a JSON_BEGIN/JSON_END wrapper was closed.
	JSONBlockComplete JSONBlockStatus = JSONBlockStatus(BlockComplete)
	// JSONBlockTruncated means a block ended before JSON_END, usually because
	// the timestamp changed or the file ended.
	JSONBlockTruncated JSONBlockStatus = JSONBlockStatus(BlockTruncated)
)

var streamLogLinePattern = regexp.MustCompile(`^(?:L )?(\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}(?:\.\d{3})?)(?::| -) (.*)$`)

// ErrBlockLimitExceeded means a multiline block exceeded processor buffer
// limits and was emitted as truncated.
var ErrBlockLimitExceeded = errors.New("stream block limit exceeded")

// BlockStatus describes whether a multiline block was complete.
type BlockStatus string

// Record is one parsed stream record. Multiline records span LineStart to
// LineEnd. Raw contains retained source text and may be redacted for sensitive
// event types. Err is set when a line cannot be parsed or a block limit is hit.
type Record struct {
	Event     cs2log.Event
	Raw       string
	LineStart int
	LineEnd   int
	Err       error
}

// ProcessorConfig controls stream processor limits.
type ProcessorConfig struct {
	// MaxBufferedBlockBytes limits raw bytes held for one multiline block.
	// Values less than or equal to zero use DefaultMaxBufferedBlockBytes.
	MaxBufferedBlockBytes int
	// MaxBufferedBlockLines limits physical lines held for one multiline block.
	// Values less than or equal to zero use DefaultMaxBufferedBlockLines.
	MaxBufferedBlockLines int
}

// Processor incrementally converts physical CS2 log lines into stream records.
// It is useful for live ingest paths that already receive one line at a time.
type Processor struct {
	parser     *cs2log.Parser
	config     ProcessorConfig
	lineNumber int
	jsonBlock  *jsonBlockBuilder
	cvarBlock  *serverCvarsBlockBuilder
}

// NewProcessor creates an incremental stream processor with default limits.
func NewProcessor(parser *cs2log.Parser) *Processor {
	return NewProcessorWithConfig(parser, ProcessorConfig{})
}

// NewProcessorWithConfig creates an incremental stream processor with explicit
// limits. Zero or negative limit values use the package defaults.
func NewProcessorWithConfig(parser *cs2log.Parser, config ProcessorConfig) *Processor {
	return &Processor{parser: parser, config: normalizeProcessorConfig(config)}
}

func normalizeProcessorConfig(config ProcessorConfig) ProcessorConfig {
	if config.MaxBufferedBlockBytes <= 0 {
		config.MaxBufferedBlockBytes = DefaultMaxBufferedBlockBytes
	}
	if config.MaxBufferedBlockLines <= 0 {
		config.MaxBufferedBlockLines = DefaultMaxBufferedBlockLines
	}
	return config
}

// PushLine processes one physical log line. It returns zero records when the
// line is buffered into an open multiline block.
func (p *Processor) PushLine(line string) []Record {
	p.lineNumber++
	return p.processLine(p.parseLine(p.lineNumber, line))
}

// Flush closes any buffered multiline block as truncated. Call this at EOF or
// shutdown for live streams.
func (p *Processor) Flush() []Record {
	if p.jsonBlock != nil {
		record := p.jsonBlock.record()
		p.jsonBlock = nil
		return []Record{record}
	}
	if p.cvarBlock != nil {
		record := p.cvarBlock.record()
		p.cvarBlock = nil
		return []Record{record}
	}
	return nil
}

func (p *Processor) processLine(line parsedLine) []Record {
	if p.jsonBlock != nil {
		return p.processJSONBlockLine(line)
	}
	if p.cvarBlock != nil {
		return p.processServerCvarsBlockLine(line)
	}

	if strings.TrimSpace(line.payload) == "JSON_BEGIN{" && line.err == nil {
		p.jsonBlock = newJSONBlockBuilder(line)
		if p.jsonBlock.exceedsLimit(p.config) {
			record := p.jsonBlock.record()
			record.Err = ErrBlockLimitExceeded
			p.jsonBlock = nil
			return []Record{record}
		}
		return nil
	}
	if isServerCvarsMarker(line.event, "server cvars start") {
		p.cvarBlock = newServerCvarsBlockBuilder(line)
		if p.cvarBlock.exceedsLimit(p.config) {
			record := p.cvarBlock.record()
			record.Err = ErrBlockLimitExceeded
			p.cvarBlock = nil
			return []Record{record}
		}
		return nil
	}

	if line.err != nil {
		return []Record{{Raw: line.raw, LineStart: line.number, LineEnd: line.number, Err: line.err}}
	}

	return []Record{{Event: line.event, Raw: line.raw, LineStart: line.number, LineEnd: line.number}}
}

func (p *Processor) processJSONBlockLine(line parsedLine) []Record {
	if line.err != nil {
		p.jsonBlock.append(line)
		record := p.jsonBlock.record()
		p.jsonBlock = nil
		return []Record{record}
	}

	if !line.timestamp.Equal(p.jsonBlock.timestamp) {
		// JSON blocks are emitted with one timestamp. A timestamp change means
		// the previous block was truncated and this line starts normal parsing.
		record := p.jsonBlock.record()
		p.jsonBlock = nil
		records := []Record{record}
		records = append(records, p.processLine(line)...)
		return records
	}

	p.jsonBlock.append(line)
	if p.jsonBlock.exceedsLimit(p.config) {
		record := p.jsonBlock.record()
		record.Err = ErrBlockLimitExceeded
		p.jsonBlock = nil
		return []Record{record}
	}
	if strings.TrimSpace(line.payload) == "}}JSON_END" {
		p.jsonBlock.status = JSONBlockComplete
		record := p.jsonBlock.record()
		p.jsonBlock = nil
		return []Record{record}
	}
	return nil
}

func (p *Processor) processServerCvarsBlockLine(line parsedLine) []Record {
	if line.err != nil {
		p.cvarBlock.appendRaw(line)
		record := p.cvarBlock.record()
		p.cvarBlock = nil
		return []Record{record}
	}

	if !line.timestamp.Equal(p.cvarBlock.timestamp) {
		// Server cvar dumps share one timestamp, so a new timestamp closes the
		// current block even if the explicit end marker was not seen.
		record := p.cvarBlock.record()
		p.cvarBlock = nil
		records := []Record{record}
		records = append(records, p.processLine(line)...)
		return records
	}

	if isServerCvarsMarker(line.event, "server cvars end") {
		p.cvarBlock.appendRaw(line)
		p.cvarBlock.status = BlockComplete
		record := p.cvarBlock.record()
		p.cvarBlock = nil
		return []Record{record}
	}

	if cvar, ok := line.event.(cs2log.ServerCvar); ok {
		p.cvarBlock.appendCvar(line, cvar)
		if p.cvarBlock.exceedsLimit(p.config) {
			record := p.cvarBlock.record()
			record.Err = ErrBlockLimitExceeded
			p.cvarBlock = nil
			return []Record{record}
		}
		return nil
	}

	p.cvarBlock.appendRaw(line)
	if p.cvarBlock.exceedsLimit(p.config) {
		record := p.cvarBlock.record()
		record.Err = ErrBlockLimitExceeded
		p.cvarBlock = nil
		return []Record{record}
	}
	return nil
}

func (p *Processor) parseLine(number int, raw string) parsedLine {
	result := streamLogLinePattern.FindStringSubmatch(raw)
	if result == nil {
		return parsedLine{number: number, raw: raw, err: cs2log.ErrNoMatch}
	}

	timestamp, err := parseTimestampInLocation(result[1], p.parser.Location())
	if err != nil {
		return parsedLine{number: number, raw: raw, payload: result[2], err: err}
	}
	event, err := p.parser.ParseLine(raw)
	return parsedLine{
		number:    number,
		raw:       raw,
		payload:   result[2],
		timestamp: timestamp.UTC(),
		event:     event,
		err:       err,
	}
}

func parseTimestampInLocation(value string, location *time.Location) (time.Time, error) {
	for _, layout := range []string{"01/02/2006 - 15:04:05.000", "01/02/2006 - 15:04:05"} {
		parsed, err := time.ParseInLocation(layout, value, location)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse cs2 log timestamp %q", value)
}

type parsedLine struct {
	number    int
	raw       string
	payload   string
	timestamp time.Time
	event     cs2log.Event
	err       error
}

func safeRawLine(line parsedLine) string {
	// Parsed events may redact sensitive raw data before exposing RawLine.
	if line.event != nil {
		return line.event.RawLine()
	}
	return line.raw
}

// IsNoMatch reports whether err means a line had no supported CS2 timestamp.
func IsNoMatch(err error) bool {
	return errors.Is(err, cs2log.ErrNoMatch)
}
