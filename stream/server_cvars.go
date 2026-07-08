package stream

import (
	"strings"
	"time"

	"github.com/yungwood/cs2log"
)

// ServerCvars is emitted for server cvar dump blocks. Cvars are aggregated from
// root parser ServerCvar events so sensitive values and raw lines stay redacted.
type ServerCvars struct {
	cs2log.BaseEvent
	Status    BlockStatus         `json:"status"`
	Cvars     []cs2log.ServerCvar `json:"cvars"`
	LineStart int                 `json:"lineStart"`
	LineEnd   int                 `json:"lineEnd"`
	Lines     []string            `json:"-"`
}

type serverCvarsBlockBuilder struct {
	timestamp time.Time
	lineStart int
	lineEnd   int
	rawLines  []string
	cvars     []cs2log.ServerCvar
	status    BlockStatus
	bytes     int
}

func newServerCvarsBlockBuilder(first parsedLine) *serverCvarsBlockBuilder {
	return &serverCvarsBlockBuilder{
		timestamp: first.timestamp,
		lineStart: first.number,
		lineEnd:   first.number,
		rawLines:  []string{safeRawLine(first)},
		status:    BlockTruncated,
		bytes:     len(safeRawLine(first)),
	}
}

func (b *serverCvarsBlockBuilder) appendRaw(line parsedLine) {
	raw := safeRawLine(line)
	b.lineEnd = line.number
	b.rawLines = append(b.rawLines, raw)
	b.bytes += len(raw)
}

func (b *serverCvarsBlockBuilder) appendCvar(line parsedLine, cvar cs2log.ServerCvar) {
	raw := cvar.RawLine()
	b.lineEnd = line.number
	b.rawLines = append(b.rawLines, raw)
	b.cvars = append(b.cvars, cvar)
	b.bytes += len(raw)
}

func (b serverCvarsBlockBuilder) exceedsLimit(config ProcessorConfig) bool {
	return len(b.rawLines) > config.MaxBufferedBlockLines || b.bytes > config.MaxBufferedBlockBytes
}

func (b serverCvarsBlockBuilder) record() Record {
	event := ServerCvars{
		BaseEvent: cs2log.BaseEvent{
			Type:    "ServerCvars",
			TimeUTC: b.timestamp,
			Raw:     strings.Join(b.rawLines, "\n"),
		},
		Status:    b.status,
		Cvars:     append([]cs2log.ServerCvar(nil), b.cvars...),
		LineStart: b.lineStart,
		LineEnd:   b.lineEnd,
		Lines:     append([]string(nil), b.rawLines...),
	}
	return Record{
		Event:     event,
		Raw:       event.RawLine(),
		LineStart: b.lineStart,
		LineEnd:   b.lineEnd,
	}
}

func isServerCvarsMarker(event cs2log.Event, payload string) bool {
	ignored, ok := event.(cs2log.Ignored)
	return ok && ignored.Reason == "ServerCvarsMarker" && ignored.Payload == payload
}
