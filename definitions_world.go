package cs2log

// FreezeTimeStart is emitted before each round freeze period starts.
type FreezeTimeStart struct{ BaseEvent }

// WorldMatchStart is emitted when a match starts on a map.
type WorldMatchStart struct {
	BaseEvent
	Map string `json:"map"`
}

// WorldRoundStart is emitted when a round starts.
type WorldRoundStart struct{ BaseEvent }

// WorldRoundRestart is emitted when the server restarts a round.
type WorldRoundRestart struct {
	BaseEvent
	Timeleft int `json:"timeleft"`
}

// WorldRoundEnd is emitted when a round ends.
type WorldRoundEnd struct{ BaseEvent }

// WorldWarmupStart is emitted when warmup starts.
type WorldWarmupStart struct{ BaseEvent }

// WorldWarmupEnd is emitted when warmup ends.
type WorldWarmupEnd struct{ BaseEvent }

// WorldGameCommencing is emitted when the game is commencing.
type WorldGameCommencing struct{ BaseEvent }

var freezTimeStartDefinition = definition{
	Type:        "FreezeTimeStart",
	Category:    "world",
	Description: "A round freeze period started.",
	Regex:       `Starting Freeze period`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return FreezeTimeStart{BaseEvent: base}, nil
	},
}

var worldMatchStartDefinition = definition{
	Type:        "WorldMatchStart",
	Category:    "world",
	Description: "A match started on a map.",
	Regex:       `World triggered "Match_Start" on "(\w+)"`,
	Fields:      []field{{Name: "map", Type: "word"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldMatchStart{BaseEvent: base, Map: matches.String("map")}, nil
	},
}

var worldRoundStartDefinition = definition{
	Type:        "WorldRoundStart",
	Category:    "world",
	Description: "A round started.",
	Regex:       `World triggered "Round_Start"`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldRoundStart{BaseEvent: base}, nil
	},
}

var worldRoundRestartDefinition = definition{
	Type:        "WorldRoundRestart",
	Category:    "world",
	Description: "A round restart was requested.",
	Regex:       `World triggered "Restart_Round_\((\d+)_second\)`,
	Fields:      []field{{Name: "timeleft", Type: "int"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		timeleft, err := matches.Int("timeleft")
		if err != nil {
			return nil, err
		}
		return WorldRoundRestart{BaseEvent: base, Timeleft: timeleft}, nil
	},
}

var worldRoundEndDefinition = definition{
	Type:        "WorldRoundEnd",
	Category:    "world",
	Description: "A round ended.",
	Regex:       `World triggered "Round_End"`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldRoundEnd{BaseEvent: base}, nil
	},
}

var worldWarmupStartDefinition = definition{
	Type:        "WorldWarmupStart",
	Category:    "world",
	Description: "Warmup started.",
	Regex:       `World triggered "Warmup_Start"`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldWarmupStart{BaseEvent: base}, nil
	},
}

var worldWarmupEndDefinition = definition{
	Type:        "WorldWarmupEnd",
	Category:    "world",
	Description: "Warmup ended.",
	Regex:       `World triggered "Warmup_End"`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldWarmupEnd{BaseEvent: base}, nil
	},
}

var worldGameCommencingDefinition = definition{
	Type:        "WorldGameCommencing",
	Category:    "world",
	Description: "The game is commencing.",
	Regex:       `World triggered "Game_Commencing"`,
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return WorldGameCommencing{BaseEvent: base}, nil
	},
}
