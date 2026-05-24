package observability

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewSlogHandler_WritesToStderrAndDoesntPanicWithoutLP(t *testing.T) {
	var buf bytes.Buffer
	globalLoggerProvider = nil // simulate Init not called

	h := NewSlogHandler(&buf, slog.LevelDebug)
	logger := slog.New(h)
	logger.InfoContext(context.Background(), "hello", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "hello") || !strings.Contains(out, "key=value") {
		t.Errorf("stderr handler did not emit: %q", out)
	}
}
