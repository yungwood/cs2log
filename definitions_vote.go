package cs2log

// VoteActor is the player identity attached to a vote log line.
type VoteActor struct {
	Player Player `json:"player"`
	Slot   int    `json:"slot"`
	Area   int    `json:"area"`
}

// VoteStarted is emitted when a player starts a vote.
type VoteStarted struct {
	BaseEvent
	Issue string    `json:"issue"`
	Actor VoteActor `json:"actor"`
}

// VoteCast is emitted when a player casts a vote option.
type VoteCast struct {
	BaseEvent
	Issue  string    `json:"issue"`
	Actor  VoteActor `json:"actor"`
	Option int       `json:"option"`
}

// VoteSucceeded is emitted when a vote succeeds.
type VoteSucceeded struct {
	BaseEvent
	Issue string    `json:"issue"`
	Actor VoteActor `json:"actor"`
}

// VoteFailed is emitted when a vote fails.
type VoteFailed struct {
	BaseEvent
	Issue string `json:"issue"`
}

var voteStartedDefinition = voteActorDefinition("VoteStarted", "started", func(base BaseEvent, issue string, actor VoteActor) Event {
	return VoteStarted{BaseEvent: base, Issue: issue, Actor: actor}
})

var voteSucceededDefinition = voteActorDefinition("VoteSucceeded", "succeeded", func(base BaseEvent, issue string, actor VoteActor) Event {
	return VoteSucceeded{BaseEvent: base, Issue: issue, Actor: actor}
})

var voteCastDefinition = definition{
	Type:        "VoteCast",
	Category:    "vote",
	Description: "A player cast a vote option.",
	Regex:       `Vote cast "([^"]+)" from #(\d+) "(` + playerTokenAnySideRegex + `)<Area (-?\d+)>" option(\d+)`,
	Fields: []field{
		{Name: "issue", Type: "string"},
		{Name: "slot", Type: "int"},
		{Name: "player", Type: "player"},
		{Name: "area", Type: "int"},
		{Name: "option", Type: "int"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		actor, err := voteActorFromMatches(matches)
		if err != nil {
			return nil, err
		}
		option, err := matches.Int("option")
		if err != nil {
			return nil, err
		}
		return VoteCast{
			BaseEvent: base,
			Issue:     matches.String("issue"),
			Actor:     actor,
			Option:    option,
		}, nil
	},
}

var voteFailedDefinition = definition{
	Type:        "VoteFailed",
	Category:    "vote",
	Description: "A vote failed.",
	Regex:       `Vote failed "([^"]+)"\s*`,
	Fields:      []field{{Name: "issue", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return VoteFailed{BaseEvent: base, Issue: matches.String("issue")}, nil
	},
}

func voteActorDefinition(eventType, action string, build func(BaseEvent, string, VoteActor) Event) definition {
	return definition{
		Type:        eventType,
		Category:    "vote",
		Description: "A player vote event was emitted.",
		Regex:       `Vote ` + action + ` "([^"]+)" from #(\d+) "(` + playerTokenAnySideRegex + `)<Area (-?\d+)>"`,
		Fields: []field{
			{Name: "issue", Type: "string"},
			{Name: "slot", Type: "int"},
			{Name: "player", Type: "player"},
			{Name: "area", Type: "int"},
		},
		Build: func(base BaseEvent, matches Matches) (Event, error) {
			actor, err := voteActorFromMatches(matches)
			if err != nil {
				return nil, err
			}
			return build(base, matches.String("issue"), actor), nil
		},
	}
}

func voteActorFromMatches(matches Matches) (VoteActor, error) {
	player, err := matches.Player("player")
	if err != nil {
		return VoteActor{}, err
	}
	slot, err := matches.Int("slot")
	if err != nil {
		return VoteActor{}, err
	}
	area, err := matches.Int("area")
	if err != nil {
		return VoteActor{}, err
	}
	return VoteActor{Player: player, Slot: slot, Area: area}, nil
}
