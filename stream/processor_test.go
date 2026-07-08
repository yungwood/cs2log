package stream

import (
	"errors"
	"strings"
	"testing"

	"github.com/yungwood/cs2log"
)

func TestProcessorReturnsRoundStatsEvent(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	records := processLines(parser, []string{
		`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`,
		`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`,
		`L 07/05/2026 - 00:00:08.000: "round_number" : "2",`,
		`L 07/05/2026 - 00:00:08.000: "score_t" : "1",`,
		`L 07/05/2026 - 00:00:08.000: "score_ct" : "0",`,
		`L 07/05/2026 - 00:00:08.000: "map" : "cs_agency",`,
		`L 07/05/2026 - 00:00:08.000: "server" : "Test Server",`,
		`L 07/05/2026 - 00:00:08.000: "fields" : " accountid, team, money, kills"`,
		`L 07/05/2026 - 00:00:08.000: "players": {`,
		`L 07/05/2026 - 00:00:08.000: "player_0": "123456789, 3, 16000, 1"`,
		`L 07/05/2026 - 00:00:08.000: }}JSON_END`,
		`L 07/05/2026 - 00:00:09.000: World triggered "Round_Start"`,
	})
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2: %#v", len(records), records)
	}
	record := records[0]
	roundStats, ok := record.Event.(RoundStats)
	if !ok {
		t.Fatalf("event type = %T, want RoundStats", record.Event)
	}
	if roundStats.RoundNumber != 2 || roundStats.ScoreT != 1 || roundStats.ScoreCT != 0 || roundStats.Map != "cs_agency" {
		t.Fatalf("round stats = %#v", roundStats)
	}
	if len(roundStats.Players) != 1 {
		t.Fatalf("players = %#v", roundStats.Players)
	}
	if roundStats.Players[0].AccountID != "123456789" || roundStats.Players[0].TeamID != cs2log.TeamIDCT || roundStats.Players[0].Side != cs2log.SideCT || roundStats.Players[0].Values["kills"] != "1" {
		t.Fatalf("player = %#v", roundStats.Players[0])
	}
	if record.LineStart != 1 || record.LineEnd != 11 {
		t.Fatalf("lines = %d-%d, want 1-11", record.LineStart, record.LineEnd)
	}

	if records[1].LineStart != 12 {
		t.Fatalf("second line = %d, want 12", records[1].LineStart)
	}
}

func TestProcessorReturnsTruncatedJSONBlockWhenTimestampChanges(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	records := processLines(parser, []string{
		`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`,
		`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`,
		`L 07/05/2026 - 00:00:09.000: World triggered "Round_Start"`,
	})
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2: %#v", len(records), records)
	}
	block, ok := records[0].Event.(JSONBlock)
	if !ok {
		t.Fatalf("event type = %T, want JSONBlock", records[0].Event)
	}
	if block.Status != JSONBlockTruncated {
		t.Fatalf("status = %s, want %s", block.Status, JSONBlockTruncated)
	}
	if records[0].LineStart != 1 || records[0].LineEnd != 2 {
		t.Fatalf("lines = %d-%d, want 1-2", records[0].LineStart, records[0].LineEnd)
	}

	if records[1].LineStart != 3 {
		t.Fatalf("pending line = %d, want 3", records[1].LineStart)
	}
}

