package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yungwood/cs2log"
)

type coverageReport struct {
	files      int
	lines      int
	parsed     int
	noMatch    int
	errors     int
	events     map[string]int
	ignored    map[string]int
	unknowns   map[string]int
	errorTexts map[string]int
}

func runCoverage(args []string) error {
	flags := flag.NewFlagSet("coverage", flag.ContinueOnError)
	flags.SetOutput(stderr)
	timezone := flags.String("timezone", "UTC", "IANA timezone used to interpret CS2 log timestamps")
	top := flags.Int("top", 25, "number of unknown payload prefixes to print")
	prefix := flags.Int("prefix", 120, "maximum unknown payload prefix length")
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

	r := coverageReport{
		events:     map[string]int{},
		ignored:    map[string]int{},
		unknowns:   map[string]int{},
		errorTexts: map[string]int{},
	}
	for _, path := range paths {
		if err := scanCoveragePath(parser, path, *prefix, &r); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	printCoverageReport(stdout, r, *top)
	return nil
}

func scanCoveragePath(parser *cs2log.Parser, path string, prefixLen int, r *coverageReport) error {
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
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		r.lines++
		line := scanner.Text()
		event, err := parser.ParseLine(line)
		if err != nil {
			if errors.Is(err, cs2log.ErrNoMatch) {
				r.noMatch++
				continue
			}
			r.errors++
			r.errorTexts[err.Error()]++
			continue
		}

		r.parsed++
		r.events[event.EventType()]++
		if ignored, ok := event.(cs2log.Ignored); ok {
			r.ignored[ignored.Reason]++
		}
		if unknown, ok := event.(cs2log.Unknown); ok {
			r.unknowns[payloadPrefix(unknown.Payload, prefixLen)]++
		}
	}
	return scanner.Err()
}

func printCoverageReport(w io.Writer, r coverageReport, top int) {
	fmt.Fprintf(w, "files: %d\n", r.files)
	fmt.Fprintf(w, "lines: %d\n", r.lines)
	fmt.Fprintf(w, "parsed timestamp lines: %d\n", r.parsed)
	fmt.Fprintf(w, "no timestamp match: %d\n", r.noMatch)
	fmt.Fprintf(w, "parse errors: %d\n", r.errors)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "events:")
	printCounts(w, r.events, 0)

	if len(r.ignored) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "ignored reasons:")
		printCounts(w, r.ignored, top)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "unknown payload prefixes:")
	printCounts(w, r.unknowns, top)

	if len(r.errorTexts) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "parse errors:")
		printCounts(w, r.errorTexts, top)
	}
}
