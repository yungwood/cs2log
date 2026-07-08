package cs2log

import "strings"

// PlayerConnected is emitted when a player connects.
type PlayerConnected struct {
	BaseEvent
	Player  Player `json:"player"`
	Address string `json:"address"`
}

// PlayerDisconnected is emitted when a player disconnects.
type PlayerDisconnected struct {
	BaseEvent
	Player Player `json:"player"`
	Reason string `json:"reason"`
}

// PlayerEntered is emitted when a player enters the game.
type PlayerEntered struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerValidated is emitted when the server validates a player's Steam user ID.
type PlayerValidated struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerBanned is emitted when a player is banned.
type PlayerBanned struct {
	BaseEvent
	Player   Player `json:"player"`
	Duration string `json:"duration"`
	By       string `json:"by"`
}

// PlayerSwitched is emitted when a player switches sides.
type PlayerSwitched struct {
	BaseEvent
	Player Player `json:"player"`
	From   string `json:"from"`
	To     string `json:"to"`
}

// PlayerPurchase is emitted when a player purchases an item.
type PlayerPurchase struct {
	BaseEvent
	Player Player `json:"player"`
	Item   string `json:"item"`
}

// PlayerPickedUp is emitted when a player picks up an item.
type PlayerPickedUp struct {
	BaseEvent
	Player Player `json:"player"`
	Item   string `json:"item"`
}

// PlayerDropped is emitted when a player drops an item.
type PlayerDropped struct {
	BaseEvent
	Player Player `json:"player"`
	Item   string `json:"item"`
}

// PlayerLeftBuyZone is emitted when a player leaves the buy zone with items.
type PlayerLeftBuyZone struct {
	BaseEvent
	Player Player   `json:"player"`
	Items  []string `json:"items"`
}

// PlayerTouchedHostage is emitted when a player touches a hostage.
type PlayerTouchedHostage struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerRescuedHostage is emitted when a player rescues a hostage.
type PlayerRescuedHostage struct {
	BaseEvent
	Player Player `json:"player"`
}

// PlayerPickedUpHostage is emitted when a player starts carrying a hostage.
type PlayerPickedUpHostage struct {
	BaseEvent
	Player   Player   `json:"player"`
	Position Position `json:"position"`
}

// PlayerDroppedOffHostage is emitted when a player drops off a hostage.
type PlayerDroppedOffHostage struct {
	BaseEvent
	Player   Player   `json:"player"`
	Position Position `json:"position"`
}

// PlayerMoneyChange is emitted when a player's money changes.
type PlayerMoneyChange struct {
	BaseEvent
	Player        Player   `json:"player"`
	Equation      Equation `json:"equation"`
	Purchase      string   `json:"purchase,omitempty"`
	AcquireReason string   `json:"acquireReason,omitempty"`
}

var playerConnectedDefinition = definition{
	Type:        "PlayerConnected",
	Category:    "player",
	Description: "A player connected.",
	Regex:       `"(` + playerTokenEmptySideRegex + `)" connected, address "(.*)"`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "address", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerConnected{BaseEvent: base, Player: player, Address: matches.String("address")}, nil
	},
}

var playerDisconnectedDefinition = definition{
	Type:        "PlayerDisconnected",
	Category:    "player",
	Description: "A player disconnected.",
	Regex:       `"(` + playerTokenKnownSideRegex + `)" disconnected \(reason "(.+)"\)`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "reason", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerDisconnected{BaseEvent: base, Player: player, Reason: matches.String("reason")}, nil
	},
}

