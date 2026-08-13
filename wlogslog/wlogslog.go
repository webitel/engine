// Package wlogslog bridges log/slog to wlog, for libraries that take a
// *slog.Logger. It has no engine dependencies.
package wlogslog

import (
	"context"
	"log/slog"

	"github.com/webitel/wlog"
)

type handler struct {
	log    *wlog.Logger
	fields []wlog.Field
}

// NewHandler returns a slog.Handler that writes through to log.
func NewHandler(log *wlog.Logger) slog.Handler {
	return &handler{log: log}
}

// wlog exposes no level query, so filtering is left to wlog.
func (h *handler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *handler) Handle(_ context.Context, rec slog.Record) error {
	fields := make([]wlog.Field, 0, len(h.fields)+rec.NumAttrs())
	fields = append(fields, h.fields...)
	rec.Attrs(func(a slog.Attr) bool {
		fields = append(fields, wlog.Any(a.Key, a.Value.Any()))
		return true
	})

	switch {
	case rec.Level >= slog.LevelError:
		h.log.Error(rec.Message, fields...)
	case rec.Level >= slog.LevelWarn:
		h.log.Warn(rec.Message, fields...)
	case rec.Level >= slog.LevelInfo:
		h.log.Info(rec.Message, fields...)
	default:
		h.log.Debug(rec.Message, fields...)
	}

	return nil
}

// Copies rather than mutates: slog handlers must be safe to share.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	fields := make([]wlog.Field, len(h.fields), len(h.fields)+len(attrs))
	copy(fields, h.fields)
	for _, a := range attrs {
		fields = append(fields, wlog.Any(a.Key, a.Value.Any()))
	}

	return &handler{log: h.log, fields: fields}
}

// wlog.Namespace is zap.Namespace, which already has slog's group semantics.
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	fields := make([]wlog.Field, len(h.fields), len(h.fields)+1)
	copy(fields, h.fields)

	return &handler{log: h.log, fields: append(fields, wlog.Namespace(name))}
}
