package cs2log

import "testing"

func TestSideFromTeamID(t *testing.T) {
	tests := []struct {
		id   TeamID
		side Side
		ok   bool
	}{
		{id: TeamIDUnassigned, side: SideUnassigned, ok: true},
		{id: TeamIDSpectator, side: SideSpectator, ok: true},
		{id: TeamIDTerrorist, side: SideTerrorist, ok: true},
		{id: TeamIDCT, side: SideCT, ok: true},
		{id: TeamID(99), ok: false},
	}

	for _, test := range tests {
		side, ok := SideFromTeamID(test.id)
		if side != test.side || ok != test.ok {
			t.Fatalf("SideFromTeamID(%d) = %q, %t; want %q, %t", test.id, side, ok, test.side, test.ok)
		}
	}
}
