package cs2log

// Accolade is emitted for end-of-match accolade score lines.
type Accolade struct {
	BaseEvent
	Phase      string  `json:"phase"`
	Metric     string  `json:"metric"`
	PlayerName string  `json:"playerName"`
	UserID     string  `json:"userId"`
	Value      float64 `json:"value"`
	Position   int     `json:"position"`
	Score      float64 `json:"score"`
}

var accoladeDefinition = definition{
	Type:        "Accolade",
	Category:    "game",
	Description: "An end-of-match accolade score was reported.",
	Regex:       `ACCOLADE, ([A-Z]+): \{([^}]+)\},\s*([^<]+)<([^>]+)>,\s*VALUE: (-?\d+(?:\.\d+)?),\s*POS: (-?\d+),\s*SCORE: (-?\d+(?:\.\d+)?)`,
	Fields: []field{
		{Name: "phase", Type: "word"},
		{Name: "metric", Type: "string"},
		{Name: "playerName", Type: "string"},
		{Name: "userID", Type: "string"},
		{Name: "value", Type: "float"},
		{Name: "position", Type: "int"},
		{Name: "score", Type: "float"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		value, err := matches.Float64("value")
		if err != nil {
			return nil, err
		}
		position, err := matches.Int("position")
		if err != nil {
			return nil, err
		}
		score, err := matches.Float64("score")
		if err != nil {
			return nil, err
		}
		return Accolade{
			BaseEvent:  base,
			Phase:      matches.String("phase"),
			Metric:     matches.String("metric"),
			PlayerName: matches.String("playerName"),
			UserID:     matches.String("userID"),
			Value:      value,
			Position:   position,
			Score:      score,
		}, nil
	},
}
