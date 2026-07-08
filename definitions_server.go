package cs2log

import (
	"fmt"
	"strings"
)

// ServerMessage is emitted for server_message log lines.
type ServerMessage struct {
	BaseEvent
	Text string `json:"text"`
}

// ServerCvar is emitted for startup server cvar dumps. Sensitive cvars are
// redacted so secrets do not survive in RawLine or Value.
type ServerCvar struct {
	BaseEvent
	Name      string `json:"name"`
	Value     string `json:"value,omitempty"`
	Sensitive bool   `json:"sensitive"`
}

// LogFileStarted is emitted when the server opens a log file.
type LogFileStarted struct {
	BaseEvent
	File    string `json:"file"`
	Game    string `json:"game"`
	Version string `json:"version"`
}

// MapLoadingStarted is emitted when the server starts loading a map.
type MapLoadingStarted struct {
	BaseEvent
	Map string `json:"map"`
}

var serverMessageDefinition = definition{
	Type:        "ServerMessage",
	Category:    "server",
	Description: "A server message was emitted.",
	Regex:       `server_message: "(\w+)"`,
	Fields:      []field{{Name: "text", Type: "word"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return ServerMessage{BaseEvent: base, Text: matches.String("text")}, nil
	},
}

var logFileStartedDefinition = definition{
	Type:        "LogFileStarted",
	Category:    "server",
	Description: "The server opened a log file.",
	Regex:       `Log file started \(file "([^"]+)"\) \(game "([^"]+)"\) \(version "([^"]+)"\)`,
	Fields: []field{
		{Name: "file", Type: "string"},
		{Name: "game", Type: "string"},
		{Name: "version", Type: "string"},
	},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return LogFileStarted{
			BaseEvent: base,
			File:      matches.String("file"),
			Game:      matches.String("game"),
			Version:   matches.String("version"),
		}, nil
	},
}

var mapLoadingStartedDefinition = definition{
	Type:        "MapLoadingStarted",
	Category:    "server",
	Description: "The server started loading a map.",
	Regex:       `Loading map "([^"]+)"`,
	Fields:      []field{{Name: "map", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return MapLoadingStarted{BaseEvent: base, Map: matches.String("map")}, nil
	},
}

var serverCvarDefinition = definition{
	Type:        "ServerCvar",
	Category:    "server",
	Description: "A server cvar was emitted during startup/config logging.",
	Regex:       `server_cvar: "([^"]+)" "(.*)"`,
	Fields:      []field{{Name: "name", Type: "string"}, {Name: "value", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return buildServerCvar(base, matches.String("name"), matches.String("value")), nil
	},
}

var serverCvarAssignmentDefinition = definition{
	Type:        "ServerCvar",
	Category:    "server",
	Description: "A server cvar assignment was emitted during startup/config logging.",
	Regex:       `"([^"]+)" = "(.*)"`,
	Fields:      []field{{Name: "name", Type: "string"}, {Name: "value", Type: "string"}},
	Build: func(base BaseEvent, matches Matches) (Event, error) {
		return buildServerCvar(base, matches.String("name"), matches.String("value")), nil
	},
}

func buildServerCvar(base BaseEvent, name, value string) ServerCvar {
	sensitive := isSensitiveCvar(name)
	if sensitive {
		value = ""
		base.Raw = redactedServerCvarRaw(base.Raw, name)
	}
	return ServerCvar{
		BaseEvent: base,
		Name:      name,
		Value:     value,
		Sensitive: sensitive,
	}
}

func isSensitiveCvar(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch normalized {
	case "sv_password", "rcon_password":
		return true
	}
	for _, marker := range []string{
		"password",
		"passwd",
		"secret",
		"token",
		"apikey",
		"api_key",
		"key",
		"auth",
		"credential",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func redactedServerCvarRaw(raw, name string) string {
	redactedPayload := fmt.Sprintf(`server_cvar: "%s" "[REDACTED]"`, name)
	if index := strings.Index(raw, "server_cvar:"); index >= 0 {
		return raw[:index] + redactedPayload
	}
	if index := strings.Index(raw, `"`+name+`" = `); index >= 0 {
		return raw[:index] + fmt.Sprintf(`"%s" = "[REDACTED]"`, name)
	}
	return redactedPayload
}
