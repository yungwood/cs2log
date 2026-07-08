package matchstate

import (
	"time"

	"github.com/yungwood/cs2log"
	"github.com/yungwood/cs2log/stream"
)

// Phase is a coarse match lifecycle state inferred from explicit log events.
type Phase string

const (
	// PhaseUnknown means no lifecycle phase is known yet.
	PhaseUnknown Phase = ""
	// PhaseLoading means a map or match is loading or initializing.
	PhaseLoading Phase = "loading"
	// PhaseWarmup means the match is in warmup.
	PhaseWarmup Phase = "warmup"
	// PhaseFreezetime means the round freeze period is active.
	PhaseFreezetime Phase = "freezetime"
	// PhaseLive means a round is live.
	PhaseLive Phase = "live"
	// PhaseRoundEnd means the last observed round has ended.
	PhaseRoundEnd Phase = "round_end"
	// PhaseGameOver means the game-over line has been observed.
	PhaseGameOver Phase = "game_over"
)

// Context is the best-known match state after applying a record. Missing
// fields mean the tracker does not know that value yet. Context is attached to
// records as they are processed; prior records are not retroactively corrected.
type Context struct {
	Map              string      `json:"map,omitempty"`
	Phase            Phase       `json:"phase,omitempty"`
	RoundNumber      int         `json:"roundNumber,omitempty"`
	RoundsPlayed     int         `json:"roundsPlayed,omitempty"`
	RoundWinnerSide  cs2log.Side `json:"roundWinnerSide,omitempty"`
	RoundEndReason   string      `json:"roundEndReason,omitempty"`
	RoundEndNotice   string      `json:"roundEndNotice,omitempty"`
	TeamNameT        string      `json:"teamNameT,omitempty"`
	TeamNameCT       string      `json:"teamNameCT,omitempty"`
	ScoreT           int         `json:"scoreT"`
	ScoreCT          int         `json:"scoreCT"`
	Warmup           bool        `json:"warmup,omitempty"`
	RoundLive        bool        `json:"roundLive,omitempty"`
	Paused           bool        `json:"paused,omitempty"`
	GameOver         bool        `json:"gameOver,omitempty"`
	LastEventAt      *time.Time  `json:"lastEventAt,omitempty"`
	MatchStartedAt   *time.Time  `json:"matchStartedAt,omitempty"`
	GameCommencingAt *time.Time  `json:"gameCommencingAt,omitempty"`
	RoundStartedAt   *time.Time  `json:"roundStartedAt,omitempty"`
	RoundEndedAt     *time.Time  `json:"roundEndedAt,omitempty"`
	GameOverAt       *time.Time  `json:"gameOverAt,omitempty"`
}

// Record is a stream record with the current tracked context attached.
type Record struct {
	stream.Record
	Context Context `json:"context"`
}

// Tracker incrementally tracks match context from parsed stream records.
type Tracker struct {
	context Context
}

// NewTracker creates a match context tracker.
func NewTracker() *Tracker {
	return &Tracker{}
}

// Snapshot returns the current tracked context.
func (t *Tracker) Snapshot() Context {
	return t.context
}

// Push records an event and returns it with the updated context attached.
// Errored records and records without events do not update tracked state.
func (t *Tracker) Push(record stream.Record) Record {
	if record.Event != nil && record.Err == nil {
		t.apply(record.Event)
	}
	return Record{
		Record:  record,
		Context: t.context,
	}
}

