package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMultiFlagSetAndString(t *testing.T) {
	var flag multiFlag

	if err := flag.Set("PlayerKill, PlayerSay,, RoundStats "); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	if got := flag.String(); got != "PlayerKill,PlayerSay,RoundStats" {
		t.Fatalf("flag string = %q", got)
	}
}

func TestPayloadPrefix(t *testing.T) {
	if got := payloadPrefix("  hello world  ", 5); got != "hello" {
		t.Fatalf("payload prefix = %q", got)
	}
	if got := payloadPrefix("  hello  ", 0); got != "hello" {
		t.Fatalf("payload prefix without limit = %q", got)
	}
	if got := payloadPrefix("  hi  ", 20); got != "hi" {
		t.Fatalf("payload prefix short = %q", got)
	}
}

func TestPrintCountsSortsAndLimits(t *testing.T) {
	var buffer bytes.Buffer

	printCounts(&buffer, map[string]int{
		"PlayerSay":  2,
		"PlayerKill": 3,
		"RoundStats": 2,
	}, 2)

	got := buffer.String()
	want := "       3  PlayerKill\n       2  PlayerSay\n"
	if got != want {
		t.Fatalf("counts = %q, want %q", got, want)
	}
}

func TestPrintUsage(t *testing.T) {
	var buffer bytes.Buffer

	printUsage(&buffer)

	got := buffer.String()
	for _, want := range []string{"usage: cs2log", "coverage", "inspect", "stream"} {
		if !strings.Contains(got, want) {
			t.Fatalf("usage missing %q: %s", want, got)
		}
	}
}
