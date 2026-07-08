package cs2log

import (
	"fmt"
	"strconv"
	"strings"
)

// PlayerThrew is emitted when a player throws a grenade.
type PlayerThrew struct {
	BaseEvent
	Player   Player   `json:"player"`
	Position Position `json:"position"`
	Entindex int      `json:"entindex,omitempty"`
	Grenade  string   `json:"grenade"`
}

// PlayerSvThrow is emitted for sv_throw_* grenade trajectory command logs.
type PlayerSvThrow struct {
	BaseEvent
	Player  Player    `json:"player"`
	Grenade string    `json:"grenade"`
	Values  []float64 `json:"values"`
}

// PlayerBlinded is emitted when a player is blinded by a flashbang.
type PlayerBlinded struct {
	BaseEvent
	Attacker Player  `json:"attacker"`
	Victim   Player  `json:"victim"`
	Duration float64 `json:"duration"`
	Entindex int     `json:"entindex"`
}

// ProjectileSpawned is emitted when a projectile spawn is reported.
type ProjectileSpawned struct {
	BaseEvent
	Position PositionFloat `json:"position"`
	Velocity Velocity      `json:"velocity"`
}

var playerThrewDefinition = definition{
	Type:        "PlayerThrew",
	Category:    "projectile",
	Description: "A player threw a grenade.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" threw (\w+) \[(-?\d+) (-?\d+) (-?\d+)\](?: flashbang entindex (\d+))?\)?`,
	Fields: append(append([]field{
		{Name: "player", Type: "player"},
		{Name: "grenade", Type: "word"},
	}, positionFields("position")...), field{Name: "entindex", Type: "int"}),
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		position, err := positionFromMatches(matches, "position")
		if err != nil {
			return nil, err
		}
		entindex, err := matches.Int("entindex")
		if err != nil {
			return nil, err
		}
		return PlayerThrew{
			BaseEvent: base,
			Player:    player,
			Position:  position,
			Entindex:  entindex,
			Grenade:   matches.String("grenade"),
		}, nil
	},
}

var playerSvThrowDefinition = definition{
	Type:        "PlayerSvThrow",
	Category:    "projectile",
	Description: "A player emitted an sv_throw_* grenade trajectory command.",
	Regex:       `"(` + playerTokenActiveSideRegex + `)" sv_throw_(\w+) (.+)`,
	Fields: []field{
		{Name: "player", Type: "player"},
		{Name: "grenade", Type: "word"},
		{Name: "values", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		player, err := matches.Player("player")
		if err != nil {
			return nil, err
		}
		values, err := parseSvThrowValues(matches.String("values"))
		if err != nil {
			return nil, err
		}
		return PlayerSvThrow{
			BaseEvent: base,
			Player:    player,
			Grenade:   matches.String("grenade"),
			Values:    values,
		}, nil
	},
}

func parseSvThrowValues(raw string) ([]float64, error) {
	fields := strings.Fields(raw)
	values := make([]float64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, fmt.Errorf("parse sv_throw value %q: %w", field, err)
		}
		values = append(values, value)
	}
	return values, nil
}

var playerBlindedDefinition = definition{
	Type:        "PlayerBlinded",
	Category:    "projectile",
	Description: "A player was blinded by a flashbang.",
	Regex:       `"(` + playerTokenAnySideRegex + `)" blinded for ([\d.]+) by "(` + playerTokenAnySideRegex + `)" from flashbang entindex (\d+)\s*`,
	Fields: []field{
		{Name: "victim", Type: "player"},
		{Name: "for", Type: "float"},
		{Name: "attacker", Type: "player"},
		{Name: "entindex", Type: "int"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		victim, err := matches.Player("victim")
		if err != nil {
			return nil, err
		}
		duration, err := matches.Float64("for")
		if err != nil {
			return nil, err
		}
		attacker, err := matches.Player("attacker")
		if err != nil {
			return nil, err
		}
		entindex, err := matches.Int("entindex")
		if err != nil {
			return nil, err
		}
		return PlayerBlinded{
			BaseEvent: base,
			Attacker:  attacker,
			Victim:    victim,
			Duration:  duration,
			Entindex:  entindex,
		}, nil
	},
}

var projectileSpawnedDefinition = definition{
	Type:        "ProjectileSpawned",
	Category:    "projectile",
	Description: "A projectile spawned.",
	Regex:       `Molotov projectile spawned at (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+), velocity (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+)`,
	Fields: append(positionFloatFields("position"), []field{
		{Name: "velocityX", Type: "float"},
		{Name: "velocityY", Type: "float"},
		{Name: "velocityZ", Type: "float"},
	}...),
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		position, err := positionFloatFromMatches(matches, "position")
		if err != nil {
			return nil, err
		}
		velocity, err := velocityFromMatches(matches, "velocity")
		if err != nil {
			return nil, err
		}
		return ProjectileSpawned{BaseEvent: base, Position: position, Velocity: velocity}, nil
	},
}
