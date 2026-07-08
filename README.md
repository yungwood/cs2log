# cs2log

`cs2log` parses Counter-Strike 2 server logs into typed, timezone-safe Go
events.

CS2 log timestamps are written in the game server host's local timezone and do
not include timezone data. Configure the game server host timezone with an IANA
timezone name; parsed event timestamps are normalized to UTC.

API documentation is available on
[pkg.go.dev](https://pkg.go.dev/github.com/yungwood/cs2log).

## Install

```sh
go get github.com/yungwood/cs2log
```

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/yungwood/cs2log"
)

func main() {
	parser, err := cs2log.NewParser(cs2log.Config{
		LogTimezone: "America/New_York",
	})
	if err != nil {
		panic(err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`)
	if err != nil {
		panic(err)
	}

	fmt.Println(event.EventType())
	fmt.Println(event.Timestamp().Format(time.RFC3339))
	// PlayerSay
	// 2026-07-05T04:00:00Z
}
```

## Timezones

Set `LogTimezone` to the timezone configured on the CS2 game server host, not
necessarily the timezone of the app, container, or machine doing the parsing.

Use IANA timezone names such as `America/New_York`, `Europe/London`, or
`Australia/Sydney` so daylight saving time is handled correctly.

## Packages

- `cs2log`: single-line parser, event types, Steam ID helpers, and shared
  values.
- `stream`: incremental stream processing, line numbers, multiline JSON blocks,
  server cvar blocks, and typed `RoundStats` events.
- `matchstate`: best-known match context derived from stream records.
- `cmd/cs2log`: developer CLI for coverage checks and event inspection.

See [`stream/README.md`](stream/README.md),
[`matchstate/README.md`](matchstate/README.md), and
[`cmd/cs2log/README.md`](cmd/cs2log/README.md) for package-specific usage.

## Error Handling

- `ErrNoMatch`: the line does not have a supported CS2 timestamp/log wrapper.
- `Unknown`: the line has a supported timestamp, but the payload has no known
  event pattern.

## Raw Log Text

Events retain source text through `RawLine()` for debugging, coverage, and
inspection. Raw text may contain player names or operational data.

Known sensitive server cvar values are redacted before they can appear in event
values, raw lines, stream records, or CLI output.

## CLI

Use the CLI helpers to scan replay logs and inspect parser behavior:

```sh
go run ./cmd/cs2log coverage -timezone America/New_York /path/to/server.log
```

See [`cmd/cs2log/README.md`](cmd/cs2log/README.md) for `coverage`, `stream`,
and `inspect` usage.

## Development

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for package boundaries and parser
semantics. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for event definition,
fixture, and testing guidance.

## Acknowledgements

This project was informed by prior CS2 log parsing work, especially
[janstuemmel/cs2-log](https://github.com/janstuemmel/cs2-log). `cs2log` takes a
separate implementation approach with explicit timezone handling,
stream/multiline processing, sensitive cvar redaction, and match-state tracking.
