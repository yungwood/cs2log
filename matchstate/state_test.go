package matchstate

import (
	"errors"
	"testing"
	"time"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/stream"
)

func TestTrackerAttachesUpdatedContext(t *testing.T) {
	tracker := NewTracker()
	eventTime := time.Date(2026, 7, 5, 0, 0, 1, 0, time.UTC)

	record := tracker.Push(stream.Record{
		Event: cs2log.MatchStatusScore{
			BaseEvent:    baseAt("MatchStatusScore", eventTime),
			ScoreCT:      4,
			ScoreT:       3,
			Map:          "de_train",
			RoundsPlayed: 7,
		},
		LineStart: 10,
		LineEnd:   10,
	})

	if record.Context.Map != "de_train" || record.Context.ScoreCT != 4 || record.Context.ScoreT != 3 || record.Context.RoundsPlayed != 7 {
		t.Fatalf("context = %#v", record.Context)
	}
	if record.Context.LastEventAt == nil || !record.Context.LastEventAt.Equal(eventTime) {
		t.Fatalf("last event at = %v, want %v", record.Context.LastEventAt, eventTime)
	}
	if tracker.Snapshot() != record.Context {
		t.Fatalf("snapshot = %#v, want %#v", tracker.Snapshot(), record.Context)
	}
}

func TestTrackerUpdatesFromRoundStats(t *testing.T) {
	tracker := NewTracker()

	record := tracker.Push(stream.Record{
		Event: stream.RoundStats{
			BaseEvent:   base("RoundStats"),
			RoundNumber: 8,
			ScoreT:      5,
			ScoreCT:     4,
			Map:         "cs_agency",
		},
		LineStart: 20,
		LineEnd:   32,
	})

	if record.Context.Map != "cs_agency" || record.Context.RoundNumber != 8 || record.Context.ScoreT != 5 || record.Context.ScoreCT != 4 {
		t.Fatalf("context = %#v", record.Context)
	}
}

func TestTrackerUpdatesScoresFromTeamScored(t *testing.T) {
	tracker := NewTracker()

	tracker.Push(stream.Record{Event: cs2log.TeamScored{
		BaseEvent: base("TeamScored"),
		Side:      cs2log.SideTerrorist,
		Score:     4,
	}})
	record := tracker.Push(stream.Record{Event: cs2log.TeamScored{
		BaseEvent: base("TeamScored"),
		Side:      cs2log.SideCT,
		Score:     7,
	}})

	if record.Context.ScoreT != 4 || record.Context.ScoreCT != 7 {
		t.Fatalf("score context = %#v", record.Context)
	}
}

func TestTrackerDoesNotUpdateFromErroredRecord(t *testing.T) {
	tracker := NewTracker()
	tracker.Push(stream.Record{
		Event: cs2log.WorldMatchStart{
			BaseEvent: base("WorldMatchStart"),
			Map:       "de_train",
		},
	})

	record := tracker.Push(stream.Record{
		Event: cs2log.GameOver{
			BaseEvent: base("GameOver"),
			Map:       "de_dust2",
			ScoreCT:   13,
			ScoreT:    11,
		},
		Err: errors.New("parse failed"),
	})

	if record.Context.Map != "de_train" || record.Context.GameOver {
		t.Fatalf("context = %#v", record.Context)
	}
}

func TestTrackerUpdatesPausedState(t *testing.T) {
	tracker := NewTracker()

	paused := tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "enabled",
	}})
	if !paused.Context.Paused {
		t.Fatalf("paused context = %#v", paused.Context)
	}

	unpaused := tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "unpaused",
	}})
	if unpaused.Context.Paused {
		t.Fatalf("unpaused context = %#v", unpaused.Context)
	}

	paused = tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "enabled",
	}})
	if !paused.Context.Paused {
		t.Fatalf("paused context = %#v", paused.Context)
	}

	disabled := tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "disabled",
	}})
	if disabled.Context.Paused {
		t.Fatalf("disabled context = %#v", disabled.Context)
	}
}

