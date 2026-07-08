package cs2log

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseLineUsesConfiguredTimezoneAndReturnsUTC(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "Australia/Adelaide"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	want := time.Date(2026, 7, 4, 14, 30, 0, 0, time.UTC)
	if !event.Timestamp().Equal(want) {
		t.Fatalf("timestamp = %s, want %s", event.Timestamp().Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
	if event.Timestamp().Location() != time.UTC {
		t.Fatalf("timestamp location = %s, want UTC", event.Timestamp().Location())
	}
}

func TestParseLineDefaultsToUTC(t *testing.T) {
	parser, err := NewParser(Config{})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00: "Player<1><STEAM_1:1:1><CT>" say "gg"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	want := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	if !event.Timestamp().Equal(want) {
		t.Fatalf("timestamp = %s, want %s", event.Timestamp().Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestParseLineReturnsTimestampErrorForMalformedTimestamp(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	_, err = parser.ParseLine(`L 13/99/2026 - 25:61:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`)
	if err == nil {
		t.Fatal("parse line succeeded with malformed timestamp")
	}
	if errors.Is(err, ErrNoMatch) {
		t.Fatalf("err = %v, want timestamp parse error", err)
	}
	if !strings.Contains(err.Error(), "parse cs2 log timestamp") {
		t.Fatalf("err = %v, want timestamp parse error", err)
	}
}

func TestNewParserRejectsInvalidTimezone(t *testing.T) {
	if _, err := NewParser(Config{LogTimezone: "Not/AZone"}); err == nil {
		t.Fatal("NewParser succeeded with invalid timezone")
	}
}

func TestParseLineReturnsTypedPlayerSay(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "hello"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	playerSay, ok := event.(PlayerSay)
	if !ok {
		t.Fatalf("event type = %T, want PlayerSay", event)
	}
	if playerSay.Player.Name != "Player" || playerSay.Player.SteamID != "STEAM_1:1:1" || playerSay.Message != "hello" {
		t.Fatalf("player say = %#v", playerSay)
	}
}

func TestParseLineReturnsRepresentativeUpstreamEvents(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "team say",
			line: `L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say_team "hold b"`,
			want: "PlayerSay",
		},
		{
			name: "kill with positions",
			line: `L 07/05/2026 - 00:00:01.000: "Player<1><STEAM_1:1:1><CT>" [10 20 30] killed "Bob<2><STEAM_1:1:2><TERRORIST>" [-10 -20 -30] with "ak47" (headshot)`,
			want: "PlayerKill",
		},
		{
			name: "kill other entity",
			line: `L 07/05/2026 - 00:00:01.500: "ExamplePlayer<3><[U:1:123456789]><CT>" [10 20 30] killed other "func_breakable<456>" [-10 -20 -30] with "m4a1_silencer" (penetrated)`,
			want: "PlayerKillOther",
		},
		{
			name: "attack",
			line: `L 07/05/2026 - 00:00:02.000: "Player<1><STEAM_1:1:1><CT>" [10 20 30] attacked "Bob<2><STEAM_1:1:2><TERRORIST>" [-10 -20 -30] with "ak47" (damage "27") (damage_armor "4") (health "73") (armor "96") (hitgroup "chest")`,
			want: "PlayerAttack",
		},
		{
			name: "bomb planted",
			line: `L 07/05/2026 - 00:00:03.000: "Bob<2><STEAM_1:1:2><TERRORIST>" triggered "Planted_The_Bomb" at bombsite A`,
			want: "PlayerBombPlanted",
		},
		{
			name: "bomb begin plant",
			line: `L 07/05/2026 - 00:00:03.250: "Bob<2><STEAM_1:1:2><TERRORIST>" triggered "Bomb_Begin_Plant" at bombsite B`,
			want: "PlayerBombBeginPlant",
		},
		{
			name: "left buyzone with inventory",
			line: `L 07/05/2026 - 00:00:03.750: "ExamplePlayer<6><[U:1:123456789]><TERRORIST>" left buyzone with [ weapon_knife_t weapon_taser weapon_glock weapon_mac10 weapon_molotov weapon_hegrenade weapon_smokegrenade kevlar(100) helmet ]`,
			want: "PlayerLeftBuyZone",
		},
		{
			name: "steam user id validated",
			line: `L 07/05/2026 - 00:00:03.875: "ExamplePlayer<65283><[U:1:123456789]><>" STEAM USERID validated`,
			want: "PlayerValidated",
		},
		{
			name: "projectile spawned",
			line: `L 07/05/2026 - 00:00:04.000: Molotov projectile spawned at 1.500000 -2.250000 3.750000, velocity 4.000000 5.000000 -6.000000`,
			want: "ProjectileSpawned",
		},
		{
			name: "sv throw trajectory",
			line: `L 07/05/2026 - 00:00:04.500: "ExamplePlayer<3><[U:1:123456789]><CT>" sv_throw_hegrenade -380.930 -580.630 347.767 0.000 0.000 0.000 618.108 -11.430 -173.134 600.000 -478.000 0.000 44 3.000`,
			want: "PlayerSvThrow",
		},
		{
			name: "game over",
			line: `L 07/05/2026 - 00:00:05.000: Game Over: competitive mg_active de_dust2 score 13:11 after 42 min`,
			want: "GameOver",
		},
		{
			name: "match status team unset",
			line: `L 07/05/2026 - 00:00:06.000: MatchStatus: Team "CT" is unset.`,
			want: "Ignored",
		},
		{
			name: "match status score",
			line: `L 07/05/2026 - 00:00:07.000: MatchStatus: Score: 1:2 on map "de_train" RoundsPlayed: 3`,
			want: "MatchStatusScore",
		},
		{
			name: "match pause enabled",
			line: `L 07/05/2026 - 00:00:07.250: Match pause is enabled`,
			want: "MatchPause",
		},
		{
			name: "team playing",
			line: `L 07/05/2026 - 00:00:07.375: MatchStatus: Team playing "CT": Counter-Terrorists`,
			want: "TeamPlaying",
		},
		{
			name: "accolade",
			line: `L 07/05/2026 - 00:00:07.500: ACCOLADE, FINAL: {uniqueweaponkills},	ExamplePlayer<3>,	VALUE: 6.000000,	POS: 1,	SCORE: 70.000000`,
			want: "Accolade",
		},
		{
			name: "world warmup start",
			line: `L 07/05/2026 - 00:00:08.000: World triggered "Warmup_Start"`,
			want: "WorldWarmupStart",
		},
		{
			name: "world warmup end",
			line: `L 07/05/2026 - 00:00:09.000: World triggered "Warmup_End"`,
			want: "WorldWarmupEnd",
		},
		{
			name: "touched hostage",
			line: `L 07/05/2026 - 00:00:10.000: "ExamplePlayer<3><[U:1:123456789]><CT>" triggered "Touched_A_Hostage"`,
			want: "PlayerTouchedHostage",
		},
		{
			name: "rescued hostage",
			line: `L 07/05/2026 - 00:00:11.000: "ExamplePlayer<3><[U:1:123456789]><CT>" triggered "Rescued_A_Hostage"`,
			want: "PlayerRescuedHostage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.ParseLine(tt.line)
			if err != nil {
				t.Fatalf("parse line: %v", err)
			}
			if event.EventType() != tt.want {
				t.Fatalf("event type = %s (%T), want %s", event.EventType(), event, tt.want)
			}
		})
	}
}

func TestParseLineReturnsMatchPause(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		line       string
		wantAction string
	}{
		{
			line:       `L 07/05/2026 - 00:00:00.000: Match pause is enabled`,
			wantAction: "enabled",
		},
		{
			line:       `L 07/05/2026 - 00:00:01.000: Match pause is disabled`,
			wantAction: "disabled",
		},
		{
			line:       `L 07/05/2026 - 00:00:02.000: Match unpaused`,
			wantAction: "unpaused",
		},
	}

	for _, tt := range tests {
		event, err := parser.ParseLine(tt.line)
		if err != nil {
			t.Fatalf("parse line %q: %v", tt.line, err)
		}
		pause, ok := event.(MatchPause)
		if !ok {
			t.Fatalf("event type = %T, want MatchPause", event)
		}
		if pause.Action != tt.wantAction {
			t.Fatalf("action = %q, want %q", pause.Action, tt.wantAction)
		}
	}
}

func TestParseLineReturnsTeamPlaying(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		line         string
		wantSide     Side
		wantTeamName string
	}{
		{
			line:         `L 07/05/2026 - 00:00:00.000: Team playing "TERRORIST": Terrorists`,
			wantSide:     SideTerrorist,
			wantTeamName: "Terrorists",
		},
		{
			line:         `L 07/05/2026 - 00:00:01.000: MatchStatus: Team playing "CT": Counter-Terrorists`,
			wantSide:     SideCT,
			wantTeamName: "Counter-Terrorists",
		},
	}

	for _, tt := range tests {
		event, err := parser.ParseLine(tt.line)
		if err != nil {
			t.Fatalf("parse line %q: %v", tt.line, err)
		}
		team, ok := event.(TeamPlaying)
		if !ok {
			t.Fatalf("event type = %T, want TeamPlaying", event)
		}
		if team.Side != tt.wantSide || team.TeamName != tt.wantTeamName {
			t.Fatalf("team playing = %#v", team)
		}
	}
}

