package cs2log

// PlayerSay is emitted when a player sends a chat message.
type PlayerSay struct {
	BaseEvent
	Player  Player `json:"player"`
	Text    string `json:"text"`
	Message string `json:"message"`
	Team    bool   `json:"team"`
}

var playerSayDefinition = definition{
	Type:        "PlayerSay",
	Category:    "chat",
	Description: "A player sent a chat message.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" say(_team)? "(.*)"`,
	Fields: []field{
		{Name: "player", Type: "player", Description: "Player who sent the message."},
		{Name: "team", Type: "string", Description: "Whether the message was sent to team chat."},
		{Name: "text", Type: "string", Description: "Chat message text."},
	},
	Examples: []example{
		{
			Line:     `L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`,
			Timezone: "America/New_York",
			UTC:      "2026-07-05T04:00:00Z",
		},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		return PlayerSay{
			BaseEvent: base,
			Player:    player,
			Text:      matches.String("text"),
			Message:   matches.String("text"),
			Team:      matches.BoolPresence("team"),
		}, nil
	},
}