func TestTrackerRoundAndWarmupState(t *testing.T) {
	tracker := NewTracker()

	warmupAt := time.Date(2026, 7, 5, 0, 0, 1, 0, time.UTC)
	warmup := tracker.Push(stream.Record{Event: cs2log.WorldWarmupStart{BaseEvent: baseAt("WorldWarmupStart", warmupAt)}})
	if warmup.Context.Phase != PhaseWarmup || !warmup.Context.Warmup || warmup.Context.RoundLive {
		t.Fatalf("warmup context = %#v", warmup.Context)
	}

	freezeAt := time.Date(2026, 7, 5, 0, 0, 2, 0, time.UTC)
	freeze := tracker.Push(stream.Record{Event: cs2log.FreezeTimeStart{BaseEvent: baseAt("FreezeTimeStart", freezeAt)}})
	if freeze.Context.Phase != PhaseFreezetime || freeze.Context.RoundLive || freeze.Context.Warmup {
		t.Fatalf("freezetime context = %#v", freeze.Context)
	}

	roundStartAt := time.Date(2026, 7, 5, 0, 0, 3, 0, time.UTC)
	roundStart := tracker.Push(stream.Record{Event: cs2log.WorldRoundStart{BaseEvent: baseAt("WorldRoundStart", roundStartAt)}})
	if roundStart.Context.Phase != PhaseLive || !roundStart.Context.RoundLive {
		t.Fatalf("round start context = %#v", roundStart.Context)
	}
	if roundStart.Context.RoundStartedAt == nil || !roundStart.Context.RoundStartedAt.Equal(roundStartAt) {
		t.Fatalf("round started at = %v, want %v", roundStart.Context.RoundStartedAt, roundStartAt)
	}

	roundEndAt := time.Date(2026, 7, 5, 0, 0, 4, 0, time.UTC)
	roundEnd := tracker.Push(stream.Record{Event: cs2log.WorldRoundEnd{BaseEvent: baseAt("WorldRoundEnd", roundEndAt)}})
	if roundEnd.Context.Phase != PhaseRoundEnd || roundEnd.Context.RoundLive {
		t.Fatalf("round end context = %#v", roundEnd.Context)
	}
	if roundEnd.Context.RoundEndedAt == nil || !roundEnd.Context.RoundEndedAt.Equal(roundEndAt) {
		t.Fatalf("round ended at = %v, want %v", roundEnd.Context.RoundEndedAt, roundEndAt)
	}
}

func TestTrackerMatchLifecycleTimestamps(t *testing.T) {
	tracker := NewTracker()

	matchAt := time.Date(2026, 7, 5, 0, 0, 1, 0, time.UTC)
	match := tracker.Push(stream.Record{Event: cs2log.WorldMatchStart{BaseEvent: baseAt("WorldMatchStart", matchAt), Map: "de_train"}})
	if match.Context.Phase != PhaseLoading || match.Context.MatchStartedAt == nil || !match.Context.MatchStartedAt.Equal(matchAt) {
		t.Fatalf("match context = %#v", match.Context)
	}

	commencingAt := time.Date(2026, 7, 5, 0, 0, 2, 0, time.UTC)
	commencing := tracker.Push(stream.Record{Event: cs2log.WorldGameCommencing{BaseEvent: baseAt("WorldGameCommencing", commencingAt)}})
	if commencing.Context.GameCommencingAt == nil || !commencing.Context.GameCommencingAt.Equal(commencingAt) {
		t.Fatalf("commencing context = %#v", commencing.Context)
	}

	gameOverAt := time.Date(2026, 7, 5, 0, 0, 3, 0, time.UTC)
	gameOver := tracker.Push(stream.Record{Event: cs2log.GameOver{
		BaseEvent: baseAt("GameOver", gameOverAt),
		Map:       "de_train",
		ScoreCT:   13,
		ScoreT:    11,
	}})
	if gameOver.Context.Phase != PhaseGameOver || !gameOver.Context.GameOver || gameOver.Context.GameOverAt == nil || !gameOver.Context.GameOverAt.Equal(gameOverAt) {
		t.Fatalf("game over context = %#v", gameOver.Context)
	}
}

