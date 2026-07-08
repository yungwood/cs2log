package cs2log

import "testing"

// Use the zero account fixture so tests do not contain a real player Steam ID.
func TestDecodeSteamID3(t *testing.T) {
	steamID, err := DecodeSteamID("[U:1:0]")
	if err != nil {
		t.Fatalf("decode steam id: %v", err)
	}
	if steamID.AccountID != 0 || steamID.Universe != 1 {
		t.Fatalf("steam id = %#v", steamID)
	}
	if steamID.SteamID64() != 76561197960265728 {
		t.Fatalf("steam id64 = %d", steamID.SteamID64())
	}
	if steamID.SteamID3() != "[U:1:0]" {
		t.Fatalf("steam id3 = %q", steamID.SteamID3())
	}
	if steamID.SteamID() != "STEAM_1:0:0" {
		t.Fatalf("steam id = %q", steamID.SteamID())
	}
}

func TestDecodeLegacySteamID(t *testing.T) {
	steamID, err := DecodeSteamID("STEAM_1:0:0")
	if err != nil {
		t.Fatalf("decode steam id: %v", err)
	}
	if steamID.AccountID != 0 || steamID.SteamID3() != "[U:1:0]" {
		t.Fatalf("steam id = %#v", steamID)
	}
}

func TestDecodeLegacySteamIDNormalizesPublicUniverseZero(t *testing.T) {
	steamID, err := DecodeSteamID("STEAM_0:0:0")
	if err != nil {
		t.Fatalf("decode steam id: %v", err)
	}
	if steamID.Universe != 1 || steamID.SteamID() != "STEAM_1:0:0" || steamID.SteamID3() != "[U:1:0]" {
		t.Fatalf("steam id = %#v legacy=%q steam3=%q", steamID, steamID.SteamID(), steamID.SteamID3())
	}
}

func TestDecodeSteamID64(t *testing.T) {
	steamID, err := DecodeSteamID("76561197960265728")
	if err != nil {
		t.Fatalf("decode steam id: %v", err)
	}
	if steamID.AccountID != 0 || steamID.SteamID3() != "[U:1:0]" {
		t.Fatalf("steam id = %#v", steamID)
	}
}

func TestDecodeSteamAccountID(t *testing.T) {
	steamID, err := DecodeSteamID("0")
	if err != nil {
		t.Fatalf("decode steam id: %v", err)
	}
	if steamID.AccountID != 0 || steamID.SteamID64() != 76561197960265728 {
		t.Fatalf("steam id = %#v id64=%d", steamID, steamID.SteamID64())
	}
}

func TestSteamIdentityZeroUniverseFormatsAsPublic(t *testing.T) {
	steamID := SteamIdentity{AccountID: 1}
	if steamID.SteamID3() != "[U:1:1]" {
		t.Fatalf("steam id3 = %q", steamID.SteamID3())
	}
	if steamID.SteamID() != "STEAM_1:1:0" {
		t.Fatalf("steam id = %q", steamID.SteamID())
	}
}

func TestDecodeSteamIDRejectsUnsupportedValues(t *testing.T) {
	tests := []string{
		"",
		"BOT",
		"<none>",
		"[U:999999999999999999999:0]",
		"[U:1:999999999999999999999]",
		"STEAM_999999999999999999999:0:0",
		"STEAM_1:0:999999999999999999999",
		"76561197960265727",
		"not-a-steam-id",
	}
	for _, value := range tests {
		if _, err := DecodeSteamID(value); err == nil {
			t.Fatalf("DecodeSteamID(%q) succeeded, want error", value)
		}
	}
}
