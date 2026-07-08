package cs2log

import "strings"

// PlayerKill is emitted when one player kills another.
type PlayerKill struct {
	BaseEvent
	Attacker         Player   `json:"attacker"`
	AttackerPosition Position `json:"attackerPosition"`
	Victim           Player   `json:"victim"`
	VictimPosition   Position `json:"victimPosition"`
	Weapon           string   `json:"weapon"`
	Headshot         bool     `json:"headshot"`
	Penetrated       bool     `json:"penetrated"`
	Flags            []string `json:"flags,omitempty"`
}

// PlayerKillOther is emitted when a player destroys or kills a non-player entity.
type PlayerKillOther struct {
	BaseEvent
	Attacker         Player   `json:"attacker"`
	AttackerPosition Position `json:"attackerPosition"`
	Target           Entity   `json:"target"`
	TargetPosition   Position `json:"targetPosition"`
	Weapon           string   `json:"weapon"`
	Flags            []string `json:"flags,omitempty"`
}

// PlayerKillAssist is emitted when a player assists a kill.
type PlayerKillAssist struct {
	BaseEvent
	Attacker Player `json:"attacker"`
	Victim   Player `json:"victim"`
}

// PlayerFlashAssist is emitted when a player flash-assists a kill.
type PlayerFlashAssist struct {
	BaseEvent
	Attacker Player `json:"attacker"`
	Victim   Player `json:"victim"`
}

// PlayerAttack is emitted when one player damages another.
type PlayerAttack struct {
	BaseEvent
	Attacker         Player   `json:"attacker"`
	AttackerPosition Position `json:"attackerPosition"`
	Victim           Player   `json:"victim"`
	VictimPosition   Position `json:"victimPosition"`
	Weapon           string   `json:"weapon"`
	Damage           int      `json:"damage"`
	DamageArmor      int      `json:"damageArmor"`
	Health           int      `json:"health"`
	Armor            int      `json:"armor"`
	Hitgroup         string   `json:"hitgroup"`
}

// PlayerKilledBomb is emitted when a player is killed by the bomb.
type PlayerKilledBomb struct {
	BaseEvent
	Player   Player   `json:"player"`
	Position Position `json:"position"`
}

// PlayerKilledSuicide is emitted when a player commits suicide.
type PlayerKilledSuicide struct {
	BaseEvent
	Player   Player   `json:"player"`
	Position Position `json:"position"`
	Method   string   `json:"method"`
}