func TestTrackerRoundEndReasonFromTeamNotice(t *testing.T) {
	tracker := NewTracker()
	roundEndAt := time.Date(2026, 7, 5, 0, 0, 1, 0, time.UTC)

	record := tracker.Push(stream.Record{Event: cs2log.TeamNotice{
		BaseEvent: baseAt("TeamNotice", roundEndAt),
		Side:      cs2log.SideCT,
		Notice:    "SFUI_Notice_Bomb_Defused",
		ScoreCT:   1,
		ScoreT:    0,
	}})

	if record.Context.RoundWinnerSide != cs2log.SideCT || record.Context.RoundEndReason != "bomb_defused" || record.Context.RoundEndNotice != "SFUI_Notice_Bomb_Defused" {
		t.Fatalf("round end context = %#v", record.Context)
	}
	if record.Context.Phase != PhaseRoundEnd || record.Context.RoundLive || record.Context.RoundEndedAt == nil || !record.Context.RoundEndedAt.Equal(roundEndAt) {
		t.Fatalf("round end lifecycle = %#v", record.Context)
	}
}

func TestTrackerUpdatesTeamNames(t *testing.T) {
	tracker := NewTracker()

	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideTerrorist,
		TeamName:  "Terrorists",
	}})
	record := tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideCT,
		TeamName:  "Counter-Terrorists",
	}})

	if record.Context.TeamNameT != "Terrorists" || record.Context.TeamNameCT != "Counter-Terrorists" {
		t.Fatalf("team names = %#v", record.Context)
	}
}

func TestTrackerClearsRoundEndContextOnRoundStart(t *testing.T) {
	tracker := NewTracker()
	tracker.Push(stream.Record{Event: cs2log.TeamNotice{
		BaseEvent: base("TeamNotice"),
		Side:      cs2log.SideTerrorist,
		Notice:    "SFUI_Notice_Target_Bombed",
		ScoreCT:   0,
		ScoreT:    1,
	}})

	record := tracker.Push(stream.Record{Event: cs2log.WorldRoundStart{BaseEvent: base("WorldRoundStart")}})
	if record.Context.RoundWinnerSide != "" || record.Context.RoundEndReason != "" || record.Context.RoundEndNotice != "" || record.Context.RoundEndedAt != nil {
		t.Fatalf("round start context = %#v", record.Context)
	}
}

func TestTrackerClearsPreviousMatchStateOnMapLoading(t *testing.T) {
	tracker := NewTracker()
	tracker.Push(stream.Record{Event: stream.RoundStats{
		BaseEvent:   base("RoundStats"),
		Map:         "de_train",
		RoundNumber: 8,
		ScoreT:      7,
		ScoreCT:     6,
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideTerrorist,
		TeamName:  "Old T",
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideCT,
		TeamName:  "Old CT",
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamNotice{
		BaseEvent: base("TeamNotice"),
		Side:      cs2log.SideTerrorist,
		Notice:    "SFUI_Notice_Target_Bombed",
		ScoreCT:   6,
		ScoreT:    8,
	}})
	tracker.Push(stream.Record{Event: cs2log.GameOver{
		BaseEvent: base("GameOver"),
		Map:       "de_train",
		ScoreCT:   6,
		ScoreT:    8,
	}})
	tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "enabled",
	}})

	record := tracker.Push(stream.Record{Event: cs2log.MapLoadingStarted{
		BaseEvent: base("MapLoadingStarted"),
		Map:       "de_cache",
	}})

	assertFreshMatchContext(t, record.Context, "de_cache")
}

func TestTrackerClearsPreviousMatchStateOnWorldMatchStart(t *testing.T) {
	tracker := NewTracker()
	tracker.Push(stream.Record{Event: stream.RoundStats{
		BaseEvent:   base("RoundStats"),
		Map:         "de_train",
		RoundNumber: 8,
		ScoreT:      7,
		ScoreCT:     6,
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideTerrorist,
		TeamName:  "Old T",
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideCT,
		TeamName:  "Old CT",
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamNotice{
		BaseEvent: base("TeamNotice"),
		Side:      cs2log.SideCT,
		Notice:    "SFUI_Notice_Bomb_Defused",
		ScoreCT:   7,
		ScoreT:    7,
	}})
	tracker.Push(stream.Record{Event: cs2log.GameOver{
		BaseEvent: base("GameOver"),
		Map:       "de_train",
		ScoreCT:   7,
		ScoreT:    7,
	}})
	tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "enabled",
	}})

	matchAt := time.Date(2026, 7, 5, 0, 0, 1, 0, time.UTC)
	record := tracker.Push(stream.Record{Event: cs2log.WorldMatchStart{
		BaseEvent: baseAt("WorldMatchStart", matchAt),
		Map:       "cs_agency",
	}})

	assertFreshMatchContext(t, record.Context, "cs_agency")
	if record.Context.MatchStartedAt == nil || !record.Context.MatchStartedAt.Equal(matchAt) {
		t.Fatalf("match started at = %v, want %v", record.Context.MatchStartedAt, matchAt)
	}
}