func (t *Tracker) apply(event cs2log.Event) {
	timestamp := event.Timestamp()
	t.context.LastEventAt = timePtr(timestamp)

	switch e := event.(type) {
	case cs2log.MapLoadingStarted:
		t.resetMatch()
		t.context.Map = e.Map
		t.context.Phase = PhaseLoading
	case cs2log.WorldMatchStart:
		t.resetMatch()
		t.context.Map = e.Map
		t.context.MatchStartedAt = timePtr(timestamp)
		t.context.Phase = PhaseLoading
	case cs2log.WorldGameCommencing:
		t.context.ScoreT = 0
		t.context.ScoreCT = 0
		t.context.RoundNumber = 0
		t.context.RoundsPlayed = 0
		t.context.Warmup = false
		t.context.RoundLive = false
		t.context.Paused = false
		t.context.GameOver = false
		t.context.GameCommencingAt = timePtr(timestamp)
		t.context.RoundStartedAt = nil
		t.context.RoundEndedAt = nil
		t.context.RoundWinnerSide = ""
		t.context.RoundEndReason = ""
		t.context.RoundEndNotice = ""
		t.context.TeamNameT = ""
		t.context.TeamNameCT = ""
		t.context.GameOverAt = nil
		t.context.Phase = PhaseLoading
	case cs2log.FreezeTimeStart:
		t.context.Warmup = false
		t.context.RoundLive = false
		t.context.Paused = false
		t.context.GameOver = false
		t.context.Phase = PhaseFreezetime
	case cs2log.WorldWarmupStart:
		t.context.Warmup = true
		t.context.RoundLive = false
		t.context.Paused = false
		t.context.GameOver = false
		t.context.Phase = PhaseWarmup
	case cs2log.WorldWarmupEnd:
		t.context.Warmup = false
		t.context.Phase = PhaseLoading
	case cs2log.WorldRoundRestart:
		t.context.RoundLive = false
		t.context.Paused = false
		t.context.GameOver = false
		t.context.Phase = PhaseFreezetime
	case cs2log.WorldRoundStart:
		t.context.Warmup = false
		t.context.RoundLive = true
		t.context.Paused = false
		t.context.GameOver = false
		t.context.RoundStartedAt = timePtr(timestamp)
		t.context.RoundEndedAt = nil
		t.context.RoundWinnerSide = ""
		t.context.RoundEndReason = ""
		t.context.RoundEndNotice = ""
		t.context.Phase = PhaseLive
	case cs2log.WorldRoundEnd:
		t.context.RoundLive = false
		t.context.RoundEndedAt = timePtr(timestamp)
		t.context.Phase = PhaseRoundEnd
	case cs2log.MatchStatusScore:
		t.context.Map = e.Map
		t.context.ScoreT = e.ScoreT
		t.context.ScoreCT = e.ScoreCT
		t.context.RoundsPlayed = e.RoundsPlayed
	case cs2log.MatchPause:
		t.context.Paused = e.Action == "enabled"
	case cs2log.TeamScored:
		switch e.Side {
		case cs2log.SideTerrorist:
			t.context.ScoreT = e.Score
		case cs2log.SideCT:
			t.context.ScoreCT = e.Score
		}
	case cs2log.TeamPlaying:
		switch e.Side {
		case cs2log.SideTerrorist:
			t.context.TeamNameT = e.TeamName
		case cs2log.SideCT:
			t.context.TeamNameCT = e.TeamName
		}
	case cs2log.TeamNotice:
		t.context.ScoreT = e.ScoreT
		t.context.ScoreCT = e.ScoreCT
		t.context.RoundWinnerSide = e.Side
		t.context.RoundEndReason = normalizeRoundEndReason(e.Notice)
		t.context.RoundEndNotice = e.Notice
		t.context.RoundLive = false
		t.context.RoundEndedAt = timePtr(timestamp)
		t.context.Phase = PhaseRoundEnd
	case cs2log.GameOver:
		t.context.Map = e.Map
		t.context.ScoreT = e.ScoreT
		t.context.ScoreCT = e.ScoreCT
		t.context.GameOver = true
		t.context.Warmup = false
		t.context.RoundLive = false
		t.context.Paused = false
		t.context.GameOverAt = timePtr(timestamp)
		t.context.Phase = PhaseGameOver
	case stream.RoundStats:
		t.context.Map = e.Map
		t.context.ScoreT = e.ScoreT
		t.context.ScoreCT = e.ScoreCT
		t.context.RoundNumber = e.RoundNumber
	}
}

func (t *Tracker) resetMatch() {
	t.context.RoundNumber = 0
	t.context.RoundsPlayed = 0
	t.context.ScoreT = 0
	t.context.ScoreCT = 0
	t.context.Warmup = false
	t.context.RoundLive = false
	t.context.Paused = false
	t.context.GameOver = false
	t.context.RoundWinnerSide = ""
	t.context.RoundEndReason = ""
	t.context.RoundEndNotice = ""
	t.context.TeamNameT = ""
	t.context.TeamNameCT = ""
	t.context.MatchStartedAt = nil
	t.context.GameCommencingAt = nil
	t.context.RoundStartedAt = nil
	t.context.RoundEndedAt = nil
	t.context.GameOverAt = nil
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func normalizeRoundEndReason(notice string) string {
	switch notice {
	case "SFUI_Notice_Bomb_Defused":
		return "bomb_defused"
	case "SFUI_Notice_Target_Bombed":
		return "bomb_exploded"
	case "SFUI_Notice_Target_Saved":
		return "target_saved"
	case "SFUI_Notice_CTs_Win":
		return "ct_win"
	case "SFUI_Notice_Terrorists_Win":
		return "terrorist_win"
	case "SFUI_Notice_All_Hostages_Rescued":
		return "all_hostages_rescued"
	case "SFUI_Notice_Hostages_Not_Rescued":
		return "hostages_not_rescued"
	default:
		return ""
	}
}
