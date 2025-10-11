package bot

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

// CompanionEventHandler handles events related to companion functionality
type CompanionEventHandler struct {
	supervisor          string
	log                 *slog.Logger
	cfg                 *config.CharacterCfg
	lastLeaderHeartbeat time.Time
	leaderInGame        bool
	currentGameName     string
	exitGameChan        chan struct{}
	mu                  sync.RWMutex // Protects heartbeat fields
}

// NewCompanionEventHandler creates a new instance of CompanionEventHandler
func NewCompanionEventHandler(supervisor string, log *slog.Logger, cfg *config.CharacterCfg) *CompanionEventHandler {
	return &CompanionEventHandler{
		supervisor:          supervisor,
		log:                 log,
		cfg:                 cfg,
		exitGameChan:        make(chan struct{}, 1),
		lastLeaderHeartbeat: time.Time{},
		leaderInGame:        false,
		currentGameName:     "",
	}
}

// GetExitGameChan returns the exit game signal channel
func (h *CompanionEventHandler) GetExitGameChan() <-chan struct{} {
	return h.exitGameChan
}

// IsLeaderInGame returns whether the leader is currently in-game (thread-safe)
func (h *CompanionEventHandler) IsLeaderInGame() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.leaderInGame
}

// GetCurrentGameName returns the current game name from heartbeat (thread-safe)
func (h *CompanionEventHandler) GetCurrentGameName() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentGameName
}

// Handle processes companion-related events
func (h *CompanionEventHandler) Handle(ctx context.Context, e event.Event) error {

	switch evt := e.(type) {

	case event.LeaderGameHeartbeatEvent:
		// Only process if we're a companion and this is our leader
		if h.cfg.Companion.Enabled && !h.cfg.Companion.Leader {
			if h.cfg.Companion.LeaderName == "" || evt.Leader == h.cfg.Companion.LeaderName {
				h.mu.Lock()
				h.lastLeaderHeartbeat = time.Now()
				h.leaderInGame = evt.InGame
				h.currentGameName = evt.GameName
				h.mu.Unlock()

				// If leader exited, signal companion to exit too
				if !evt.InGame && h.currentGameName == h.cfg.Companion.CompanionGameName {
					h.log.Info("Leader exited game, companion will exit too",
						slog.String("supervisor", h.supervisor),
						slog.String("leader", evt.Leader),
						slog.String("game", evt.GameName))

					// Signal to exit game (non-blocking)
					select {
					case h.exitGameChan <- struct{}{}:
					default:
					}
				}
			}
		}

	case event.RequestCompanionJoinGameEvent:

		if h.cfg.Companion.Enabled && !h.cfg.Companion.Leader {

			// Check if the leader matches the one in our config or no leader set
			if h.cfg.Companion.LeaderName == "" || evt.Leader == h.cfg.Companion.LeaderName {
				h.log.Info("Companion join game event received", slog.String("supervisor", h.supervisor), slog.String("leader", evt.Leader), slog.String("name", evt.Name), slog.String("password", evt.Password))
				h.cfg.Companion.CompanionGameName = evt.Name
				h.cfg.Companion.CompanionGamePassword = evt.Password
			}
		}

	case event.ResetCompanionGameInfoEvent:

		// If this character is a companion (not a leader), clear game info
		if h.cfg.Companion.Enabled && !h.cfg.Companion.Leader {

			// Check if the leader matches the one in our config or no leader set.
			// Additional check for if LeaderName is the same as the character name for Manual join triggers
			if h.cfg.Companion.LeaderName == "" || evt.Leader == h.cfg.Companion.LeaderName || h.cfg.CharacterName == evt.Leader {
				h.log.Info("Companion reset game info event received", slog.String("supervisor", h.supervisor), slog.String("leader", evt.Leader))
				h.cfg.Companion.CompanionGameName = ""
				h.cfg.Companion.CompanionGamePassword = ""
			}
		}
	}

	return nil
}

// StartHeartbeatMonitor starts monitoring for leader heartbeat timeouts
func (h *CompanionEventHandler) StartHeartbeatMonitor(ctx context.Context) {
	if !h.cfg.Companion.Enabled || h.cfg.Companion.Leader {
		return // Only companions need to monitor heartbeats
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.mu.RLock()
				lastHeartbeat := h.lastLeaderHeartbeat
				leaderInGame := h.leaderInGame
				h.mu.RUnlock()

				// If we haven't received a heartbeat in 30 seconds and leader was in-game, assume crash
				if !lastHeartbeat.IsZero() && time.Since(lastHeartbeat) > 30*time.Second && leaderInGame {
					h.log.Warn("Leader heartbeat timeout - leader may have crashed",
						slog.String("supervisor", h.supervisor),
						slog.Duration("since_last_heartbeat", time.Since(lastHeartbeat)))

					h.mu.Lock()
					h.leaderInGame = false
					h.mu.Unlock()

					// Signal exit (non-blocking)
					select {
					case h.exitGameChan <- struct{}{}:
					default:
					}
				}
			}
		}
	}()
}
