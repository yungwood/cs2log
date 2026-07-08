package matchstate_test

import (
	"fmt"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/matchstate"
	"github.com/yungwood/cs2log/stream"
)

func ExampleTracker() {
	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: "UTC"})
	if err != nil {
		panic(err)
	}
	processor := stream.NewProcessor(parser)
	tracker := matchstate.NewTracker()

	lines := []string{
		`L 07/05/2026 - 00:00:00.000: World triggered "Match_Start" on "de_train"`,
		`L 07/05/2026 - 00:00:01.000: World triggered "Round_Start"`,
	}
	for _, line := range lines {
		for _, record := range processor.PushLine(line) {
			enriched := tracker.Push(record)
			fmt.Printf("%s %s %t\n", enriched.Event.EventType(), enriched.Context.Map, enriched.Context.RoundLive)
		}
	}

	// Output:
	// WorldMatchStart de_train false
	// WorldRoundStart de_train true
}
