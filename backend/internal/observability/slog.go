package observability

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// NewSlogHandler returns a slog.Handler that fans out every record to (a) a
// human-friendly text handler for the local stderr stream and (b) the OTel
// log SDK when Init has been called (so the record ends up in Loki, tagged
// with the active trace_id/span_id). When Init has NOT run, the OTel branch
// is a silent no-op and only stderr receives the record.
func NewSlogHandler(stderr io.Writer, level slog.Level) slog.Handler {
	text := slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level})
	if globalLoggerProvider == nil {
		return text
	}
	otel := otelslog.NewHandler("backend", otelslog.WithLoggerProvider(globalLoggerProvider))
	return multiHandler{handlers: []slog.Handler{text, otel}}
}

// multiHandler fans every record out to a fixed list of slog.Handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (m multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return multiHandler{handlers: next}
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return multiHandler{handlers: next}
}
