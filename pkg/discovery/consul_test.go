package discovery

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
)

// (false, nil) is the case that matters: update() used to panic on that branch.
func TestTTLVerdict(t *testing.T) {
	tests := []struct {
		name       string
		ok         bool
		err        error
		wantPass   bool
		wantOutput string
	}{
		{
			name:       "ready",
			ok:         true,
			err:        nil,
			wantPass:   true,
			wantOutput: "ready...",
		},
		{
			name:       "not ready with reason",
			ok:         false,
			err:        errors.New("grpc listener not accepting"),
			wantPass:   false,
			wantOutput: "grpc listener not accepting",
		},
		{
			name:       "not ready without an error must not panic",
			ok:         false,
			err:        nil,
			wantPass:   false,
			wantOutput: "not ready",
		},
		{
			name:       "ready wins even if an error is handed along",
			ok:         true,
			err:        errors.New("ignored"),
			wantPass:   true,
			wantOutput: "ready...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, output := ttlVerdict(tt.ok, tt.err)

			if pass != tt.wantPass {
				t.Errorf("pass = %v, want %v", pass, tt.wantPass)
			}

			if output != tt.wantOutput {
				t.Errorf("output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

// Consul answers 404 once DeregisterCriticalServiceAfter has dropped the
// service. Treating that as fatal leaves a recovered node out of discovery.
func TestShouldReregister(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"service was deregistered", api.StatusError{Code: http.StatusNotFound}, true},
		{"agent internal error", api.StatusError{Code: http.StatusInternalServerError}, true},
		{"bad request", api.StatusError{Code: http.StatusBadRequest}, false},
		{"wrapped 404", fmt.Errorf("update ttl: %w", api.StatusError{Code: http.StatusNotFound}), true},
		{"network error", errors.New("connection refused"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReregister(tt.err); got != tt.want {
				t.Errorf("shouldReregister(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// Recovery runs inside the TTL updater, so it must not start another one:
// each extra updater duplicates TTL traffic and races on checkId.
func TestReregisterStartsNoSecondUpdater(t *testing.T) {
	src, err := os.ReadFile("consul.go")
	if err != nil {
		t.Fatal(err)
	}

	body := funcBody(t, string(src), "func (c *consul) handlePassTTLError(")
	if strings.Contains(body, "c.register(") {
		t.Error("handlePassTTLError calls register, which spawns a TTL updater; use putRegistration")
	}

	if !strings.Contains(body, "c.putRegistration(") {
		t.Error("handlePassTTLError no longer re-registers")
	}

	// putRegistration must not call update: update is what routes failures back
	// into handlePassTTLError, which would recurse.
	if put := funcBody(t, string(src), "func (c *consul) putRegistration("); strings.Contains(put, "c.update(") {
		t.Error("putRegistration calls update, which can recurse through handlePassTTLError")
	}

	if n := strings.Count(string(src), "go c.updateTTL("); n != 1 {
		t.Errorf("found %d updateTTL launches, want exactly 1", n)
	}
}

// funcBody returns the source between a function's signature and its closing brace.
func funcBody(t *testing.T, src, signature string) string {
	t.Helper()

	i := strings.Index(src, signature)
	if i < 0 {
		t.Fatalf("not found: %s", signature)
	}

	rest := src[i:]

	end := strings.Index(rest, "\n}")
	if end < 0 {
		t.Fatalf("no closing brace: %s", signature)
	}

	return rest[:end]
}