var playerEnteredDefinition = definition{
	Type:        "PlayerEntered",
	Category:    "player",
	Description: "A player entered the game.",
	Regex:       `"(` + playerTokenEmptySideRegex + `)" entered the game`,
	Fields:      []field{{Name: "player", Type: "player"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerEntered{BaseEvent: base, Player: player}, nil
	},
}

var playerValidatedDefinition = definition{
	Type:        "PlayerValidated",
	Category:    "player",
	Description: "A player's Steam user ID was validated.",
	Regex:       `"(` + playerTokenEmptySideRegex + `)" STEAM USERID validated`,
	Fields:      []field{{Name: "player", Type: "player"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerValidated{BaseEvent: base, Player: player}, nil
	},
}

var playerBannedDefinition = definition{
	Type:        "PlayerBanned",
	Category:    "player",
	Description: "A player was banned.",
	Regex:       `Banid: "(` + playerTokenAnySideRegex + `)" was banned "([\w. ]+)" by "(\w+)"`,
	Fields: []field{
		{Name: "player", Type: "player"},
		{Name: "duration", Type: "string"},
		{Name: "by", Type: "word"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerBanned{BaseEvent: base, Player: player, Duration: matches.String("duration"), By: matches.String("by")}, nil
	},
}

var playerSwitchedDefinition = definition{
	Type:        "PlayerSwitched",
	Category:    "player",
	Description: "A player switched sides.",
	Regex:       `"(` + playerTokenNoSideRegex + `)" switched from team <(Unassigned|Spectator|TERRORIST|CT)> to <(Unassigned|Spectator|TERRORIST|CT)>`,
	Fields: []field{
		{Name: "player", Type: "player"},
		{Name: "from", Type: "side"},
		{Name: "to", Type: "side"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := parsePlayer(matches.String("player") + "<>")
		if err != nil {
			return nil, err
		}
		return PlayerSwitched{BaseEvent: base, Player: player, From: matches.String("from"), To: matches.String("to")}, nil
	},
}

var playerPurchaseDefinition = definition{
	Type:        "PlayerPurchase",
	Category:    "player",
	Description: "A player purchased an item.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" purchased "(\w+)"`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "item", Type: "word"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerPurchase{BaseEvent: base, Player: player, Item: matches.String("item")}, nil
	},
}

var playerPickedUpDefinition = definition{
	Type:        "PlayerPickedUp",
	Category:    "player",
	Description: "A player picked up an item.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" picked up "(\w+)"`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "item", Type: "word"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerPickedUp{BaseEvent: base, Player: player, Item: matches.String("item")}, nil
	},
}

var playerDroppedDefinition = definition{
	Type:        "PlayerDropped",
	Category:    "player",
	Description: "A player dropped an item.",
	Regex:       `"(` + playerTokenKnownSideRegex + `)" dropped "(\w+)"`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "item", Type: "word"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerDropped{BaseEvent: base, Player: player, Item: matches.String("item")}, nil
	},
}

var playerLeftBuyZoneDefinition = definition{
	Type:        "PlayerLeftBuyZone",
	Category:    "player",
	Description: "A player left the buy zone with an item list.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" left buyzone with \[\s*([^\]]*?)\s*\]`,
	Fields:      []field{{Name: "player", Type: "player"}, {Name: "items", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerLeftBuyZone{
			BaseEvent: base,
			Player:    player,
			Items:     strings.Fields(matches.String("items")),
		}, nil
	},
}

var playerTouchedHostageDefinition = playerTriggerDefinition("PlayerTouchedHostage", "Touched_A_Hostage", func(base BaseEvent, player Player) Event {
	return PlayerTouchedHostage{BaseEvent: base, Player: player}
})

var playerRescuedHostageDefinition = playerTriggerDefinition("PlayerRescuedHostage", "Rescued_A_Hostage", func(base BaseEvent, player Player) Event {
	return PlayerRescuedHostage{BaseEvent: base, Player: player}
})

var playerPickedUpHostageDefinition = playerHostagePositionDefinition("PlayerPickedUpHostage", "picked up", func(base BaseEvent, player Player, position Position) Event {
	return PlayerPickedUpHostage{BaseEvent: base, Player: player, Position: position}
})

var playerDroppedOffHostageDefinition = playerHostagePositionDefinition("PlayerDroppedOffHostage", "dropped off", func(base BaseEvent, player Player, position Position) Event {
	return PlayerDroppedOffHostage{BaseEvent: base, Player: player, Position: position}
})

func playerTriggerDefinition(eventType, trigger string, build func(BaseEvent, Player) Event) definition {
	return definition{
		Type:        eventType,
		Category:    "player",
		Description: "A player trigger was emitted.",
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

func playerHostagePositionDefinition(eventType, action string, build func(BaseEvent, Player, Position) Event) definition {
	return definition{
		Type:        eventType,
		Category:    "player",
		Description: "A player hostage action was emitted.",
		Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] ` + action + ` a hostage`,
		Fields: append([]field{
			{Name: "player", Type: "player"},
		}, positionFields("position")...),
		Build: func(base BaseEvent, matches Matches) (Event, error) {
			player, err := matches.Player("player")
			if err != nil {
				return nil, err
			}
			position, err := positionFromMatches(matches, "position")
			if err != nil {
				return nil, err
			}
			return build(base, player, position), nil
		},
	}
}

var playerMoneyChangeDefinition = definition{
	Type:        "PlayerMoneyChange",
	Category:    "player",
	Description: "A player's money changed.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" money change (\d+)\+?(-?\d+) = \$(\d+) \(tracked\)(?: \((purchase|acquire): (\w+)\))?`,
	Fields: []field{
		{Name: "player", Type: "player"},
		{Name: "a", Type: "int"},
		{Name: "b", Type: "int"},
		{Name: "result", Type: "int"},
		{Name: "reasonType", Type: "word"},
		{Name: "reason", Type: "word"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		a, err := matches.Int("a")
		if err != nil {
			return nil, err
		}
		b, err := matches.Int("b")
		if err != nil {
			return nil, err
		}
		result, err := matches.Int("result")
		if err != nil {
			return nil, err
		}
		var purchase string
		var acquireReason string
		switch matches.String("reasonType") {
		case "purchase":
			purchase = matches.String("reason")
		case "acquire":
			acquireReason = matches.String("reason")
		}
		return PlayerMoneyChange{
			BaseEvent:     base,
			Player:        player,
			Equation:      Equation{A: a, B: b, Result: result},
			Purchase:      purchase,
			AcquireReason: acquireReason,
		}, nil
	},
}
