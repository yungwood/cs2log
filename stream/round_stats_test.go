package stream

import (
	"testing"
	"time"

	"github.com/yungwood/cs2log"
)

func TestParseRoundStatsBlockRejectsUnsupportedBlocks(t *testing.T) {
	block := JSONBlock{
		BaseEvent: cs2log.BaseEvent{
			Type:    "JSONBlock",
			TimeUTC: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC),
			Raw:     `{"name":"round_stats"}`,
		},
		Name:      "round_stats",
		Status:    JSONBlockTruncated,
		ValidJSON: true,
		Payload:   []byte(`{"name":"round_stats"}`),
	}
	if _, ok := parseRoundStatsBlock(block); ok {
		t.Fatal("truncated block parsed as round stats")
	}

	block.Status = JSONBlockComplete
	block.ValidJSON = false
	if _, ok := parseRoundStatsBlock(block); ok {
		t.Fatal("invalid JSON block parsed as round stats")
	}

	block.ValidJSON = true
	block.Name = "other"
	if _, ok := parseRoundStatsBlock(block); ok {
		t.Fatal("non-round_stats block parsed as round stats")
	}

	block.Name = "round_stats"
	block.Payload = []byte(`{`)
	if _, ok := parseRoundStatsBlock(block); ok {
		t.Fatal("malformed JSON payload parsed as round stats")
	}
}

func TestRoundStatsHelpersHandleMalformedValues(t *testing.T) {
	fields := []string{"accountid", "team", "money", "kills"}
	values := mapRoundStatsValues(fields, []string{"123456789", "2"})
	if values["accountid"] != "123456789" || values["team"] != "2" || values["money"] != "" || values["kills"] != "" {
		t.Fatalf("values = %#v", values)
	}

	if got := parseCSVFields(`"unterminated`); got != nil {
		t.Fatalf("parseCSVFields returned %#v, want nil", got)
	}
	if got := roundStatsPlayerSlot("unknown"); got != 0 {
		t.Fatalf("roundStatsPlayerSlot = %d, want 0", got)
	}
	if got := atoi("not-an-int"); got != 0 {
		t.Fatalf("atoi = %d, want 0", got)
	}
}
