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
	// prefix is the accumulated group path, "" or "a.b." with a trailing dot.
	// Groups are flattened into dotted keys rather than wlog.Namespace: a zap
	// namespace stays open, so a later sibling group would nest inside the
	// earlier one instead of beside it.
	prefix string
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
		fields = appendAttr(fields, a, h.prefix)

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
		fields = appendAttr(fields, a, h.prefix)
	}

	return &handler{log: h.log, fields: fields, prefix: h.prefix}
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	fields := make([]wlog.Field, len(h.fields))
	copy(fields, h.fields)

	return &handler{log: h.log, fields: fields, prefix: h.prefix + name + "."}
}

// appendAttr converts one slog.Attr, following the slog contract: resolve
// LogValuer, drop empty attrs, drop empty groups, and inline a group with an
// empty key rather than qualifying its children.
func appendAttr(fields []wlog.Field, a slog.Attr, prefix string) []wlog.Field {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return fields
	}

	if a.Value.Kind() == slog.KindGroup {
		group := a.Value.Group()
		if len(group) == 0 {
			return fields
		}

		if a.Key != "" {
			prefix += a.Key + "."
		}

		for _, ga := range group {
			fields = appendAttr(fields, ga, prefix)
		}

		return fields
	}

	if a.Key == "" {
		return fields
	}

	return append(fields, wlog.Any(prefix+a.Key, a.Value.Any()))
}
