package main

import (
	"fmt"
	"io"
	"os"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	if len(os.Args) < 2 {
		printUsage(stderr)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "coverage":
		err = runCoverage(os.Args[2:])
	case "inspect":
		err = runInspect(os.Args[2:])
	case "stream":
		err = runStream(os.Args[2:])
	case "-h", "--help", "help":
		printUsage(stdout)
		return
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", os.Args[1])
		printUsage(stderr)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: cs2log <command> [options] [log ...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  coverage  scan logs with the single-line parser and report unknowns")
	fmt.Fprintln(w, "  inspect   print parsed event records")
	fmt.Fprintln(w, "  stream    scan logs with the stream parser and report multiline blocks")
}
