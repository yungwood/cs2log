# cs2log/stream

`stream` parses CS2 log streams from incrementally pushed log lines.

The root `cs2log` package parses one physical log line at a time. This package
adds stream-level behavior: line numbers, multiline block assembly, truncated
block detection, and typed events for block formats that need more than one log
line.

## Example

```go
parser, err := cs2log.NewParser(cs2log.Config{
	LogTimezone: "America/New_York",
})
if err != nil {
	return err
}

processor := stream.NewProcessor(parser)

scanner := bufio.NewScanner(reader)
scanner.Buffer(make([]byte, 64*1024), 1024*1024)
for scanner.Scan() {
	for _, record := range processor.PushLine(scanner.Text()) {
		handle(record)
	}
}
if err := scanner.Err(); err != nil {
	return err
}
for _, record := range processor.Flush() {
	handle(record)
}
```

## Records

Each processor result is a `Record`:

- `Event` is the parsed `cs2log.Event`.
- `LineStart` and `LineEnd` are inclusive physical line numbers.
- `Raw` contains retained source text for the line or joined block. It may be
  redacted for sensitive event types.
- `Err` is set when a physical line cannot be parsed or a multiline block is
  emitted because it exceeded processor limits.

## Incremental Processing

Use `Processor` when a caller already receives complete log lines from a live
source such as HTTP ingest, a socket, or a container log stream:

```go
processor := stream.NewProcessor(parser)

for line := range lines {
	for _, record := range processor.PushLine(line) {
		handle(record)
	}
}

for _, record := range processor.Flush() {
	handle(record)
}
```

`PushLine` can return zero records when a line is buffered into a multiline
block. It can return more than one record when a timestamp change closes a
truncated block and the current line also produces a record. `Flush` closes any
open block as truncated.

## Buffer Limits

`Processor` limits buffered multiline blocks to guard live ingest paths from
unbounded memory growth. Defaults are 1 MiB or 10,000 physical lines per block.

```go
processor := stream.NewProcessorWithConfig(parser, stream.ProcessorConfig{
	MaxBufferedBlockBytes: 1 << 20,
	MaxBufferedBlockLines: 10000,
})
```

When a block exceeds either limit, the processor emits the block as truncated
and sets `Record.Err` to `ErrBlockLimitExceeded`.

## JSON Blocks

Some CS2 logs emit JSON-style blocks where every physical line has the same
timestamp:

```text
09/02/2024 - 12:28:43.426 - JSON_BEGIN{
09/02/2024 - 12:28:43.426 - "name": "round_stats",
09/02/2024 - 12:28:43.426 - "players" : {
09/02/2024 - 12:28:43.426 - }}JSON_END
```

The stream processor collapses these into one record. Complete `round_stats`
blocks are parsed as `RoundStats` events. Unknown, malformed, or truncated
blocks remain generic `JSONBlock` events so callers can still inspect the raw
body and line range.

A JSON block is treated as truncated if the timestamp changes before
`}}JSON_END` is seen, or if the stream ends before the block closes. This makes
missing lines during replay easier to detect.

## Server Cvar Blocks

Server cvar dumps wrapped by `server cvars start` and `server cvars end` are
collapsed into one `ServerCvars` record. The stream processor aggregates parsed
`cs2log.ServerCvar` events from the root parser, so sensitive values and raw
lines use the same redaction path as single-line parsing.

Standalone `server_cvar:` lines outside a wrapped dump remain individual
`cs2log.ServerCvar` events.

## Test Command

Use the stream helper to inspect stream parsing behavior:

```sh
go run ./cmd/cs2log stream -timezone America/New_York /path/to/server.log
go run ./cmd/cs2log stream -timezone America/New_York -blocks /path/to/server.log
```