func TestParseLineReturnsBombDefuseKitVariants(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		line    string
		wantKit bool
	}{
		{
			line:    `L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" triggered "Begin_Bomb_Defuse_With_Kit"`,
			wantKit: true,
		},
		{
			line:    `L 07/05/2026 - 00:00:01.000: "Player<1><STEAM_1:1:1><CT>" triggered "Begin_Bomb_Defuse_Without_Kit"`,
			wantKit: false,
		},
	}

	for _, tt := range tests {
		event, err := parser.ParseLine(tt.line)
		if err != nil {
			t.Fatalf("parse line %q: %v", tt.line, err)
		}
		defuse, ok := event.(PlayerBombBeginDefuse)
		if !ok {
			t.Fatalf("event type = %T, want PlayerBombBeginDefuse", event)
		}
		if defuse.Kit != tt.wantKit {
			t.Fatalf("kit = %t, want %t", defuse.Kit, tt.wantKit)
		}
	}
}

func TestParseLineReturnsBombObjectiveTriggers(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		line string
		want string
	}{
		{
			line: `L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><TERRORIST>" triggered "Got_The_Bomb"`,
			want: "PlayerBombGot",
		},
		{
			line: `L 07/05/2026 - 00:00:01.000: "Player<1><STEAM_1:1:1><TERRORIST>" triggered "Dropped_The_Bomb"`,
			want: "PlayerBombDropped",
		},
		{
			line: `L 07/05/2026 - 00:00:02.000: "Player<1><STEAM_1:1:1><CT>" triggered "Defused_The_Bomb"`,
			want: "PlayerBombDefused",
		},
	}

	for _, tt := range tests {
		event, err := parser.ParseLine(tt.line)
		if err != nil {
			t.Fatalf("parse line %q: %v", tt.line, err)
		}
		if event.EventType() != tt.want {
			t.Fatalf("event type = %s (%T), want %s", event.EventType(), event, tt.want)
		}
	}
}

