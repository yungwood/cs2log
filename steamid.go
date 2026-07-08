package cs2log

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const steamID64Base uint64 = 76561197960265728

var (
	steamID3Pattern = regexp.MustCompile(`^\[U:(\d+):(\d+)\]$`)
	steamID2Pattern = regexp.MustCompile(`^STEAM_(\d+):([01]):(\d+)$`)
)

// SteamIdentity describes a decoded Steam identifier.
type SteamIdentity struct {
	Raw string `json:"raw"`
	// AccountID is the raw 32-bit Steam account ID, also known as SteamID32.
	AccountID uint32 `json:"accountId"`
	Universe  uint32 `json:"universe"`
}

// SteamID64 returns the 64-bit Steam ID for this account.
func (s SteamIdentity) SteamID64() uint64 {
	return steamID64Base + uint64(s.AccountID)
}

// SteamID3 returns the modern bracketed Steam ID representation.
func (s SteamIdentity) SteamID3() string {
	universe := s.Universe
	if universe == 0 {
		universe = 1
	}
	return fmt.Sprintf("[U:%d:%d]", universe, s.AccountID)
}

// SteamID returns the legacy STEAM_X:Y:Z SteamID format.
func (s SteamIdentity) SteamID() string {
	universe := s.Universe
	if universe == 0 {
		universe = 1
	}
	y := s.AccountID % 2
	z := s.AccountID / 2
	return fmt.Sprintf("STEAM_%d:%d:%d", universe, y, z)
}

// DecodeSteamID decodes common Steam ID forms found in CS2 logs.
func DecodeSteamID(value string) (SteamIdentity, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "BOT" || value == "<none>" {
		return SteamIdentity{}, fmt.Errorf("unsupported steam id %q", value)
	}

	if result := steamID3Pattern.FindStringSubmatch(value); result != nil {
		universe, err := parseUint32(result[1])
		if err != nil {
			return SteamIdentity{}, fmt.Errorf("parse steam id universe: %w", err)
		}
		accountID, err := parseUint32(result[2])
		if err != nil {
			return SteamIdentity{}, fmt.Errorf("parse steam account id: %w", err)
		}
		return SteamIdentity{Raw: value, AccountID: accountID, Universe: universe}, nil
	}

	if result := steamID2Pattern.FindStringSubmatch(value); result != nil {
		universe, err := parseUint32(result[1])
		if err != nil {
			return SteamIdentity{}, fmt.Errorf("parse steam id universe: %w", err)
		}
		if universe == 0 {
			universe = 1
		}
		y, err := parseUint32(result[2])
		if err != nil {
			return SteamIdentity{}, fmt.Errorf("parse steam id auth server: %w", err)
		}
		z, err := parseUint32(result[3])
		if err != nil {
			return SteamIdentity{}, fmt.Errorf("parse steam account number: %w", err)
		}
		return SteamIdentity{Raw: value, AccountID: z*2 + y, Universe: universe}, nil
	}

	id64, err := strconv.ParseUint(value, 10, 64)
	if err == nil && id64 >= steamID64Base {
		return SteamIdentity{Raw: value, AccountID: uint32(id64 - steamID64Base), Universe: 1}, nil
	}

	accountID, err := parseUint32(value)
	if err == nil {
		return SteamIdentity{Raw: value, AccountID: accountID, Universe: 1}, nil
	}

	return SteamIdentity{}, fmt.Errorf("unsupported steam id %q", value)
}

func parseUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}
