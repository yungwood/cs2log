# Architecture

`cs2log` is split into small packages with narrow responsibilities. Keep these
boundaries clear when adding parser behavior.

## Packages

- root package: parses one timestamped physical CS2 log line into a typed event.
- `stream`: incrementally processes log streams, adds line ranges, and groups
  multiline records.
- `matchstate`: derives best-known match context from parsed stream records.
- `cmd/cs2log`: developer CLI for parser coverage and inspection.

CLI code should call library APIs. Parser behavior belongs in the library
packages, not in `cmd/cs2log`.

## Parsing Model

CS2 log timestamps do not include a timezone. The root parser interprets them in
the configured game server host timezone using an IANA timezone name, then
normalizes event timestamps to UTC.

The parser distinguishes two failure modes:

- `ErrNoMatch`: the line has no supported CS2 timestamp/log wrapper.
- `Unknown`: the line has a supported timestamp, but the payload has no known
  event pattern.

Known log lines should become typed event structs. Avoid generic maps when the
line shape is known.

## Event Definitions

Event definitions live in `definitions_*.go` and are registered in
`defaultDefinitions`. Definitions combine parser metadata, fields, examples, and
build logic with the public event structs.

Use external projects and forks as coverage checklists, not as architecture
sources. Prefer adding events from observed logs, fabricated/anonymized fixtures,
or strong evidence.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the event workflow.

## Stream Processing

The root parser is intentionally single-line only. Use `stream.Processor` for
live ingest, replay files, and any behavior that depends on multiple physical
lines.

`stream.Processor` is incremental:

- callers own file/socket scanning
- `PushLine` may return zero, one, or multiple records
- `Flush` emits any open multiline block as truncated

Multiline JSON and server-cvar blocks are grouped by timestamp. Timestamp
changes, EOF, or buffer limits emit open blocks as truncated. Buffer limits are
intentional guardrails for live ingest paths.

## Match State

`matchstate` is derived context layered after `stream`. It tracks best-known
context after applying the current record. It is not retroactive: later records
do not rewrite earlier context.

Errored stream records and records without parsed events must not update
`matchstate`.

Only add fields to `matchstate.Context` when an event changes durable match
context, such as map, phase, score, round, pause state, team names, or
round-end state.

## Raw Text And Redaction

Events retain source text through `RawLine()` for debugging, coverage, and
inspection. Raw text may contain player names or operational data.

Sensitive server cvars must be redacted before values or raw text can escape
through events, stream records, payload JSON, CLI output, tests, or logs.

Do not commit real server startup logs, raw cvar dumps, secrets, tokens,
passwords, or private production logs.
