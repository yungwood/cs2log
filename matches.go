package cs2log

import (
	"strconv"
	"strings"
)

// Matches contains raw field captures for a matched event pattern.
type Matches struct {
	values map[string]string
}

// String returns a captured string field.
func (m Matches) String(name string) string {
	return m.values[name]
}

// Player returns a captured player field.
func (m Matches) Player(name string) (Player, error) {
	return parsePlayer(m.values[name])
}

// Int returns a captured integer field, or zero when an optional field is empty.
func (m Matches) Int(name string) (int, error) {
	value := strings.TrimSpace(m.values[name])
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// Float64 returns a captured floating-point field, or zero when an optional field is empty.
func (m Matches) Float64(name string) (float64, error) {
	value := strings.TrimSpace(m.values[name])
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// BoolPresence returns true when a captured optional field is not empty.
func (m Matches) BoolPresence(name string) bool {
	return m.values[name] != ""
}
