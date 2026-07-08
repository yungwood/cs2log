package stream_test

import (
	"fmt"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/stream"
)

func ExampleProcessor() {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		panic(err)
	}
	processor := stream.NewProcessor(parser)

	lines := []string{
		`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`,
		`L 07/05/2026 - 00:00:01.000: World triggered "Round_Start"`,
	}
	for _, line := range lines {
		for _, record := range processor.PushLine(line) {
			fmt.Printf("%d %s\n", record.LineStart, record.Event.EventType())
		}
	}

	// Output:
	// 1 PlayerSay
	// 2 WorldRoundStart
}
