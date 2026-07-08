package cs2log

// TeamScored is emitted when a team score is reported.
type TeamScored struct {
	BaseEvent
	Side       Side `json:"side"`
	Score      int  `json:"score"`
	NumPlayers int  `json:"numPlayers"`
}

// TeamNotice is emitted when a team notice such as a round win is reported.
type TeamNotice struct {
	BaseEvent
	Side    Side   `json:"side"`
	Notice  string `json:"notice"`
	ScoreCT int    `json:"scoreCT"`
	ScoreT  int    `json:"scoreT"`
}

// TeamPlaying is emitted when a side's configured team name is reported.
type TeamPlaying struct {
	BaseEvent
	Side     Side   `json:"side"`
	TeamName string `json:"teamName"`
}

var teamScoredDefinition = definition{
	Type:        "TeamScored",
	Category:    "team",
	Description: "A team score was reported.",
	Regex:       `Team "(CT|TERRORIST)" scored "(\d+)" with "(\d+)" players`,
	Fields: []field{
		{Name: "side", Type: "side"},
		{Name: "score", Type: "int"},
		{Name: "numPlayers", Type: "int"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		score, err := matches.Int("score")
		if err != nil {
			return nil, err
		}
		numPlayers, err := matches.Int("numPlayers")
		if err != nil {
			return nil, err
		}
		return TeamScored{
			BaseEvent:  base,
			Side:       Side(matches.String("side")),
			Score:      score,
			NumPlayers: numPlayers,
		}, nil
	},
}

var teamPlayingDefinition = definition{
	Type:        "TeamPlaying",
	Category:    "team",
	Description: "A side's configured team name was reported.",
	Regex:       `(?:MatchStatus: )?Team playing "(CT|TERRORIST)": (.+)`,
	Fields: []field{
		{Name: "side", Type: "side"},
		{Name: "teamName", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return TeamPlaying{
			BaseEvent: base,
			Side:      Side(matches.String("side")),
			TeamName:  matches.String("teamName"),
		}, nil
	},
}

var teamNoticeDefinition = definition{
	Type:        "TeamNotice",
	Category:    "team",
	Description: "A team notice was reported.",
	Regex:       `Team "(CT|TERRORIST)" triggered "(\w+)" \(CT "(\d+)"\) \(T "(\d+)"\)`,
	Fields: []field{
		{Name: "side", Type: "side"},
		{Name: "notice", Type: "word"},
		{Name: "scoreCT", Type: "int"},
		{Name: "scoreT", Type: "int"},
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
		return TeamNotice{
			BaseEvent: base,
			Side:      Side(matches.String("side")),
			Notice:    matches.String("notice"),
			ScoreCT:   scoreCT,
			ScoreT:    scoreT,
		}, nil
	},
}
