package cs2log

var matchStatusTeamUnsetDefinition = ignoredDefinition(
	"MatchStatusTeamUnset",
	`MatchStatus: Team "(?:CT|TERRORIST)" is unset\.`,
)

var logFileClosedDefinition = ignoredDefinition(
	"LogFileClosed",
	`Log file closed`,
)

var startedDefinition = ignoredDefinition(
	"Started",
	`Started:\s*""`,
)

var serverCvarsMarkerDefinition = ignoredDefinition(
	"ServerCvarsMarker",
	`server cvars (?:start|end)`,
)

var rconCommandDefinition = ignoredDefinition(
	"RconCommand",
	`rcon from "[^"]+": command "[^"]*"`,
)

var metaPluginLoadedDefinition = ignoredDefinition(
	"MetaPluginLoaded",
	`\[META\] Loaded \d+ plugins? \(\d+ already loaded\)`,
)

var roundStatsJSONFragmentDefinition = ignoredDefinition(
	"RoundStatsJSONFragment",
	`(?:JSON_BEGIN\{|\}+JSON_END|"name"\s*:\s*"round_stats",?|"round_number"\s*:\s*"\d+",?|"score_t"\s*:\s*"\d+",?|"score_ct"\s*:\s*"\d+",?|"map"\s*:\s*"[^"]+",?|"server"\s*:\s*"[^"]+",?|"fields"\s*:\s*"[^"]+",?|"players"\s*:\s*\{|"player_\d+"\s*:\s*"[^"]+",?)`,
)

func ignoredDefinition(reason, regex string) definition {
	return definition{
		Type:        "Ignored",
		Category:    "ignored",
		Description: "A known log line that is intentionally ignored.",
		Regex:       "(" + regex + ")",
		Fields:      []field{{Name: "payload", Type: "string"}},
		Build: func(base BaseEvent, matches Matches) (Event, error) {
			return Ignored{
				BaseEvent: base,
				Reason:    reason,
				Payload:   matches.String("payload"),
			}, nil
		},
	}
}
