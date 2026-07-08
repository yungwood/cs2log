package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCoverageReportsParsedAndUnknownLines(t *testing.T) {
	logPath := writeTempLog(t, []string{
		`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`,
		`L 07/05/2026 - 00:00:01.000: unsupported payload`,
		`not a cs2 log line`,
	})

	out, _ := captureOutput(t, func() error {
		return runCoverage([]string{"-timezone", "UTC", "-top", "5", logPath})
	})

	for _, want := range []string{
		"files: 1",
		"lines: 3",
		"parsed timestamp lines: 2",
		"no timestamp match: 1",
		"       1  PlayerSay",
		"       1  Unknown",
		"unsupported payload",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("coverage output missing %q:\n%s", want, out)
		}
	}
}

func TestRunInspectJSONL(t *testing.T) {
	logPath := writeTempLog(t, []string{
		`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`,
	})

	out, _ := captureOutput(t, func() error {
		return runInspect([]string{"-timezone", "UTC", "-format", "jsonl", logPath})
	})

	var record inspectRecord
	if err := json.Unmarshal([]byte(out), &record); err != nil {
		t.Fatalf("decode jsonl: %v\n%s", err, out)
	}
	if record.Type != "PlayerSay" || record.LineStart != 1 || record.File != logPath {
		t.Fatalf("record = %#v", record)
	}
}

func TestRunInspectJSON(t *testing.T) {
	logPath := writeTempLog(t, []string{
		`L 07/05/2026 - 00:00:00.000: "Player<1><STEAM_1:1:1><CT>" say "gg"`,
	})

	out, _ := captureOutput(t, func() error {
		return runInspect([]string{"-timezone", "UTC", "-format", "json", logPath})
	})

	var records []inspectRecord
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if len(records) != 1 || records[0].Type != "PlayerSay" {
		t.Fatalf("records = %#v", records)
	}
}

func TestRunInspectStreamRoundStatsJSONL(t *testing.T) {
	logPath := writeTempLog(t, roundStatsLogLines())

	out, _ := captureOutput(t, func() error {
		return runInspect([]string{"-stream", "-timezone", "UTC", "-type", "RoundStats", "-format", "jsonl", logPath})
	})

	var record inspectRecord
	if err := json.Unmarshal([]byte(out), &record); err != nil {
		t.Fatalf("decode jsonl: %v\n%s", err, out)
	}
	if record.Type != "RoundStats" || record.LineStart != 1 || record.LineEnd != 10 {
		t.Fatalf("record = %#v", record)
	}
}

func TestRunInspectIncludesMatchState(t *testing.T) {
	logPath := writeTempLog(t, []string{
		`L 07/05/2026 - 00:00:00.000: World triggered "Match_Start" on "de_train"`,
		`L 07/05/2026 - 00:00:01.000: World triggered "Round_Start"`,
	})

	out, _ := captureOutput(t, func() error {
		return runInspect([]string{"-state", "-timezone", "UTC", "-format", "jsonl", "-type", "WorldRoundStart", logPath})
	})

	var record inspectRecord
	if err := json.Unmarshal([]byte(out), &record); err != nil {
		t.Fatalf("decode jsonl: %v\n%s", err, out)
	}
	if record.Context == nil || record.Context.Map != "de_train" || !record.Context.RoundLive {
		t.Fatalf("context = %#v", record.Context)
	}
}

func TestRunInspectStateTextOutput(t *testing.T) {
	logPath := writeTempLog(t, []string{
		`L 07/05/2026 - 00:00:00.000: World triggered "Match_Start" on "de_train"`,
		`L 07/05/2026 - 00:00:01.000: Team "CT" triggered "SFUI_Notice_CTs_Win" (CT "1") (T "0")`,
		`L 07/05/2026 - 00:00:02.000: World triggered "Round_Start"`,
	})

	out, _ := captureOutput(t, func() error {
		return runInspect([]string{"-state", "-timezone", "UTC", "-type", "WorldRoundStart", logPath})
	})

	for _, want := range []string{
		"WorldRoundStart",
		"map=de_train",
		"phase=live",
		"score=0-1",
		"roundLive=true",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("inspect state output missing %q:\n%s", want, out)
		}
	}
}

