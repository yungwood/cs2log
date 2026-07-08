package stream

import (
	"encoding/csv"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/yungwood/cs2log"
)

// RoundStats is emitted for round_stats JSON blocks.
type RoundStats struct {
	cs2log.BaseEvent
	RoundNumber int                `json:"roundNumber"`
	ScoreT      int                `json:"scoreT"`
	ScoreCT     int                `json:"scoreCT"`
	Map         string             `json:"map"`
	Server      string             `json:"server"`
	Fields      []string           `json:"fields"`
	Players     []RoundStatsPlayer `json:"players"`
	LineStart   int                `json:"lineStart"`
	LineEnd     int                `json:"lineEnd"`
}

// RoundStatsPlayer is one player row in a round_stats JSON block. Values keeps
// every stat column keyed by the block's fields header.
type RoundStatsPlayer struct {
	Key       string            `json:"key"`
	Slot      int               `json:"slot"`
	AccountID string            `json:"accountId,omitempty"`
	TeamID    cs2log.TeamID     `json:"teamId"`
	Side      cs2log.Side       `json:"side,omitempty"`
	Values    map[string]string `json:"values"`
}

func parseRoundStatsBlock(block JSONBlock) (RoundStats, bool) {
	if block.Status != JSONBlockComplete || !block.ValidJSON || block.Name != "round_stats" {
		return RoundStats{}, false
	}

	var raw struct {
		RoundNumber string            `json:"round_number"`
		ScoreT      string            `json:"score_t"`
		ScoreCT     string            `json:"score_ct"`
		Map         string            `json:"map"`
		Server      string            `json:"server"`
		Fields      string            `json:"fields"`
		Players     map[string]string `json:"players"`
	}
	if err := json.Unmarshal(block.Payload, &raw); err != nil {
		return RoundStats{}, false
	}

	fields := parseCSVFields(raw.Fields)
	return RoundStats{
		BaseEvent: cs2log.BaseEvent{
			Type:    "RoundStats",
			TimeUTC: block.Timestamp(),
			Raw:     block.RawLine(),
		},
		RoundNumber: atoi(raw.RoundNumber),
		ScoreT:      atoi(raw.ScoreT),
		ScoreCT:     atoi(raw.ScoreCT),
		Map:         raw.Map,
		Server:      raw.Server,
		Fields:      fields,
		Players:     parseRoundStatsPlayers(fields, raw.Players),
		LineStart:   block.LineStart,
		LineEnd:     block.LineEnd,
	}, true
}

func parseRoundStatsPlayers(fields []string, rows map[string]string) []RoundStatsPlayer {
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return roundStatsPlayerSlot(keys[i]) < roundStatsPlayerSlot(keys[j])
	})

	players := make([]RoundStatsPlayer, 0, len(keys))
	for _, key := range keys {
		values := mapRoundStatsValues(fields, parseCSVFields(rows[key]))
		teamID := cs2log.TeamID(atoi(values["team"]))
		side, _ := cs2log.SideFromTeamID(teamID)
		players = append(players, RoundStatsPlayer{
			Key:       key,
			Slot:      roundStatsPlayerSlot(key),
			AccountID: values["accountid"],
			TeamID:    teamID,
			Side:      side,
			Values:    values,
		})
	}
	return players
}

func mapRoundStatsValues(fields, values []string) map[string]string {
	mapped := make(map[string]string, len(fields))
	for i, field := range fields {
		// Keep the fields header authoritative so callers can distinguish a
		// missing trailing value from an omitted column.
		if i >= len(values) {
			mapped[field] = ""
			continue
		}
		mapped[field] = values[i]
	}
	return mapped
}

func parseCSVFields(value string) []string {
	reader := csv.NewReader(strings.NewReader(value))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	fields, err := reader.Read()
	if err != nil {
		return nil
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}

func roundStatsPlayerSlot(key string) int {
	slot, ok := strings.CutPrefix(key, "player_")
	if !ok {
		return 0
	}
	return atoi(slot)
}

func atoi(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}
