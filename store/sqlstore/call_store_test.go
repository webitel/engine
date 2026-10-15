package sqlstore

import (
	"strings"
	"testing"
)

func TestRewriteCallSort(t *testing.T) {
	tests := []struct {
		name    string
		sort    string
		wantRaw bool
		prefix  string
	}{
		{name: "from is rewritten", sort: "+from", wantRaw: true, prefix: "coalesce("},
		{name: "to is rewritten", sort: "-to", wantRaw: true, prefix: "coalesce("},
		{name: "plain field passes through", sort: "+created_at", wantRaw: false, prefix: "+created_at"},
		{name: "expression passes through as not raw", sort: "+id,(select 1)", wantRaw: false, prefix: "+id,(select 1)"},
		{name: "empty", sort: "", wantRaw: false, prefix: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, raw := rewriteCallSort(tt.sort)
			if raw != tt.wantRaw {
				t.Errorf("rewriteCallSort(%q) raw = %v, want %v", tt.sort, raw, tt.wantRaw)
			}

			if !strings.HasPrefix(got, tt.prefix) {
				t.Errorf("rewriteCallSort(%q) = %q, want prefix %q", tt.sort, got, tt.prefix)
			}
		})
	}
}
