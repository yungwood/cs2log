package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/stream"
)

type streamReport struct {
	files       int
	records     int
	errors      int
	events      map[string]int
	jsonBlocks  map[string]int
	jsonInvalid int
	roundStats  int
	serverCvars int
}

func runStream(args []string) error {
	flags := flag.NewFlagSet("stream", flag.ContinueOnError)
	flags.SetOutput(stderr)
	timezone := flags.String("timezone", "UTC", "IANA timezone used to interpret CS2 log timestamps")
	printBlocks := flags.Bool("blocks", false, "print each multiline block")
	if err := flags.Parse(args); err != nil {
		return err
	}

	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: *timezone})
	if err != nil {
		return fmt.Errorf("new parser: %w", err)
	}

	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{"-"}
	}

	r := streamReport{
		events:     map[string]int{},
		jsonBlocks: map[string]int{},
	}
	for _, path := range paths {
		if err := scanStreamPath(parser, path, *printBlocks, &r); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	printStreamReport(stdout, r)
	return nil
}

func scanStreamPath(parser *cs2log.Parser, path string, printBlocks bool, r *streamReport) error {
	var reader io.Reader
	if path == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		reader = file
	}

	r.files++
	processor := stream.NewProcessor(parser)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		collectStreamRecords(path, processor.PushLine(scanner.Text()), printBlocks, r)
	}
	collectStreamRecords(path, processor.Flush(), printBlocks, r)
	return scanner.Err()
}

func collectStreamRecords(path string, records []stream.Record, printBlocks bool, r *streamReport) {
	for _, record := range records {
		r.records++
		if record.Err != nil {
			r.errors++
			continue
		}
		r.events[record.Event.EventType()]++
		if block, ok := record.Event.(stream.JSONBlock); ok {
			r.jsonBlocks[string(block.Status)]++
			if !block.ValidJSON {
				r.jsonInvalid++
			}
			if printBlocks {
				fmt.Fprintf(stdout, "%s:%d-%d JSONBlock status=%s valid=%t name=%q\n", path, record.LineStart, record.LineEnd, block.Status, block.ValidJSON, block.Name)
			}
		}
		if roundStats, ok := record.Event.(stream.RoundStats); ok {
			r.roundStats++
			if printBlocks {
				fmt.Fprintf(stdout, "%s:%d-%d RoundStats round=%d score=%d-%d map=%q players=%d\n", path, record.LineStart, record.LineEnd, roundStats.RoundNumber, roundStats.ScoreT, roundStats.ScoreCT, roundStats.Map, len(roundStats.Players))
			}
		}
		if serverCvars, ok := record.Event.(stream.ServerCvars); ok {
			r.serverCvars++
			if printBlocks {
				fmt.Fprintf(stdout, "%s:%d-%d ServerCvars status=%s cvars=%d\n", path, record.LineStart, record.LineEnd, serverCvars.Status, len(serverCvars.Cvars))
			}
		}
	}
}

func printStreamReport(w io.Writer, r streamReport) {
	fmt.Fprintf(w, "files: %d\n", r.files)
	fmt.Fprintf(w, "records: %d\n", r.records)
	fmt.Fprintf(w, "errors: %d\n", r.errors)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "events:")
	printCounts(w, r.events, 0)

	if len(r.jsonBlocks) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "json blocks:")
		printCounts(w, r.jsonBlocks, 0)
		fmt.Fprintf(w, "invalid json bodies: %d\n", r.jsonInvalid)
	}
	if r.roundStats > 0 {
		fmt.Fprintf(w, "round stats events: %d\n", r.roundStats)
	}
	if r.serverCvars > 0 {
		fmt.Fprintf(w, "server cvars events: %d\n", r.serverCvars)
	}
}
