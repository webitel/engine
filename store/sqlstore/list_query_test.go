package sqlstore

import (
	"strings"
	"testing"
)

func TestGetOrderBySortOrder(t *testing.T) {
	tests := []struct {
		name      string
		sort      string
		sortOrder []string
		want      string
	}{
		{
			name:      "asc",
			sort:      "+type",
			sortOrder: []string{"voice", "default", "chat"},
			want:      `order by array_position(array['voice','default','chat']::text[], "type"::text) asc NULLS LAST, "type" asc NULLS LAST`,
		},
		{
			name:      "desc",
			sort:      "-editor",
			sortOrder: []string{"false", "true"},
			want:      `order by array_position(array['false','true']::text[], "editor"::text) desc NULLS LAST, "editor" desc NULLS FIRST`,
		},
		{
			name:      "quoted value",
			sort:      "+type",
			sortOrder: []string{"a'b"},
			want:      `order by array_position(array['a''b']::text[], "type"::text) asc NULLS LAST, "type" asc NULLS LAST`,
		},
		{
			name:      "raw sort ignores sort order",
			sort:      "type, name",
			sortOrder: []string{"voice"},
			want:      "order by type, name",
		},
		{
			name:      "empty sort",
			sort:      "",
			sortOrder: []string{"voice"},
			want:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOrderBy("acr_routing_scheme_view", tt.sort, tt.sortOrder...); got != tt.want {
				t.Errorf("GetOrderBy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetOrderByWithoutSortOrder(t *testing.T) {
	got := GetOrderBy("acr_routing_scheme_view", "+type")
	if strings.Contains(got, "array_position") || !strings.Contains(got, "call_center.cc_is_lookup('acr_routing_scheme_view', 'type')") {
		t.Errorf("GetOrderBy() without sort order = %q", got)
	}
}