func TestParseLineReturnsPlayerLeftBuyZoneItems(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.750: "ExamplePlayer<6><[U:1:123456789]><TERRORIST>" left buyzone with [ weapon_knife_t weapon_taser weapon_glock kevlar(100) helmet ]`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	leftBuyZone, ok := event.(PlayerLeftBuyZone)
	if !ok {
		t.Fatalf("event type = %T, want PlayerLeftBuyZone", event)
	}
	if leftBuyZone.Player.Name != "ExamplePlayer" || leftBuyZone.Player.SteamID != "[U:1:123456789]" {
		t.Fatalf("player = %#v", leftBuyZone.Player)
	}
	wantItems := []string{"weapon_knife_t", "weapon_taser", "weapon_glock", "kevlar(100)", "helmet"}
	if len(leftBuyZone.Items) != len(wantItems) {
		t.Fatalf("items = %#v, want %#v", leftBuyZone.Items, wantItems)
	}
	for i, want := range wantItems {
		if leftBuyZone.Items[i] != want {
			t.Fatalf("items[%d] = %q, want %q", i, leftBuyZone.Items[i], want)
		}
	}
}

func TestParseLineReturnsPlayerLeftBuyZoneEmptyItems(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.800: "ExampleBot<12><<none>><TERRORIST>" left buyzone with [ ]`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	leftBuyZone, ok := event.(PlayerLeftBuyZone)
	if !ok {
		t.Fatalf("event type = %T, want PlayerLeftBuyZone", event)
	}
	if leftBuyZone.Player.Name != "ExampleBot" || leftBuyZone.Player.SteamID != "<none>" {
		t.Fatalf("player = %#v", leftBuyZone.Player)
	}
	if len(leftBuyZone.Items) != 0 {
		t.Fatalf("items = %#v, want empty", leftBuyZone.Items)
	}
}

