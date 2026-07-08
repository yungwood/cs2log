# matchstate

`matchstate` tracks lightweight match context from parsed CS2 log events.

Use it after `stream.Processor` when callers want each parsed record annotated
with the current best-known map, score, round, warmup, pause, and game-over flag.
It also tracks a coarse lifecycle phase and key timestamps such as the last
event time, match start, game commencing, round start/end, and game over.
Round-ending team notices are normalized into stable reason strings such as
`bomb_defused`, while preserving the raw notice string for traceability.

## Semantics

- `Context` is the best-known state after applying the current record's event.
- Missing or omitted context fields mean the tracker does not know that value yet.
- Context is not retroactive. Earlier records are not corrected when later log
  lines reveal better map, round, score, or timing information.
- Errored records and records without parsed events do not update tracked state.

```go
parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "America/New_York"})
if err != nil {
	return err
}

processor := stream.NewProcessor(parser)
tracker := matchstate.NewTracker()

for _, record := range processor.PushLine(line) {
	enriched := tracker.Push(record)
	_ = enriched.Context
}
```

The tracker only updates from explicit parsed events. It does not infer durable
match identity, player sessions, or missing round numbers.
