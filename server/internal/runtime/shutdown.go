package runtime

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// ShutdownDeadline bounds the entire process cleanup, including pool.Close which
// otherwise waits indefinitely for borrowed connections. Normal cleanup cancels it.
func ShutdownDeadline(signalContext context.Context, logger *slog.Logger) func() {
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-signalContext.Done():
		}
		timer := time.NewTimer(15 * time.Second)
		defer timer.Stop()
		select {
		case <-done:
			return
		case <-timer.C:
			logger.Error("process cleanup exceeded shutdown deadline")
			os.Exit(1)
		}
	}()
	return func() { close(done) }
}
