# AGENTS.md

## Scope

These instructions apply to the whole repository.

## Required Reading

Read these before making non-trivial changes:

- [`ARCHITECTURE.md`](ARCHITECTURE.md)
- [`CONTRIBUTING.md`](CONTRIBUTING.md)

## Agent Rules

- Preserve the package boundaries described in `ARCHITECTURE.md`.
- Do not move parser behavior into `cmd/cs2log`.
- Do not add speculative events only because an external fork has them.
- Do not commit real logs, secrets, tokens, passwords, or raw cvar dumps.
- Preserve sensitive cvar redaction and raw-text safety behavior.
- Use fabricated or anonymized values in tests and docs.

## Go Style

- Keep APIs small, explicit, and boring.
- Avoid option-heavy abstractions or new dependencies unless they solve a real
  problem.
- Keep exported comments terse and Go-doc friendly.
- Use structured parsing/helpers over ad hoc string manipulation when practical.
- Keep tests close to changed behavior; table-driven tests are preferred for log
  variants.
- Public examples should use broadly recognizable timezones such as
  `America/New_York`; keep half-hour timezone coverage, such as
  `Australia/Adelaide`, in tests.

## Checks

Run these before finishing changes:

```sh
golangci-lint run ./...
go test ./...
go vet ./...
```

For package-shape sanity checks, also use:

```sh
go list -json ./...
```