func TestParseLineReturnsPlayerValidated(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.875: "ExamplePlayer<65283><[U:1:123456789]><>" STEAM USERID validated`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	validated, ok := event.(PlayerValidated)
	if !ok {
		t.Fatalf("event type = %T, want PlayerValidated", event)
	}
	if validated.Player.Name != "ExamplePlayer" || validated.Player.UserID != "65283" || validated.Player.SteamID != "[U:1:123456789]" || validated.Player.Side != "" {
		t.Fatalf("player = %#v", validated.Player)
	}
}

func TestParseLineReturnsPlayerDisconnectedWithSpectatorSide(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.890: "ExamplePlayer<9><[U:1:123456789]><Spectator>" disconnected (reason "NETWORK_DISCONNECT_DISCONNECT_BY_USER")`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	disconnected, ok := event.(PlayerDisconnected)
	if !ok {
		t.Fatalf("event type = %T, want PlayerDisconnected", event)
	}
	if disconnected.Player.Side != "Spectator" || disconnected.Reason != "NETWORK_DISCONNECT_DISCONNECT_BY_USER" {
		t.Fatalf("disconnected = %#v", disconnected)
	}
}

func TestParseLineReturnsPlayerKillOther(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:01.500: "ExamplePlayer<3><[U:1:123456789]><CT>" [10 20 30] killed other "func_breakable<456>" [-10 -20 -30] with "ssg08" (noscope attackerinair)`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	killOther, ok := event.(PlayerKillOther)
	if !ok {
		t.Fatalf("event type = %T, want PlayerKillOther", event)
	}
	if killOther.Attacker.Name != "ExamplePlayer" || killOther.Target.Name != "func_breakable" || killOther.Target.ID != "456" {
		t.Fatalf("kill other = %#v", killOther)
	}
	if killOther.Weapon != "ssg08" {
		t.Fatalf("weapon = %q, want ssg08", killOther.Weapon)
	}
	wantFlags := []string{"noscope", "attackerinair"}
	if len(killOther.Flags) != len(wantFlags) {
		t.Fatalf("flags = %#v, want %#v", killOther.Flags, wantFlags)
	}
	for i, want := range wantFlags {
		if killOther.Flags[i] != want {
			t.Fatalf("flags[%d] = %q, want %q", i, killOther.Flags[i], want)
		}
	}
}

func TestParseLineReturnsPlayerKillFlags(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:01.750: "ExamplePlayer<3><[U:1:123456789]><CT>" [10 20 30] killed "OtherPlayer<4><[U:1:987654321]><TERRORIST>" [-10 -20 -30] with "m4a1_silencer" (headshot throughsmoke attackerblind)`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	kill, ok := event.(PlayerKill)
	if !ok {
		t.Fatalf("event type = %T, want PlayerKill", event)
	}
	if !kill.Headshot || kill.Penetrated {
		t.Fatalf("headshot/penetrated = %#v", kill)
	}
	wantFlags := []string{"headshot", "throughsmoke", "attackerblind"}
	if len(kill.Flags) != len(wantFlags) {
		t.Fatalf("flags = %#v, want %#v", kill.Flags, wantFlags)
	}
	for i, want := range wantFlags {
		if kill.Flags[i] != want {
			t.Fatalf("flags[%d] = %q, want %q", i, kill.Flags[i], want)
		}
	}
}

func TestParseLineReturnsPlayerFlashAssist(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:01.875: "ExamplePlayer<3><[U:1:123456789]><CT>" flash-assisted killing "OtherPlayer<4><[U:1:987654321]><TERRORIST>"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	assist, ok := event.(PlayerFlashAssist)
	if !ok {
		t.Fatalf("event type = %T, want PlayerFlashAssist", event)
	}
	if assist.Attacker.Name != "ExamplePlayer" || assist.Victim.Name != "OtherPlayer" {
		t.Fatalf("flash assist = %#v", assist)
	}
}

