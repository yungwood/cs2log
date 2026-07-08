# cs2log CLI

`cmd/cs2log` provides developer-oriented helpers for scanning CS2 log files,
checking parser coverage, and inspecting parsed events.

Run commands from the repository root:

```sh
go run ./cmd/cs2log <command> [options] [log ...]
```

If no log path is provided, commands read from stdin.

## coverage

Scan logs with the single-line parser and summarize parser coverage.

```sh
go run ./cmd/cs2log coverage -timezone America/New_York /path/to/server.log
```

Useful options:

- `-timezone`: IANA timezone used to interpret CS2 log timestamps.
- `-top`: number of unknown payload prefixes or errors to print.
- `-prefix`: maximum unknown payload prefix length.

`coverage` reports event counts, parse errors, ignored reasons, and common
unknown payload prefixes. It does not use stream/multiline processing.

## stream

Scan logs with `cs2log/stream` and summarize stream records.

```sh
go run ./cmd/cs2log stream -timezone America/New_York /path/to/server.log
go run ./cmd/cs2log stream -timezone America/New_York -blocks /path/to/server.log
```

Useful options:

- `-timezone`: IANA timezone used to interpret CS2 log timestamps.
- `-blocks`: print each multiline block record.

Use this when checking JSON blocks, `RoundStats`, and aggregated server cvar
blocks.

## inspect

Print parsed event records.

```sh
go run ./cmd/cs2log inspect -timezone America/New_York -type PlayerKill -format jsonl /path/to/server.log
go run ./cmd/cs2log inspect -timezone America/New_York -type PlayerKill,PlayerSay -limit 20 /path/to/server.log
go run ./cmd/cs2log inspect -stream -timezone America/New_York -type RoundStats -format jsonl /path/to/server.log
go run ./cmd/cs2log inspect -state -timezone America/New_York -type PlayerKill -format jsonl /path/to/server.log
```

Useful options:

- `-type`: event type to include. It can be repeated or comma-separated.
- `-format`: `text`, `json`, or `jsonl`.
- `-limit`: maximum matching events to print.
- `-raw`: include retained raw log text at the top level. Raw output may include
  player names or operational data, though known sensitive cvar values are
  redacted.
- `-stream`: use stream processing for multiline records.
- `-state`: include tracked match state context. This implies `-stream`.
- `-timezone`: IANA timezone used to interpret CS2 log timestamps.

When `-state` is enabled, all stream records update the matchstate tracker
before type filtering is applied, so filtered event output still receives the
best available context.