func TestTrackerClearsPreviousRoundEndContextOnGameCommencing(t *testing.T) {
	tracker := NewTracker()
	tracker.Push(stream.Record{Event: cs2log.TeamNotice{
		BaseEvent: base("TeamNotice"),
		Side:      cs2log.SideCT,
		Notice:    "SFUI_Notice_CTs_Win",
		ScoreCT:   1,
		ScoreT:    0,
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideTerrorist,
		TeamName:  "Old T",
	}})
	tracker.Push(stream.Record{Event: cs2log.TeamPlaying{
		BaseEvent: base("TeamPlaying"),
		Side:      cs2log.SideCT,
		TeamName:  "Old CT",
	}})
	tracker.Push(stream.Record{Event: cs2log.MatchPause{
		BaseEvent: base("MatchPause"),
		Action:    "enabled",
	}})

	record := tracker.Push(stream.Record{Event: cs2log.WorldGameCommencing{BaseEvent: base("WorldGameCommencing")}})
	if record.Context.RoundWinnerSide != "" || record.Context.RoundEndReason != "" || record.Context.RoundEndNotice != "" || record.Context.RoundEndedAt != nil || record.Context.TeamNameT != "" || record.Context.TeamNameCT != "" || record.Context.Paused {
		t.Fatalf("game commencing context = %#v", record.Context)
	}
}

func TestNormalizeRoundEndReason(t *testing.T) {
	tests := map[string]string{
		"SFUI_Notice_Bomb_Defused":            "bomb_defused",
		"SFUI_Notice_Target_Bombed":           "bomb_exploded",
		"SFUI_Notice_Target_Saved":            "target_saved",
		"SFUI_Notice_CTs_Win":                 "ct_win",
		"SFUI_Notice_Terrorists_Win":          "terrorist_win",
		"SFUI_Notice_All_Hostages_Rescued":    "all_hostages_rescued",
		"SFUI_Notice_Hostages_Not_Rescued":    "hostages_not_rescued",
		"SFUI_Notice_Previously_Unseen_Value": "",
	}
	for notice, want := range tests {
		if got := normalizeRoundEndReason(notice); got != want {
			t.Fatalf("normalizeRoundEndReason(%q) = %q, want %q", notice, got, want)
		}
	}
}

func assertFreshMatchContext(t *testing.T, context Context, wantMap string) {
	t.Helper()
	if context.Map != wantMap || context.Phase != PhaseLoading {
		t.Fatalf("fresh context = %#v, want map %q phase %q", context, wantMap, PhaseLoading)
	}
	if context.RoundNumber != 0 || context.RoundsPlayed != 0 || context.ScoreT != 0 || context.ScoreCT != 0 {
		t.Fatalf("fresh score/round context = %#v", context)
	}
	if context.Warmup || context.RoundLive || context.Paused || context.GameOver {
		t.Fatalf("fresh lifecycle context = %#v", context)
	}
	if context.RoundWinnerSide != "" || context.RoundEndReason != "" || context.RoundEndNotice != "" {
		t.Fatalf("fresh round end context = %#v", context)
	}
	if context.TeamNameT != "" || context.TeamNameCT != "" {
		t.Fatalf("fresh team name context = %#v", context)
	}
	if context.GameCommencingAt != nil || context.RoundStartedAt != nil || context.RoundEndedAt != nil || context.GameOverAt != nil {
		t.Fatalf("fresh timestamp context = %#v", context)
	}
}

func base(eventType string) cs2log.BaseEvent {
	return baseAt(eventType, time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC))
}

func baseAt(eventType string, timestamp time.Time) cs2log.BaseEvent {
	return cs2log.BaseEvent{
		Type:    eventType,
		TimeUTC: timestamp,
	}
}
