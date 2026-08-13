package discovery

import (
	"errors"
	"testing"
)

// TestTTLVerdict covers the three shapes a health verdict can take on its way
// to Consul's TTL check.
//
// The case that matters is (false, nil). The health package promises a false
// verdict always carries a non-nil error, but engine is what panics if that
// promise is ever broken — update() used to call err.Error() unconditionally on
// the not-ok branch. This test fails against that code, and it is hermetic: no
// Consul agent, no network.
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
