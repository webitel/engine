package discovery

import (
	"errors"
	"fmt"
	"net/http"
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
