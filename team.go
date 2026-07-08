package cs2log

// Side is a CS2 team/side label as emitted in regular log player tokens.
type Side string

const (
	// SideUnassigned means the player is not assigned to a side.
	SideUnassigned Side = "Unassigned"
	// SideSpectator means the player is spectating.
	SideSpectator Side = "Spectator"
	// SideTerrorist is the terrorist side label used in CS2 logs.
	SideTerrorist Side = "TERRORIST"
	// SideCT is the counter-terrorist side label used in CS2 logs.
	SideCT Side = "CT"
)

// TeamID is the numeric Source team id used by some structured CS2 log blocks.
type TeamID int

const (
	// TeamIDUnassigned is the Source team id for unassigned players.
	TeamIDUnassigned TeamID = 0
	// TeamIDSpectator is the Source team id for spectators.
	TeamIDSpectator TeamID = 1
	// TeamIDTerrorist is the Source team id for terrorists.
	TeamIDTerrorist TeamID = 2
	// TeamIDCT is the Source team id for counter-terrorists.
	TeamIDCT TeamID = 3
)

// SideFromTeamID maps a numeric team id to the matching CS2 side label.
func SideFromTeamID(id TeamID) (Side, bool) {
	switch id {
	case TeamIDUnassigned:
		return SideUnassigned, true
	case TeamIDSpectator:
		return SideSpectator, true
	case TeamIDTerrorist:
		return SideTerrorist, true
	case TeamIDCT:
		return SideCT, true
	default:
		return "", false
	}
}
