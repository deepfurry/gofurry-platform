package runtime

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestDependencyLoggerNeverForwardsDriverMaterial(t *testing.T) {
	var output bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "test", "environment", "test")
	DependencyLogger(base, "river").With("dsn", "sentinel-private").WithGroup("group").Error("sentinel-private", "error", errors.New("sentinel-private"))
	if strings.Contains(output.String(), "sentinel-private") {
		t.Fatal("dependency logger leaked")
	}
	if !strings.Contains(output.String(), `"component":"river"`) || !strings.Contains(output.String(), `"service":"test"`) {
		t.Fatal("safe context lost")
	}
}
