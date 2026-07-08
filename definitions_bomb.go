package cs2log

// PlayerBombGot is emitted when a player gets the bomb.
type PlayerBombGot struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerBombPlanted is emitted when a player plants the bomb.
type PlayerBombPlanted struct {
	BaseEvent
	Player Player `json:"player"`
	Site   string `json:"site,omitempty"`
}

// PlayerBombBeginPlant is emitted when a player begins planting the bomb.
type PlayerBombBeginPlant struct {
	BaseEvent
	Player Player `json:"player"`
	Site   string `json:"site,omitempty"`
}

// PlayerBombDropped is emitted when a player drops the bomb.
type PlayerBombDropped struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerBombBeginDefuse is emitted when a player begins defusing the bomb.
type PlayerBombBeginDefuse struct {
	BaseEvent
	Player Player `json:"player"`
	Kit    bool   `json:"kit"`
}

// PlayerBombDefused is emitted when a player defuses the bomb.
type PlayerBombDefused struct {
	BaseEvent
	Player Player `json:"player"`
}

var playerBombGotDefinition = bombTriggerDefinition("PlayerBombGot", "Got_The_Bomb", func(base BaseEvent, player Player) Event {
	return PlayerBombGot{BaseEvent: base, Player: player}
})

var playerBombBeginPlantDefinition = bombsiteTriggerDefinition("PlayerBombBeginPlant", "Bomb_Begin_Plant", func(base BaseEvent, player Player, site string) Event {
	return PlayerBombBeginPlant{BaseEvent: base, Player: player, Site: site}
})

var playerBombPlantedDefinition = bombsiteTriggerDefinition("PlayerBombPlanted", "Planted_The_Bomb", func(base BaseEvent, player Player, site string) Event {
	return PlayerBombPlanted{BaseEvent: base, Player: player, Site: site}
})

var playerBombDroppedDefinition = bombTriggerDefinition("PlayerBombDropped", "Dropped_The_Bomb", func(base BaseEvent, player Player) Event {
	return PlayerBombDropped{BaseEvent: base, Player: player}
})

var playerBombDefusedDefinition = bombTriggerDefinition("PlayerBombDefused", "Defused_The_Bomb", func(base BaseEvent, player Player) Event {
	return PlayerBombDefused{BaseEvent: base, Player: player}
})

var playerBombBeginDefuseDefinition = definition{
	Type:        "PlayerBombBeginDefuse",
	Category:    "bomb",
	Description: "A player began defusing the bomb.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" triggered "Begin_Bomb_Defuse_With(out)?_Kit"`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "without", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerBombBeginDefuse{BaseEvent: base, Player: player, Kit: !matches.BoolPresence("without")}, nil
	},
}

func bombTriggerDefinition(eventType, trigger string, build func(BaseEvent, Player) Event) definition {
	return definition{
		Type:        eventType,
		Category:    "bomb",
		Description: "A bomb objective trigger was emitted.",
		Regex:       `"(` + playerTokenActiveSideRegex + `)" triggered "` + trigger + `"`,
		Fields:      []field{{Name: "player", Type: "player"}},
		Build: func(base BaseEvent, matches Matches) (Event, error) {
			player, err := matches.Player("player")
			if err != nil {
				return nil, err
			}
			return build(base, player), nil
		},
	}
}

func bombsiteTriggerDefinition(eventType, trigger string, build func(BaseEvent, Player, string) Event) definition {
	return definition{
		Type:        eventType,
		Category:    "bomb",
		Description: "A bomb objective trigger was emitted.",
		Regex:       `"(` + playerTokenActiveSideRegex + `)" triggered "` + trigger + `"(?: at bombsite ([A-Z]))?`,
		Fields:      []field{{Name: "player", Type: "player"}, {Name: "site", Type: "string"}},
		Build: func(base BaseEvent, matches Matches) (Event, error) {
			player, err := matches.Player("player")
			if err != nil {
				return nil, err
			}
			return build(base, player, matches.String("site")), nil
		},
	}
}
