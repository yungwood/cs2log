package cs2log_test

import (
	"fmt"
	"time"

	"github.com/yungwood/cs2log"
)

func ExampleNewParser() {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "America/New_York"})
	if err != nil {
		panic(err)
	}

	fmt.Println(parser.Location())
	// Output: America/New_York
}

func ExampleParser_ParseLine() {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "America/New_York"})
	if err != nil {
		panic(err)
	}

	event, err := parser.ParseLine(`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`)
	if err != nil {
		panic(err)
	}

	fmt.Println(event.EventType())
	fmt.Println(event.Timestamp().Format(time.RFC3339))
	// Output:
	// PlayerSay
	// 2026-07-05T04:00:00Z
}

func ExampleDecodeSteamID() {
	steamID, err := cs2log.DecodeSteamID("[U:1:0]")
	if err != nil {
		panic(err)
	}

	fmt.Println(steamID.AccountID)
	fmt.Println(steamID.SteamID())
	fmt.Println(steamID.SteamID3())
	fmt.Println(steamID.SteamID64())
	// Output:
	// 0
	// STEAM_1:0:0
	// [U:1:0]
	// 76561197960265728
}
