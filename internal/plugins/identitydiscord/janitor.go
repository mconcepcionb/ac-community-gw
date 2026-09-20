package identitydiscord

import (
	"context"
	"log/slog"
	"time"
)

// RunCleanup periodically deletes expired sessions and OAuth states. It returns
// when the context is cancelled.
func RunCleanup(ctx context.Context, interval time.Duration, logger *slog.Logger, cleaner Cleaner) {
	if cleaner == nil {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := cleaner.DeleteExpiredSessions(ctx); err != nil {
				logger.Warn("identity-discord: delete expired sessions", "error", err)
			}
			if err := cleaner.DeleteExpiredOAuthStates(ctx); err != nil {
				logger.Warn("identity-discord: delete expired oauth states", "error", err)
			}
		}
	}
}
