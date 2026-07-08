package cs2log

// definition describes how to parse one CS2 log event and provides metadata
// that can be used for tests and documentation.
type definition struct {
	Type        string
	Category    string
	Description string
	Regex       string
	Fields      []field
	Examples    []example
	Build       func(BaseEvent, Matches) (Event, error)
}

// field describes a named capture in an event regex.
type field struct {
	Name        string
	Type        string
	Description string
}

// example describes one executable parser example.
type example struct {
	Line     string
	Timezone string
	UTC      string
}

// defaultDefinitions is the built-in parser registry.
var defaultDefinitions = []definition{
	// definitions_server.go
	serverMessageDefinition,
	logFileStartedDefinition,
	mapLoadingStartedDefinition,
	serverCvarDefinition,
	serverCvarAssignmentDefinition,

	// definitions_world.go
	freezTimeStartDefinition,
	worldMatchStartDefinition,
	worldRoundStartDefinition,
	worldRoundRestartDefinition,
	worldRoundEndDefinition,
	worldWarmupStartDefinition,
	worldWarmupEndDefinition,
	worldGameCommencingDefinition,

	// definitions_team.go
	teamScoredDefinition,
	teamPlayingDefinition,
	teamNoticeDefinition,

	// definitions_vote.go
	voteStartedDefinition,
	voteCastDefinition,
	voteSucceededDefinition,
	voteFailedDefinition,

	// definitions_player.go
	playerConnectedDefinition,
	playerDisconnectedDefinition,
	playerEnteredDefinition,
	playerValidatedDefinition,
	playerBannedDefinition,
	playerSwitchedDefinition,
	playerPurchaseDefinition,
	playerPickedUpDefinition,
	playerDroppedDefinition,
	playerLeftBuyZoneDefinition,
	playerMoneyChangeDefinition,
	playerTouchedHostageDefinition,
	playerRescuedHostageDefinition,
	playerPickedUpHostageDefinition,
	playerDroppedOffHostageDefinition,

	// definitions_chat.go
	playerSayDefinition,

	// definitions_combat.go
	playerKillDefinition,
	playerKillOtherDefinition,
	playerKillAssistDefinition,
	playerFlashAssistDefinition,
	playerAttackDefinition,
	playerKilledBombDefinition,
	playerKilledSuicideDefinition,

	// definitions_bomb.go
	playerBombGotDefinition,
	playerBombBeginPlantDefinition,
	playerBombPlantedDefinition,
	playerBombDroppedDefinition,
	playerBombBeginDefuseDefinition,
	playerBombDefusedDefinition,

	// definitions_projectile.go
	playerThrewDefinition,
	playerSvThrowDefinition,
	playerBlindedDefinition,
	projectileSpawnedDefinition,

	// definitions_game.go
	matchPauseDefinition,
	matchStatusScoreDefinition,
	gameOverDefinition,

	// definitions_accolade.go
	accoladeDefinition,

	// definitions_ignored.go
	matchStatusTeamUnsetDefinition,
	logFileClosedDefinition,
	startedDefinition,
	serverCvarsMarkerDefinition,
	rconCommandDefinition,
	metaPluginLoadedDefinition,
	roundStatsJSONFragmentDefinition,
}
