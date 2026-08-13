package wlogslog

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/webitel/wlog"
)

// newTestLogger returns a wlog logger writing JSON to a temp file, plus a
// reader for whatever it wrote.
func newTestLogger(t *testing.T) (*wlog.Logger, func() string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.log")
	log := wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableFile:   true,
		FileJson:     true,
		FileLevel:    "debug",
		FileLocation: path,
	})

	return log, func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read log: %v", err)
		}
		return string(b)
	}
}

func TestWlogHandlerEnabledAlwaysTrue(t *testing.T) {
	// wlog exposes no level query, so the handler must not filter — otherwise
	// it would silently drop records wlog would have kept.
	log, _ := newTestLogger(t)
	h := NewHandler(log)

	for _, lvl := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if !h.Enabled(context.Background(), lvl) {
			t.Errorf("Enabled(%v) = false, want true", lvl)
		}
	}
}

func TestWlogHandlerWritesThroughToWlog(t *testing.T) {
	log, read := newTestLogger(t)
	slog.New(NewHandler(log)).Info("registry started", "checks", 4)

	out := read()
	if !strings.Contains(out, "registry started") {
		t.Errorf("message missing from wlog output: %s", out)
	}
	if !strings.Contains(out, "checks") {
		t.Errorf("attribute missing from wlog output: %s", out)
	}
}

func TestWlogHandlerLevelMapping(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{"debug", slog.LevelDebug, "DEBUG"},
		{"info", slog.LevelInfo, "INFO"},
		{"warn", slog.LevelWarn, "WARN"},
		{"error", slog.LevelError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, read := newTestLogger(t)
			slog.New(NewHandler(log)).Log(context.Background(), tt.level, "msg-"+tt.name)

			var rec struct {
				Level string `json:"level"`
			}
			line := strings.TrimSpace(read())
			if line == "" {
				t.Fatal("nothing logged")
			}
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("log line is not JSON: %q", line)
			}
			if rec.Level != tt.want {
				t.Errorf("level = %q, want %q", rec.Level, tt.want)
			}
		})
	}
}

// slog's contract says a handler must be safe to share, so WithAttrs must copy.
// Mutating in place would leak attributes between unrelated loggers.
func TestWlogHandlerWithAttrsDoesNotLeak(t *testing.T) {
	log, _ := newTestLogger(t)
	base := NewHandler(log)

	a := base.WithAttrs([]slog.Attr{slog.String("branch", "a")})
	b := base.WithAttrs([]slog.Attr{slog.String("branch", "b")})

	if got := len(base.(*handler).fields); got != 0 {
		t.Errorf("base handler mutated: has %d fields, want 0", got)
	}
	if got := len(a.(*handler).fields); got != 1 {
		t.Errorf("branch a: %d fields, want 1", got)
	}
	if got := len(b.(*handler).fields); got != 1 {
		t.Errorf("branch b: %d fields, want 1", got)
	}

	// A second derivation must not disturb the first.
	a2 := a.WithAttrs([]slog.Attr{slog.String("extra", "x")})
	if got := len(a.(*handler).fields); got != 1 {
		t.Errorf("parent grew to %d fields after deriving a child", got)
	}
	if got := len(a2.(*handler).fields); got != 2 {
		t.Errorf("child: %d fields, want 2", got)
	}
}

func TestWlogHandlerWithGroupAndEmptyCases(t *testing.T) {
	log, _ := newTestLogger(t)
	base := NewHandler(log)

	if got := base.WithAttrs(nil); got != base {
		t.Error("WithAttrs(nil) should return the receiver unchanged")
	}
	if got := base.WithGroup(""); got != base {
		t.Error(`WithGroup("") should return the receiver unchanged`)
	}
	if got := len(base.WithGroup("sub").(*handler).fields); got != 1 {
		t.Errorf("WithGroup: %d fields, want 1", got)
	}
}