func TestRunStreamReportsBlocks(t *testing.T) {
	logPath := writeTempLog(t, append(roundStatsLogLines(),
		`L 07/05/2026 - 00:00:09.000: server cvars start`,
		`L 07/05/2026 - 00:00:09.000: "mp_freezetime" = "12"`,
		`L 07/05/2026 - 00:00:09.000: server cvars end`,
	))

	out, _ := captureOutput(t, func() error {
		return runStream([]string{"-timezone", "UTC", "-blocks", logPath})
	})

	for _, want := range []string{
		"RoundStats round=2 score=1-0 map=\"cs_agency\" players=1",
		"ServerCvars status=complete cvars=1",
		"records: 2",
		"round stats events: 1",
		"server cvars events: 1",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stream output missing %q:\n%s", want, out)
		}
	}
}

func TestRunInspectRejectsInvalidFormat(t *testing.T) {
	err := runInspect([]string{"-format", "yaml"})
	if err == nil || !strings.Contains(err.Error(), "unsupported inspect format") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunCommandsReturnFlagParseErrors(t *testing.T) {
	tests := map[string]func([]string) error{
		"coverage": runCoverage,
		"inspect":  runInspect,
		"stream":   runStream,
	}
	for name, run := range tests {
		_, errOut, err := captureOutputAllowError(t, func() error {
			return run([]string{"-missing"})
		})
		if err == nil {
			t.Fatalf("%s returned nil error", name)
		}
		if !strings.Contains(errOut, "flag provided but not defined") {
			t.Fatalf("%s stderr = %q", name, errOut)
		}
	}
}

func TestRunCommandsReturnFileOpenErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.log")
	tests := map[string]func([]string) error{
		"coverage": runCoverage,
		"inspect":  runInspect,
		"stream":   runStream,
	}
	for name, run := range tests {
		err := run([]string{missing})
		if err == nil || !strings.Contains(err.Error(), missing) {
			t.Fatalf("%s err = %v", name, err)
		}
	}
}

func TestRunCoverageRejectsInvalidTimezone(t *testing.T) {
	err := runCoverage([]string{"-timezone", "Not/AZone"})
	if err == nil || !strings.Contains(err.Error(), "new parser") {
		t.Fatalf("err = %v", err)
	}
}

func writeTempLog(t *testing.T, lines []string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "server.log")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	return path
}

func captureOutput(t *testing.T, run func() error) (string, string) {
	t.Helper()
	out, errOut, err := captureOutputAllowError(t, run)
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	return out, errOut
}

func captureOutputAllowError(t *testing.T, run func() error) (string, string, error) {
	t.Helper()

	oldStdout := stdout
	oldStderr := stderr
	var out bytes.Buffer
	var errOut bytes.Buffer
	stdout = &out
	stderr = &errOut
	t.Cleanup(func() {
		stdout = oldStdout
		stderr = oldStderr
	})

	err := run()
	return out.String(), errOut.String(), err
}

func roundStatsLogLines() []string {
	return []string{
		`L 07/05/2026 - 00:00:08.000: JSON_BEGIN{`,
		`L 07/05/2026 - 00:00:08.000: "name": "round_stats",`,
		`L 07/05/2026 - 00:00:08.000: "round_number" : "2",`,
		`L 07/05/2026 - 00:00:08.000: "score_t" : "1",`,
		`L 07/05/2026 - 00:00:08.000: "score_ct" : "0",`,
		`L 07/05/2026 - 00:00:08.000: "map" : "cs_agency",`,
		`L 07/05/2026 - 00:00:08.000: "fields" : "accountid, team"`,
		`L 07/05/2026 - 00:00:08.000: "players": {`,
		`L 07/05/2026 - 00:00:08.000: "player_0": "123456789, 3"`,
		`L 07/05/2026 - 00:00:08.000: }}JSON_END`,
	}
}