var playerKillDefinition = definition{
	Type:        "PlayerKill",
	Category:    "combat",
	Description: "A player killed another player.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] killed "(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)"(?: \(([^)]*)\))?`,
	Fields: []field{
		{Name: "attacker", Type: "player"},
		{Name: "attackerX", Type: "int"},
		{Name: "attackerY", Type: "int"},
		{Name: "attackerZ", Type: "int"},
		{Name: "victim", Type: "player"},
		{Name: "victimX", Type: "int"},
		{Name: "victimY", Type: "int"},
		{Name: "victimZ", Type: "int"},
		{Name: "weapon", Type: "word"},
		{Name: "flags", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		attackerPosition, err := positionFromMatches(matches, "attacker")
		if err != nil {
			return nil, err
		}
		victim, err := matches.Player("victim")
		if err != nil {
			return nil, err
		}
		victimPosition, err := positionFromMatches(matches, "victim")
		if err != nil {
			return nil, err
		}
		flags := matches.String("flags")
		flagFields := strings.Fields(flags)
		return PlayerKill{
			BaseEvent:        base,
			Attacker:         attacker,
			AttackerPosition: attackerPosition,
			Victim:           victim,
			VictimPosition:   victimPosition,
			Weapon:           matches.String("weapon"),
			Headshot:         strings.Contains(flags, "headshot"),
			Penetrated:       strings.Contains(flags, "penetrated"),
			Flags:            flagFields,
		}, nil
	},
}

var playerKillOtherDefinition = definition{
	Type:        "PlayerKillOther",
	Category:    "combat",
	Description: "A player killed or destroyed a non-player entity.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] killed other "([^"<>]+)<([^<>]+)>" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)"(?: \(([^)]*)\))?`,
	Fields: []field{
		{Name: "attacker", Type: "player"},
		{Name: "attackerX", Type: "int"},
		{Name: "attackerY", Type: "int"},
		{Name: "attackerZ", Type: "int"},
		{Name: "targetName", Type: "string"},
		{Name: "targetID", Type: "string"},
		{Name: "targetX", Type: "int"},
		{Name: "targetY", Type: "int"},
		{Name: "targetZ", Type: "int"},
		{Name: "weapon", Type: "word"},
		{Name: "flags", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		attackerPosition, err := positionFromMatches(matches, "attacker")
		if err != nil {
			return nil, err
		}
		targetPosition, err := positionFromMatches(matches, "target")
		if err != nil {
			return nil, err
		}
		return PlayerKillOther{
			BaseEvent:        base,
			Attacker:         attacker,
			AttackerPosition: attackerPosition,
			Target: Entity{
				Name: matches.String("targetName"),
				ID:   matches.String("targetID"),
			},
			TargetPosition: targetPosition,
			Weapon:         matches.String("weapon"),
			Flags:          strings.Fields(matches.String("flags")),
		}, nil
	},
}

var playerKillAssistDefinition = definition{
	Type:        "PlayerKillAssist",
	Category:    "combat",
	Description: "A player assisted a kill.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" assisted killing "(` + playerTokenActiveSideRegex + `)"`,
	Fields:      []field{{Name: "attacker", Type: "player"}, {Name: "victim", Type: "player"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		victim, err := matches.Player("victim")
		if err != nil {
			return nil, err
		}
		return PlayerKillAssist{BaseEvent: base, Attacker: attacker, Victim: victim}, nil
	},
}

var playerFlashAssistDefinition = definition{
	Type:        "PlayerFlashAssist",
	Category:    "combat",
	Description: "A player flash-assisted a kill.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" flash-assisted killing "(` + playerTokenActiveSideRegex + `)"`,
	Fields:      []field{{Name: "attacker", Type: "player"}, {Name: "victim", Type: "player"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		victim, err := matches.Player("victim")
		if err != nil {
			return nil, err
		}
		return PlayerFlashAssist{BaseEvent: base, Attacker: attacker, Victim: victim}, nil
	},
}

var playerAttackDefinition = definition{
	Type:        "PlayerAttack",
	Category:    "combat",
	Description: "A player attacked another player.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] attacked "(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w*)" \(damage "(\d+)"\) \(damage_armor "(\d+)"\) \(health "(\d+)"\) \(armor "(\d+)"\) \(hitgroup "([\w ]+)"\)`,
	Fields: []field{
		{Name: "attacker", Type: "player"},
		{Name: "attackerX", Type: "int"},
		{Name: "attackerY", Type: "int"},
		{Name: "attackerZ", Type: "int"},
		{Name: "victim", Type: "player"},
		{Name: "victimX", Type: "int"},
		{Name: "victimY", Type: "int"},
		{Name: "victimZ", Type: "int"},
		{Name: "weapon", Type: "word"},
		{Name: "damage", Type: "int"},
		{Name: "damageArmor", Type: "int"},
		{Name: "health", Type: "int"},
		{Name: "armor", Type: "int"},
		{Name: "hitgroup", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		attackerPosition, err := positionFromMatches(matches, "attacker")
		if err != nil {
			return nil, err
		}
		victim, err := matches.Player("victim")
		if err != nil {
			return nil, err
		}
		victimPosition, err := positionFromMatches(matches, "victim")
		if err != nil {
			return nil, err
		}
		damage, err := matches.Int("damage")
		if err != nil {
			return nil, err
		}
		damageArmor, err := matches.Int("damageArmor")
		if err != nil {
			return nil, err
		}
		health, err := matches.Int("health")
		if err != nil {
			return nil, err
		}
		armor, err := matches.Int("armor")
		if err != nil {
			return nil, err
		}
		return PlayerAttack{
			BaseEvent:        base,
			Attacker:         attacker,
			AttackerPosition: attackerPosition,
			Victim:           victim,
			VictimPosition:   victimPosition,
			Weapon:           matches.String("weapon"),
			Damage:           damage,
			DamageArmor:      damageArmor,
			Health:           health,
			Armor:            armor,
			Hitgroup:         matches.String("hitgroup"),
		}, nil
	},
}

var playerKilledBombDefinition = definition{
	Type:        "PlayerKilledBomb",
	Category:    "combat",
	Description: "A player was killed by the bomb.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] was killed by the bomb\.`,
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
		return PlayerKilledBomb{BaseEvent: base, Player: player, Position: position}, nil
	},
}

var playerKilledSuicideDefinition = definition{
	Type:        "PlayerKilledSuicide",
	Category:    "combat",
	Description: "A player committed suicide.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" \[(-?\d+) (-?\d+) (-?\d+)\] committed suicide with "(.*)"`,
	Fields: append(append([]field{
		{Name: "player", Type: "player"},
	}, positionFields("position")...), field{Name: "with", Type: "string"}),
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		position, err := positionFromMatches(matches, "position")
		if err != nil {
			return nil, err
		}
		return PlayerKilledSuicide{BaseEvent: base, Player: player, Position: position, Method: matches.String("with")}, nil
	},
}
