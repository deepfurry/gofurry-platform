// Package runtime contains reusable process lifecycle mechanisms.
package runtime

import (
	"context"
	"log/slog"
	"os"
)

func Logger(service, environment string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", service, "environment", environment)
	slog.SetDefault(logger)
	return logger
}

// DependencyLogger reports upstream events without forwarding uncontrolled text
// or attributes that can contain credentials, connection addresses or job payloads.
func DependencyLogger(logger *slog.Logger, component string) *slog.Logger {
	return slog.New(dependencyHandler{logger: logger, component: component})
}

type dependencyHandler struct {
	logger    *slog.Logger
	component string
}

func (h dependencyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.logger.Enabled(ctx, level)
}
func (h dependencyHandler) Handle(ctx context.Context, record slog.Record) error {
	h.logger.Log(ctx, record.Level, "dependency event (details withheld)", "component", h.component)
	return nil
}
func (h dependencyHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h dependencyHandler) WithGroup(_ string) slog.Handler      { return h }
