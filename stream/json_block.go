package stream

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/yungwood/cs2log"
)

// JSONBlockStatus describes whether a multiline JSON-style block was complete.
type JSONBlockStatus BlockStatus

// JSONBlock is emitted for multiline JSON_BEGIN/JSON_END blocks.
type JSONBlock struct {
	cs2log.BaseEvent
	Name           string          `json:"name,omitempty"`
	Status         JSONBlockStatus `json:"status"`
	ValidJSON      bool            `json:"validJson"`
	Body           string          `json:"body"`
	NormalizedBody string          `json:"normalizedBody,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	LineStart      int             `json:"lineStart"`
	LineEnd        int             `json:"lineEnd"`
	Lines          []string        `json:"-"`
}

type jsonBlockBuilder struct {
	timestamp time.Time
	lineStart int
	lineEnd   int
	rawLines  []string
	payloads  []string
	status    JSONBlockStatus
	bytes     int
}

func newJSONBlockBuilder(first parsedLine) *jsonBlockBuilder {
	return &jsonBlockBuilder{
		timestamp: first.timestamp,
		lineStart: first.number,
		lineEnd:   first.number,
		rawLines:  []string{first.raw},
		payloads:  []string{first.payload},
		status:    JSONBlockTruncated,
		bytes:     len(first.raw),
	}
}

func (b *jsonBlockBuilder) append(line parsedLine) {
	b.lineEnd = line.number
	b.rawLines = append(b.rawLines, line.raw)
	b.payloads = append(b.payloads, line.payload)
	b.bytes += len(line.raw)
}

func (b jsonBlockBuilder) exceedsLimit(config ProcessorConfig) bool {
	return len(b.rawLines) > config.MaxBufferedBlockLines || b.bytes > config.MaxBufferedBlockBytes
}

func (b jsonBlockBuilder) record() Record {
	body := jsonBlockBody(b.payloads)
	normalizedBody := normalizeJSONBlockBody(body)
	validJSON := json.Valid([]byte(normalizedBody))
	name := jsonBlockName(normalizedBody)
	block := JSONBlock{
		BaseEvent: cs2log.BaseEvent{
			Type:    "JSONBlock",
			TimeUTC: b.timestamp,
			Raw:     strings.Join(b.rawLines, "\n"),
		},
		Name:           name,
		Status:         b.status,
		ValidJSON:      validJSON,
		Body:           body,
		NormalizedBody: normalizedBody,
		LineStart:      b.lineStart,
		LineEnd:        b.lineEnd,
		Lines:          append([]string(nil), b.rawLines...),
	}
	if block.ValidJSON {
		block.Payload = json.RawMessage(normalizedBody)
	}

	event := cs2log.Event(block)
	if roundStats, ok := parseRoundStatsBlock(block); ok {
		// Known JSON block shapes are promoted to typed events while retaining
		// the original block raw line and line span.
		event = roundStats
	}
	return Record{
		Event:     event,
		Raw:       block.RawLine(),
		LineStart: b.lineStart,
		LineEnd:   b.lineEnd,
	}
}

func jsonBlockBody(payloads []string) string {
	lines := make([]string, 0, len(payloads))
	for _, payload := range payloads {
		payload = strings.TrimSpace(payload)
		switch payload {
		case "JSON_BEGIN{":
			lines = append(lines, "{")
		case "}}JSON_END":
			lines = append(lines, "}}")
		default:
			lines = append(lines, payload)
		}
	}
	return strings.Join(lines, "\n")
}

func normalizeJSONBlockBody(body string) string {
	lines := strings.Split(body, "\n")
	for i := 0; i < len(lines)-1; i++ {
		current := strings.TrimSpace(lines[i])
		next := strings.TrimSpace(lines[i+1])
		if current == "" || next == "" {
			continue
		}
		if current == "{" || current == "}" || strings.HasSuffix(current, "{") || strings.HasSuffix(current, ",") {
			continue
		}
		if strings.HasPrefix(current, `"`) && strings.HasPrefix(next, `"`) {
			// CS2 emits adjacent object fields without commas in some dumps.
			lines[i] += ","
		}
	}
	return strings.Join(lines, "\n")
}

func jsonBlockName(body string) string {
	var decoded struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err == nil && decoded.Name != "" {
		return decoded.Name
	}

	namePattern := regexp.MustCompile(`(?m)"name"\s*:\s*"([^"]+)"`)
	result := namePattern.FindStringSubmatch(body)
	if result == nil {
		return ""
	}
	return result[1]
}