func TestParseLineReturnsPlayerAttackWithEmptyWeapon(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:02.000: "ExampleBot<0><BOT><CT>" [2512 -504 -343] attacked "ExampleBot<0><BOT><CT>" [2512 -504 -343] with "" (damage "1") (damage_armor "0") (health "0") (armor "0") (hitgroup "generic")`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	attack, ok := event.(PlayerAttack)
	if !ok {
		t.Fatalf("event type = %T, want PlayerAttack", event)
	}
	if attack.Weapon != "" || attack.Damage != 1 || attack.Hitgroup != "generic" {
		t.Fatalf("attack = %#v", attack)
	}
	if attack.Attacker.Name != attack.Victim.Name {
		t.Fatalf("attacker/victim = %#v/%#v", attack.Attacker, attack.Victim)
	}
}

func TestParseLineReturnsPlayerSvThrow(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:04.500: "ExamplePlayer<3><[U:1:123456789]><CT>" sv_throw_smokegrenade -380.930 -580.630 347.767 0.000 0.000 0.000 618.108 -11.430 -173.134 600.000 -478.000 0.000 45 2`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	svThrow, ok := event.(PlayerSvThrow)
	if !ok {
		t.Fatalf("event type = %T, want PlayerSvThrow", event)
	}
	if svThrow.Player.Name != "ExamplePlayer" || svThrow.Grenade != "smokegrenade" {
		t.Fatalf("sv throw = %#v", svThrow)
	}
	if len(svThrow.Values) != 14 {
		t.Fatalf("values length = %d, want 14: %#v", len(svThrow.Values), svThrow.Values)
	}
	if svThrow.Values[0] != -380.930 || svThrow.Values[12] != 45 || svThrow.Values[13] != 2 {
		t.Fatalf("values = %#v", svThrow.Values)
	}
}

func TestParseLineReturnsPlayerMoneyChangeReasons(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:04.600: "ExamplePlayer<3><[U:1:123456789]><CT>" money change 1000-300 = $700 (tracked) (purchase: weapon_hegrenade)`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	purchase, ok := event.(PlayerMoneyChange)
	if !ok {
		t.Fatalf("event type = %T, want PlayerMoneyChange", event)
	}
	if purchase.Purchase != "weapon_hegrenade" || purchase.AcquireReason != "" {
		t.Fatalf("purchase money change = %#v", purchase)
	}
	if purchase.Equation != (Equation{A: 1000, B: -300, Result: 700}) {
		t.Fatalf("purchase equation = %#v", purchase.Equation)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:04.650: "ExamplePlayer<3><[U:1:123456789]><CT>" money change 1000+650 = $1650 (tracked) (acquire: sellback)`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	acquire, ok := event.(PlayerMoneyChange)
	if !ok {
		t.Fatalf("event type = %T, want PlayerMoneyChange", event)
	}
	if acquire.AcquireReason != "sellback" || acquire.Purchase != "" {
		t.Fatalf("acquire money change = %#v", acquire)
	}
	if acquire.Equation != (Equation{A: 1000, B: 650, Result: 1650}) {
		t.Fatalf("acquire equation = %#v", acquire.Equation)
	}
}

