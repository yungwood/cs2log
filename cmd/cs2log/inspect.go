package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/matchstate"
	"github.com/yungwood/cs2log/stream"
)

type inspectOptions struct {
	format       string
	limit        int
	raw          bool
	stream       bool
	includeState bool
	types        map[string]bool
}

type inspectRecord struct {
	File      string              `json:"file"`
	LineStart int                 `json:"lineStart"`
	LineEnd   int                 `json:"lineEnd"`
	Type      string              `json:"type"`
	Event     interface{}         `json:"event"`
	Context   *matchstate.Context `json:"context,omitempty"`
	Raw       string              `json:"raw,omitempty"`
	Error     string              `json:"error,omitempty"`
}

func runInspect(args []string) error {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.SetOutput(stderr)
	timezone := flags.String("timezone", "UTC", "IANA timezone used to interpret CS2 log timestamps")
	format := flags.String("format", "text", "output format: text, json, or jsonl")
	limit := flags.Int("limit", 0, "maximum matching events to print")
	raw := flags.Bool("raw", false, "include raw log line text")
	useStream := flags.Bool("stream", false, "use stream processor for multiline records")
	includeState := flags.Bool("state", false, "include tracked match state context; implies -stream")
	var typeFilters multiFlag
	flags.Var(&typeFilters, "type", "event type to include; repeatable or comma-separated")
	if err := flags.Parse(args); err != nil {
		return err
	}
	options := inspectOptions{
		format:       *format,
		limit:        *limit,
		raw:          *raw,
		stream:       *useStream || *includeState,
		includeState: *includeState,
		types:        map[string]bool{},
	}
	for _, typ := range typeFilters {
		options.types[typ] = true
	}
	switch options.format {
	case "text", "json", "jsonl":
	default:
		return fmt.Errorf("unsupported inspect format %q", options.format)
	}

	parser, err := cs2log.NewParser(cs2log.Config{LogTimezone: *timezone})
	if err != nil {
		return fmt.Errorf("new parser: %w", err)
	}

	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{"-"}
	}

	var records []inspectRecord
	total := 0
	for _, path := range paths {
		var count int
		pathOptions := options
		if options.limit > 0 {
			pathOptions.limit = options.limit - total
			if pathOptions.limit <= 0 {
				break
			}
		}
		if options.stream {
			count, err = inspectStreamPath(parser, path, pathOptions, &records)
		} else {
			count, err = inspectLinePath(parser, path, pathOptions, &records)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		total += count
		if options.limit > 0 && total >= options.limit {
			break
		}
	}
	if options.format == "json" {
		return json.NewEncoder(stdout).Encode(records)
	}
	return nil
}

func inspectLinePath(parser *cs2log.Parser, path string, options inspectOptions, records *[]inspectRecord) (int, error) {
	var reader io.Reader
	if path == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(path)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		reader = file
	}

	var count int
	lineNumber := 0
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		lineNumber++
		event, err := parser.ParseLine(scanner.Text())
		if err != nil {
			continue
		}
		record := inspectRecord{
			File:      path,
			LineStart: lineNumber,
			LineEnd:   lineNumber,
			Type:      event.EventType(),
			Event:     event,
		}
		if options.raw {
			record.Raw = event.RawLine()
		}
		if !writeInspectRecord(stdout, record, options, records) {
			continue
		}
		count++
		if options.limit > 0 && count >= options.limit {
			break
		}
	}
	return count, scanner.Err()
}

func inspectStreamPath(parser *cs2log.Parser, path string, options inspectOptions, records *[]inspectRecord) (int, error) {
	var reader io.Reader
	if path == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(path)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		reader = file
	}

	var count int
	processor := stream.NewProcessor(parser)
	var tracker *matchstate.Tracker
	if options.includeState {
		tracker = matchstate.NewTracker()
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		for _, record := range processor.PushLine(scanner.Text()) {
			if writeInspectStreamRecord(stdout, path, record, tracker, options, records) {
				count++
				if options.limit > 0 && count >= options.limit {
					return count, scanner.Err()
				}
			}
		}
	}
	for _, record := range processor.Flush() {
		if writeInspectStreamRecord(stdout, path, record, tracker, options, records) {
			count++
			if options.limit > 0 && count >= options.limit {
				break
			}
		}
	}
	return count, scanner.Err()
}

func writeInspectStreamRecord(w io.Writer, path string, record stream.Record, tracker *matchstate.Tracker, options inspectOptions, records *[]inspectRecord) bool {
	var context *matchstate.Context
	if tracker != nil {
		enriched := tracker.Push(record)
		context = &enriched.Context
	}
	if record.Event == nil {
		return false
	}
	inspect := inspectRecord{
		File:      path,
		LineStart: record.LineStart,
		LineEnd:   record.LineEnd,
		Type:      record.Event.EventType(),
		Event:     record.Event,
		Context:   context,
	}
	if options.raw {
		inspect.Raw = record.Raw
	}
	if record.Err != nil {
		inspect.Error = record.Err.Error()
	}
	return writeInspectRecord(w, inspect, options, records)
}

func writeInspectRecord(w io.Writer, record inspectRecord, options inspectOptions, records *[]inspectRecord) bool {
	if len(options.types) > 0 && !options.types[record.Type] {
		return false
	}
	switch options.format {
	case "json":
		*records = append(*records, record)
	case "jsonl":
		if err := json.NewEncoder(w).Encode(record); err != nil {
			fmt.Fprintf(stderr, "encode inspect record: %v\n", err)
		}
	default:
		if record.LineStart == record.LineEnd {
			fmt.Fprintf(w, "%s:%d %s", record.File, record.LineStart, record.Type)
		} else {
			fmt.Fprintf(w, "%s:%d-%d %s", record.File, record.LineStart, record.LineEnd, record.Type)
		}
		if record.Error != "" {
			fmt.Fprintf(w, " error=%q", record.Error)
		}
		if record.Context != nil {
			fmt.Fprint(w, inspectContextText(*record.Context))
		}
		data, err := json.Marshal(record.Event)
		if err != nil {
			fmt.Fprintf(w, " %#v\n", record.Event)
			return true
		}
		fmt.Fprintf(w, " %s\n", data)
	}
	return true
}

func inspectContextText(context matchstate.Context) string {
	var parts []string
	if context.Map != "" {
		parts = append(parts, fmt.Sprintf("map=%s", context.Map))
	}
	if context.Phase != "" {
		parts = append(parts, fmt.Sprintf("phase=%s", context.Phase))
	}
	if context.RoundNumber != 0 {
		parts = append(parts, fmt.Sprintf("round=%d", context.RoundNumber))
	}
	if context.RoundsPlayed != 0 {
		parts = append(parts, fmt.Sprintf("roundsPlayed=%d", context.RoundsPlayed))
	}
	if context.RoundWinnerSide != "" {
		parts = append(parts, fmt.Sprintf("winner=%s", context.RoundWinnerSide))
	}
	if context.RoundEndReason != "" {
		parts = append(parts, fmt.Sprintf("reason=%s", context.RoundEndReason))
	}
	parts = append(parts, fmt.Sprintf("score=%d-%d", context.ScoreT, context.ScoreCT))
	if context.Warmup {
		parts = append(parts, "warmup=true")
	}
	if context.RoundLive {
		parts = append(parts, "roundLive=true")
	}
	if context.GameOver {
		parts = append(parts, "gameOver=true")
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}
