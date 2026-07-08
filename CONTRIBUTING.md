# Contributing

This project keeps parser behavior close to real CS2 server logs. Prefer adding
fixtures from fabricated or anonymized examples, not raw production logs.

## Project Layout

- root package: single-line parser, event types, shared values, Steam ID
  helpers, and team helpers.
- `stream`: incremental stream processing, line numbers, multiline JSON blocks,
  server cvar blocks, and `RoundStats`.
- `matchstate`: best-known match context derived from stream records.
- `cmd/cs2log`: developer CLI for coverage, stream inspection, and event output.

## Event Definitions

Events are defined in Go so public structs, parser metadata, examples, and build
logic stay together. Definitions are grouped by category in files such as
`definitions_chat.go`, `definitions_combat.go`, and `definitions_team.go`.

When adding an event:

1. Add the exported event struct with a short Go doc comment.
2. Add a `definition` with `Type`, `Category`, `Description`, fields, and a
   regex-backed build function.
3. Keep `Fields` in the same order as the regex capture groups.
4. Register the definition in `defaultDefinitions` in `definitions.go`.
5. Add parser tests with fabricated values.
6. Add or update matchstate behavior only when the event changes durable match
   context.

Definition examples are executable test data. They should use IANA timezone
names and expected UTC timestamps.

## Streams And State

Keep the root parser focused on one timestamped physical line. Use `stream` for
anything that depends on multiple lines, such as JSON blocks or grouped cvar
dumps.

`matchstate` is intentionally derived and non-retroactive. It should only track
best-known context after each parsed record. Errored records must not update
state.

## Sensitive Data

Do not commit real server startup logs or raw cvar dumps. Server cvar handling
must redact sensitive values before they can appear in event values, raw lines,
payload JSON, CLI output, or tests.

Raw log text may still contain player names or operational data. Tests and docs
should use fabricated names and fake identifiers unless a specific public test
fixture is intentional.

## Checks

Run these before committing:

```sh
golangci-lint run ./...
go test ./...
go vet ./...
```

When checking package shape, this is also useful:

```sh
go list -json ./...
```