func TestProcessorReturnsRecordsIncrementally(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`)
	if len(records) != 0 {
		t.Fatalf("records after begin = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`)
	if len(records) != 0 {
		t.Fatalf("records after body = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "round_number" : "2",`)
	if len(records) != 0 {
		t.Fatalf("records after round = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "score_t" : "1",`)
	if len(records) != 0 {
		t.Fatalf("records after score t = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "score_ct" : "0",`)
	if len(records) != 0 {
		t.Fatalf("records after score ct = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "fields" : " accountid, team"`)
	if len(records) != 0 {
		t.Fatalf("records after fields = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "players": {`)
	if len(records) != 0 {
		t.Fatalf("records after players = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: "player_0": "0, 3"`)
	if len(records) != 0 {
		t.Fatalf("records after player = %#v", records)
	}
	records = processor.PushLine(`L 07/05/2026 - 00:00:08.000: }}JSON_END`)
	if len(records) != 1 {
		t.Fatalf("records after end = %d, want 1: %#v", len(records), records)
	}
	if _, ok := records[0].Event.(RoundStats); !ok {
		t.Fatalf("event type = %T, want RoundStats", records[0].Event)
	}
	if flushed := processor.Flush(); len(flushed) != 0 {
		t.Fatalf("flush records = %#v", flushed)
	}
}

func TestProcessorReturnsNoMatchErrorForInvalidLine(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	records := processor.PushLine(`not a cs2 log line`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	if !IsNoMatch(records[0].Err) {
		t.Fatalf("err = %v, want no match", records[0].Err)
	}
	if IsNoMatch(errors.New("different error")) {
		t.Fatal("different error reported as no match")
	}
}

func TestProcessorClosesJSONBlockOnInvalidLine(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`)
	records := processor.PushLine(`not a cs2 log line`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	block, ok := records[0].Event.(JSONBlock)
	if !ok {
		t.Fatalf("event type = %T, want JSONBlock", records[0].Event)
	}
	if block.Status != JSONBlockTruncated || records[0].LineEnd != 2 {
		t.Fatalf("block = %#v record = %#v", block, records[0])
	}
}

func TestProcessorReturnsTruncatedBlockAndCurrentLine(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`)
	processor.PushLine(`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`)
	records := processor.PushLine(`L 07/05/2026 - 00:00:09.000: World triggered "Round_Start"`)
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2: %#v", len(records), records)
	}
	block, ok := records[0].Event.(JSONBlock)
	if !ok {
		t.Fatalf("first event type = %T, want JSONBlock", records[0].Event)
	}
	if block.Status != JSONBlockTruncated {
		t.Fatalf("status = %s, want %s", block.Status, JSONBlockTruncated)
	}
	if records[1].LineStart != 3 || records[1].Event.EventType() != "WorldRoundStart" {
		t.Fatalf("second record = %#v", records[1])
	}
}

func TestProcessorFlushReturnsTruncatedBlock(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`)
	records := processor.Flush()
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	block, ok := records[0].Event.(JSONBlock)
	if !ok {
		t.Fatalf("event type = %T, want JSONBlock", records[0].Event)
	}
	if block.Status != JSONBlockTruncated {
		t.Fatalf("status = %s, want %s", block.Status, JSONBlockTruncated)
	}
}

func TestProcessorTruncatesJSONBlockWhenLineLimitExceeded(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessorWithConfig(parser, ProcessorConfig{MaxBufferedBlockLines: 2})

	if records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`); len(records) != 0 {
		t.Fatalf("records after begin = %#v", records)
	}
	if records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`); len(records) != 0 {
		t.Fatalf("records after body = %#v", records)
	}
	records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: "round_number" : "2",`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	if !errors.Is(records[0].Err, ErrBlockLimitExceeded) {
		t.Fatalf("err = %v, want %v", records[0].Err, ErrBlockLimitExceeded)
	}
	block, ok := records[0].Event.(JSONBlock)
	if !ok {
		t.Fatalf("event type = %T, want JSONBlock", records[0].Event)
	}
	if block.Status != JSONBlockTruncated || records[0].LineEnd != 3 {
		t.Fatalf("block = %#v record = %#v", block, records[0])
	}
}

func TestProcessorTruncatesJSONBlockWhenByteLimitExceeded(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessorWithConfig(parser, ProcessorConfig{MaxBufferedBlockBytes: 10})

	records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	if !errors.Is(records[0].Err, ErrBlockLimitExceeded) {
		t.Fatalf("err = %v, want %v", records[0].Err, ErrBlockLimitExceeded)
	}
}

func TestProcessorReturnsServerCvarsBlock(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	records := processLines(parser, []string{
		`L 07/05/2026 - 00:00:08.000: server cvars start`,
		`L 07/05/2026 - 00:00:08.000: "mp_freezetime" = "12"`,
		`L 07/05/2026 - 00:00:08.000: "sv_password" = "fake-secret-password"`,
		`L 07/05/2026 - 00:00:08.000: server cvars end`,
	})
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	cvars, ok := records[0].Event.(ServerCvars)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvars", records[0].Event)
	}
	if cvars.Status != BlockComplete || cvars.LineStart != 1 || cvars.LineEnd != 4 {
		t.Fatalf("cvars = %#v", cvars)
	}
	if len(cvars.Cvars) != 2 {
		t.Fatalf("cvars length = %d, want 2: %#v", len(cvars.Cvars), cvars.Cvars)
	}
	if cvars.Cvars[0].Name != "mp_freezetime" || cvars.Cvars[0].Value != "12" || cvars.Cvars[0].Sensitive {
		t.Fatalf("first cvar = %#v", cvars.Cvars[0])
	}
	if cvars.Cvars[1].Name != "sv_password" || cvars.Cvars[1].Value != "" || !cvars.Cvars[1].Sensitive {
		t.Fatalf("second cvar = %#v", cvars.Cvars[1])
	}
	if strings.Contains(cvars.RawLine(), "fake-secret-password") {
		t.Fatalf("raw line contains secret: %s", cvars.RawLine())
	}
	if !strings.Contains(cvars.RawLine(), `"`+"sv_password"+`" = "[REDACTED]"`) {
		t.Fatalf("raw line missing redaction: %s", cvars.RawLine())
	}
}

func TestProcessorClosesServerCvarsBlockOnInvalidLine(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: server cvars start`)
	records := processor.PushLine(`not a cs2 log line`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	cvars, ok := records[0].Event.(ServerCvars)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvars", records[0].Event)
	}
	if cvars.Status != BlockTruncated || records[0].LineEnd != 2 {
		t.Fatalf("cvars = %#v record = %#v", cvars, records[0])
	}
}

func TestProcessorReturnsTruncatedServerCvarsBlockWhenTimestampChanges(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	records := processLines(parser, []string{
		`L 07/05/2026 - 00:00:08.000: server cvars start`,
		`L 07/05/2026 - 00:00:08.000: "mp_freezetime" = "12"`,
		`L 07/05/2026 - 00:00:09.000: World triggered "Round_Start"`,
	})
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2: %#v", len(records), records)
	}
	cvars, ok := records[0].Event.(ServerCvars)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvars", records[0].Event)
	}
	if cvars.Status != BlockTruncated || len(cvars.Cvars) != 1 {
		t.Fatalf("cvars = %#v", cvars)
	}
	if records[1].Event.EventType() != "WorldRoundStart" {
		t.Fatalf("second record = %#v", records[1])
	}
}

func TestProcessorFlushReturnsTruncatedServerCvarsBlock(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessor(parser)

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: server cvars start`)
	records := processor.Flush()
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	cvars, ok := records[0].Event.(ServerCvars)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvars", records[0].Event)
	}
	if cvars.Status != BlockTruncated {
		t.Fatalf("status = %s, want %s", cvars.Status, BlockTruncated)
	}
}

func TestProcessorTruncatesServerCvarsBlockWhenLineLimitExceeded(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessorWithConfig(parser, ProcessorConfig{MaxBufferedBlockLines: 2})

	processor.PushLine(`L 07/05/2026 - 00:00:08.000: server cvars start`)
	processor.PushLine(`L 07/05/2026 - 00:00:08.000: "mp_freezetime" = "12"`)
	records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: "sv_password" = "fake-secret-password"`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	if !errors.Is(records[0].Err, ErrBlockLimitExceeded) {
		t.Fatalf("err = %v, want %v", records[0].Err, ErrBlockLimitExceeded)
	}
	cvars, ok := records[0].Event.(ServerCvars)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvars", records[0].Event)
	}
	if cvars.Status != BlockTruncated || len(cvars.Cvars) != 2 {
		t.Fatalf("cvars = %#v", cvars)
	}
	if strings.Contains(cvars.RawLine(), "fake-secret-password") {
		t.Fatalf("raw line contains secret: %s", cvars.RawLine())
	}
}

func TestProcessorTruncatesServerCvarsBlockWhenByteLimitExceeded(t *testing.T) {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	processor := NewProcessorWithConfig(parser, ProcessorConfig{MaxBufferedBlockBytes: 10})

	records := processor.PushLine(`L 07/05/2026 - 00:00:08.000: server cvars start`)
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1: %#v", len(records), records)
	}
	if !errors.Is(records[0].Err, ErrBlockLimitExceeded) {
		t.Fatalf("err = %v, want %v", records[0].Err, ErrBlockLimitExceeded)
	}
}

func processLines(parser *cs2log.Parser, lines []string) []Record {
	processor := NewProcessor(parser)
	var records []Record
	for _, line := range lines {
		records = append(records, processor.PushLine(line)...)
	}
	records = append(records, processor.Flush()...)
	return records
}