func TestParseLineReturnsPlayerBlindedWithBotVictim(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:04.750: "ExampleBot<13><BOT><TERRORIST>" blinded for 4.83 by "ExamplePlayer<21><[U:1:123456789]><TERRORIST>" from flashbang entindex 685 `)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	blinded, ok := event.(PlayerBlinded)
	if !ok {
		t.Fatalf("event type = %T, want PlayerBlinded", event)
	}
	if blinded.Victim.Name != "ExampleBot" || blinded.Victim.SteamID != "BOT" || blinded.Attacker.Name != "ExamplePlayer" {
		t.Fatalf("blinded = %#v", blinded)
	}
	if blinded.Duration != 4.83 || blinded.Entindex != 685 {
		t.Fatalf("duration/entindex = %#v", blinded)
	}
}

func TestParseLineReturnsBombsiteTrigger(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.250: "ExamplePlayer<3><[U:1:123456789]><TERRORIST>" triggered "Bomb_Begin_Plant" at bombsite B`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	beginPlant, ok := event.(PlayerBombBeginPlant)
	if !ok {
		t.Fatalf("event type = %T, want PlayerBombBeginPlant", event)
	}
	if beginPlant.Player.Name != "ExamplePlayer" || beginPlant.Site != "B" {
		t.Fatalf("begin plant = %#v", beginPlant)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:03.500: "ExamplePlayer<3><[U:1:123456789]><TERRORIST>" triggered "Planted_The_Bomb" at bombsite A`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	planted, ok := event.(PlayerBombPlanted)
	if !ok {
		t.Fatalf("event type = %T, want PlayerBombPlanted", event)
	}
	if planted.Site != "A" {
		t.Fatalf("site = %q, want A", planted.Site)
	}
}

func TestParseLineReturnsHostagePositionEvents(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:12.000: "ExamplePlayer<3><[U:1:123456789]><CT>" [-1432 -64 144] picked up a hostage`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	pickedUp, ok := event.(PlayerPickedUpHostage)
	if !ok {
		t.Fatalf("event type = %T, want PlayerPickedUpHostage", event)
	}
	if pickedUp.Player.Name != "ExamplePlayer" || pickedUp.Position != (Position{X: -1432, Y: -64, Z: 144}) {
		t.Fatalf("picked up hostage = %#v", pickedUp)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:13.000: "ExamplePlayer<3><[U:1:123456789]><CT>" [888 653 -8] dropped off a hostage`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	droppedOff, ok := event.(PlayerDroppedOffHostage)
	if !ok {
		t.Fatalf("event type = %T, want PlayerDroppedOffHostage", event)
	}
	if droppedOff.Player.Name != "ExamplePlayer" || droppedOff.Position != (Position{X: 888, Y: 653, Z: -8}) {
		t.Fatalf("dropped off hostage = %#v", droppedOff)
	}
}

func TestParseLineReturnsVoteEvents(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:14.000: Vote started "ChangeLevel de_cache" from #14 "ExamplePlayer<14><[U:1:123456789]><CT><Area 1595>"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	started, ok := event.(VoteStarted)
	if !ok {
		t.Fatalf("event type = %T, want VoteStarted", event)
	}
	if started.Issue != "ChangeLevel de_cache" || started.Actor.Player.Name != "ExamplePlayer" || started.Actor.Slot != 14 || started.Actor.Area != 1595 {
		t.Fatalf("started = %#v", started)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:15.000: Vote cast "Kick ExamplePlayer" from #0 "OtherPlayer<0><[U:1:987654321]><TERRORIST><Area -1>" option1`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	cast, ok := event.(VoteCast)
	if !ok {
		t.Fatalf("event type = %T, want VoteCast", event)
	}
	if cast.Issue != "Kick ExamplePlayer" || cast.Actor.Player.Side != "TERRORIST" || cast.Actor.Area != -1 || cast.Option != 1 {
		t.Fatalf("cast = %#v", cast)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:16.000: Vote succeeded "ChangeLevel de_cache" from #14 "ExamplePlayer<14><[U:1:123456789]><CT><Area 1595>"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	succeeded, ok := event.(VoteSucceeded)
	if !ok {
		t.Fatalf("event type = %T, want VoteSucceeded", event)
	}
	if succeeded.Issue != "ChangeLevel de_cache" || succeeded.Actor.Player.SteamID != "[U:1:123456789]" {
		t.Fatalf("succeeded = %#v", succeeded)
	}

	event, err = parser.ParseLine(`L 07/05/2026 - 00:00:17.000: Vote failed "Kick ExamplePlayer" `)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	failed, ok := event.(VoteFailed)
	if !ok {
		t.Fatalf("event type = %T, want VoteFailed", event)
	}
	if failed.Issue != "Kick ExamplePlayer" {
		t.Fatalf("failed = %#v", failed)
	}
}

func TestParseLineReturnsAccolade(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:07.500: ACCOLADE, FINAL: {uniqueweaponkills},	ExamplePlayer<3>,	VALUE: 6.000000,	POS: 1,	SCORE: 70.000000`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}
	accolade, ok := event.(Accolade)
	if !ok {
		t.Fatalf("event type = %T, want Accolade", event)
	}
	if accolade.Phase != "FINAL" || accolade.Metric != "uniqueweaponkills" || accolade.PlayerName != "ExamplePlayer" || accolade.UserID != "3" {
		t.Fatalf("accolade = %#v", accolade)
	}
	if accolade.Value != 6 || accolade.Position != 1 || accolade.Score != 70 {
		t.Fatalf("accolade numbers = %#v", accolade)
	}
}

func TestParseLineReturnsServerCvar(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.900: server_cvar: "mp_freezetime" "12"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	cvar, ok := event.(ServerCvar)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvar", event)
	}
	if cvar.Name != "mp_freezetime" || cvar.Value != "12" || cvar.Sensitive {
		t.Fatalf("cvar = %#v", cvar)
	}
	if cvar.RawLine() != `L 07/05/2026 - 00:00:03.900: server_cvar: "mp_freezetime" "12"` {
		t.Fatalf("raw line = %q", cvar.RawLine())
	}
}

func TestParseLineReturnsLogFileStarted(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.850: Log file started (file "logs/example.log") (game "csgo") (version "10072")`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	started, ok := event.(LogFileStarted)
	if !ok {
		t.Fatalf("event type = %T, want LogFileStarted", event)
	}
	if started.File != "logs/example.log" || started.Game != "csgo" || started.Version != "10072" {
		t.Fatalf("log file started = %#v", started)
	}
}

func TestParseLineReturnsServerCvarAssignment(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:03.950: "cash_player_interact_with_hostage" = "300"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	cvar, ok := event.(ServerCvar)
	if !ok {
		t.Fatalf("event type = %T, want ServerCvar", event)
	}
	if cvar.Name != "cash_player_interact_with_hostage" || cvar.Value != "300" || cvar.Sensitive {
		t.Fatalf("cvar = %#v", cvar)
	}
}

func TestParseLineRedactsSensitiveServerCvar(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		name string
		line string
	}{
		{
			name: "exact sensitive name",
			line: `L 07/05/2026 - 00:00:04.000: server_cvar: "rcon_password" "fake-secret-password"`,
		},
		{
			name: "sensitive marker",
			line: `07/05/2026 - 00:00:05.000 - server_cvar: "my_api_token" "fake-secret-token"`,
		},
		{
			name: "sensitive assignment",
			line: `07/05/2026 - 00:00:06.000 - "sv_password" = "fake-secret-assignment"`,
		},
		{
			name: "mixed case sensitive marker",
			line: `07/05/2026 - 00:00:07.000 - server_cvar: "My_API_Key" "fake-secret-key"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.ParseLine(tt.line)
			if err != nil {
				t.Fatalf("parse line: %v", err)
			}

			cvar, ok := event.(ServerCvar)
			if !ok {
				t.Fatalf("event type = %T, want ServerCvar", event)
			}
			if !cvar.Sensitive {
				t.Fatalf("sensitive = false")
			}
			if cvar.Value != "" {
				t.Fatalf("value = %q, want empty", cvar.Value)
			}
			if strings.Contains(cvar.RawLine(), "fake-secret") {
				t.Fatalf("raw line contains secret: %q", cvar.RawLine())
			}
			if !strings.Contains(cvar.RawLine(), "[REDACTED]") {
				t.Fatalf("raw line = %q, want redacted marker", cvar.RawLine())
			}
		})
	}
}

func TestServerCvarRedactionFallbackDoesNotIncludeSecret(t *testing.T) {
	cvar := buildServerCvar(BaseEvent{
		Type: "ServerCvar",
		Raw:  `unrecognized raw text containing fake-secret-password`,
	}, "rcon_password", "fake-secret-password")

	if !cvar.Sensitive {
		t.Fatal("sensitive = false")
	}
	if cvar.Value != "" {
		t.Fatalf("value = %q, want empty", cvar.Value)
	}
	if strings.Contains(cvar.RawLine(), "fake-secret-password") {
		t.Fatalf("raw line contains secret: %q", cvar.RawLine())
	}
	if cvar.RawLine() != `server_cvar: "rcon_password" "[REDACTED]"` {
		t.Fatalf("raw line = %q", cvar.RawLine())
	}
}

func TestSensitiveCvarDetectionTrimsAndLowercasesName(t *testing.T) {
	if !isSensitiveCvar("  RCON_PASSWORD  ") {
		t.Fatal("RCON_PASSWORD was not detected as sensitive")
	}
}

func TestParseLineReturnsIgnoredForKnownNoise(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	tests := []struct {
		name        string
		line        string
		wantReason  string
		wantPayload string
	}{
		{
			name:        "match status team unset",
			line:        `L 07/05/2026 - 00:00:06.000: MatchStatus: Team "CT" is unset.`,
			wantReason:  "MatchStatusTeamUnset",
			wantPayload: `MatchStatus: Team "CT" is unset.`,
		},
		{
			name:        "log file closed",
			line:        `L 07/05/2026 - 00:00:07.000: Log file closed`,
			wantReason:  "LogFileClosed",
			wantPayload: `Log file closed`,
		},
		{
			name:        "started",
			line:        `L 07/05/2026 - 00:00:07.125: Started:  ""`,
			wantReason:  "Started",
			wantPayload: `Started:  ""`,
		},
		{
			name:        "server cvars start",
			line:        `L 07/05/2026 - 00:00:07.250: server cvars start`,
			wantReason:  "ServerCvarsMarker",
			wantPayload: `server cvars start`,
		},
		{
			name:        "server cvars end",
			line:        `L 07/05/2026 - 00:00:07.500: server cvars end`,
			wantReason:  "ServerCvarsMarker",
			wantPayload: `server cvars end`,
		},
		{
			name:        "rcon command",
			line:        `L 07/05/2026 - 00:00:07.750: rcon from "127.0.0.1:35484": command "bot_quota 18"`,
			wantReason:  "RconCommand",
			wantPayload: `rcon from "127.0.0.1:35484": command "bot_quota 18"`,
		},
		{
			name:        "meta plugin loaded",
			line:        `L 07/05/2026 - 00:00:07.875: [META] Loaded 0 plugins (2 already loaded)`,
			wantReason:  "MetaPluginLoaded",
			wantPayload: `[META] Loaded 0 plugins (2 already loaded)`,
		},
		{
			name:        "round stats json begin",
			line:        `L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`,
			wantReason:  "RoundStatsJSONFragment",
			wantPayload: `JSON_BEGIN{`,
		},
		{
			name:        "round stats json scalar",
			line:        `L 07/05/2026 - 00:00:09.000: "score_ct" : "13",`,
			wantReason:  "RoundStatsJSONFragment",
			wantPayload: `"score_ct" : "13",`,
		},
		{
			name:        "round stats json player",
			line:        `L 07/05/2026 - 00:00:10.000: "player_0" : "           123456789,      3,  16000,      0"`,
			wantReason:  "RoundStatsJSONFragment",
			wantPayload: `"player_0" : "           123456789,      3,  16000,      0"`,
		},
		{
			name:        "round stats json end",
			line:        `L 07/05/2026 - 00:00:11.000: }}JSON_END`,
			wantReason:  "RoundStatsJSONFragment",
			wantPayload: `}}JSON_END`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.ParseLine(tt.line)
			if err != nil {
				t.Fatalf("parse line: %v", err)
			}

			ignored, ok := event.(Ignored)
			if !ok {
				t.Fatalf("event type = %T, want Ignored", event)
			}
			if ignored.Reason != tt.wantReason {
				t.Fatalf("reason = %q, want %s", ignored.Reason, tt.wantReason)
			}
			if ignored.Payload != tt.wantPayload {
				t.Fatalf("payload = %q, want %q", ignored.Payload, tt.wantPayload)
			}
		})
	}
}

func TestParseLineAcceptsCapturedLogFileFormat(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`07/05/2026 - 00:00:00.000 - "Player<1><STEAM_1:1:1><CT>" say "hello"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	playerSay, ok := event.(PlayerSay)
	if !ok {
		t.Fatalf("event type = %T, want PlayerSay", event)
	}
	if playerSay.Message != "hello" {
		t.Fatalf("message = %q, want hello", playerSay.Message)
	}
}

func TestParseLineAcceptsCapturedLogFileFormatWithoutMilliseconds(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`07/05/2026 - 00:00:00 - Unrecognized payload "de_dust2"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	if _, ok := event.(Unknown); !ok {
		t.Fatalf("event type = %T, want Unknown", event)
	}
}

func TestParseLineReturnsUnknownForKnownTimestampWithUnknownPayload(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: Unrecognized payload "de_dust2"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	unknown, ok := event.(Unknown)
	if !ok {
		t.Fatalf("event type = %T, want Unknown", event)
	}
	if unknown.Payload != `Unrecognized payload "de_dust2"` {
		t.Fatalf("payload = %q", unknown.Payload)
	}
}

func TestParseLineReturnsMapLoadingStarted(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: Loading map "cs_agency"`)
	if err != nil {
		t.Fatalf("parse line: %v", err)
	}

	mapLoading, ok := event.(MapLoadingStarted)
	if !ok {
		t.Fatalf("event type = %T, want MapLoadingStarted", event)
	}
	if mapLoading.Map != "cs_agency" {
		t.Fatalf("map = %q, want cs_agency", mapLoading.Map)
	}
}

func TestParseLineReturnsErrNoMatch(t *testing.T) {
	parser, err := NewParser(Config{LogTimezone: "UTC"})
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	_, err = parser.ParseLine(`not a cs2 log line`)
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("error = %v, want ErrNoMatch", err)
	}
}
