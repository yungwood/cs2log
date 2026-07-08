package cs2log

import (
	"fmt"
	"regexp"
)

// Player is a player identity as represented in CS2 server logs.
type Player struct {
	Name    string `json:"name"`
	UserID  string `json:"userId"`
	SteamID string `json:"steamId"`
	Side    Side   `json:"side,omitempty"`
}

const (
	playerIdentityRegex        = `[^"]+<[^<>]*><.*>`
	playerSideActiveRegex      = `TERRORIST|CT`
	playerSideKnownRegex       = playerSideActiveRegex + `|Unassigned|Spectator|`
	playerTokenAnySideRegex    = playerIdentityRegex + `<[^<>]*>`
	playerTokenActiveSideRegex = playerIdentityRegex + `<(?:` + playerSideActiveRegex + `)>`
	playerTokenKnownSideRegex  = playerIdentityRegex + `<(?:` + playerSideKnownRegex + `)>`
	playerTokenEmptySideRegex  = playerIdentityRegex + `<>`
	playerTokenNoSideRegex     = playerIdentityRegex
)

var playerPattern = regexp.MustCompile(`^(.*)<([^<>]*)><(.*)><([^<>]*)>$`)

func parsePlayer(value string) (Player, error) {
	result := playerPattern.FindStringSubmatch(value)
	if result == nil {
		return Player{}, fmt.Errorf("invalid player token %q", value)
	}
	return Player{
		Name:    result[1],
		UserID:  result[2],
		SteamID: result[3],
		Side:    Side(result[4]),
	}, nil
}
