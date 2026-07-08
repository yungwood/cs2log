package cs2log

// GameOver is emitted when the server reports a game-over line.
type GameOver struct {
	BaseEvent
	Mode     string `json:"mode"`
	MapGroup string `json:"mapGroup"`
	Map      string `json:"map"`
	ScoreCT  int    `json:"scoreCT"`
	ScoreT   int    `json:"scoreT"`
	Duration int    `json:"duration"`
}

// MatchStatusScore is emitted when a match status dump reports the score.
type MatchStatusScore struct {
	BaseEvent
	ScoreCT      int    `json:"scoreCT"`
	ScoreT       int    `json:"scoreT"`
	Map          string `json:"map"`
	RoundsPlayed int    `json:"roundsPlayed"`
}

// MatchPause is emitted when match pause state changes.
type MatchPause struct {
	BaseEvent
	Action string `json:"action"`
}

var matchPauseDefinition = definition{
	Type:        "MatchPause",
	Category:    "game",
	Description: "Match pause state changed.",
	Regex:       `Match (pause is enabled|pause is disabled|unpaused)`,
	Fields:      []field{{Name: "actionText", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		action := ""
		switch matches.String("actionText") {
		case "pause is enabled":
			action = "enabled"
		case "pause is disabled":
			action = "disabled"
		case "unpaused":
			action = "unpaused"
		}
		return MatchPause{BaseEvent: base, Action: action}, nil
	},
}

var matchStatusScoreDefinition = definition{
	Type:        "MatchStatusScore",
	Category:    "game",
	Description: "A match status dump reported score, map, and rounds played.",
	Regex:       `MatchStatus: Score: (\d+):(\d+) on map "([^"]+)" RoundsPlayed: (-?\d+)`,
	Fields: []field{
		{Name: "scoreCT", Type: "int"},
		{Name: "scoreT", Type: "int"},
		{Name: "map", Type: "string"},
		{Name: "roundsPlayed", Type: "int"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		scoreCT, err := matches.Int("scoreCT")
		if err != nil {
			return nil, err
		}
		scoreT, err := matches.Int("scoreT")
		if err != nil {
			return nil, err
		}
		roundsPlayed, err := matches.Int("roundsPlayed")
		if err != nil {
			return nil, err
		}
		return MatchStatusScore{
			BaseEvent:    base,
			ScoreCT:      scoreCT,
			ScoreT:       scoreT,
			Map:          matches.String("map"),
			RoundsPlayed: roundsPlayed,
		}, nil
	},
}

var gameOverDefinition = definition{
	Type:        "GameOver",
	Category:    "game",
	Description: "A game ended.",
	Regex:       `Game Over: (\w+) (\w+) (\w+) score (\d+):(\d+) after (\d+) min`,
	Fields: []field{
		{Name: "mode", Type: "word"},
		{Name: "mapGroup", Type: "word"},
		{Name: "map", Type: "word"},
		{Name: "scoreCT", Type: "int"},
		{Name: "scoreT", Type: "int"},
		{Name: "duration", Type: "int"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		scoreCT, err := matches.Int("scoreCT")
		if err != nil {
			return nil, err
		}
		scoreT, err := matches.Int("scoreT")
		if err != nil {
			return nil, err
		}
		duration, err := matches.Int("duration")
		if err != nil {
			return nil, err
		}
		return GameOver{
			BaseEvent: base,
			Mode:      matches.String("mode"),
			MapGroup:  matches.String("mapGroup"),
			Map:       matches.String("map"),
			ScoreCT:   scoreCT,
			ScoreT:    scoreT,
			Duration:  duration,
		}, nil
	},
}
