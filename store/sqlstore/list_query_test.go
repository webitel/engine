package sqlstore

import (
	"strings"
	"testing"

	"github.com/webitel/engine/model"
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
			name:      "expression in sort is quoted as identifier",
			sort:      "+id,(select pg_sleep(5))",
			sortOrder: []string{"voice"},
			want:      `order by array_position(array['voice']::text[], "id,(select pg_sleep(5))"::text) asc NULLS LAST, "id,(select pg_sleep(5))" asc NULLS LAST`,
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

type testEntity struct{ defaultOrder string }

func (e testEntity) AllowFields() []string   { return []string{"id", "name"} }
func (e testEntity) DefaultFields() []string { return []string{"id", "name"} }
func (e testEntity) EntityName() string      { return "test_view" }
func (e testEntity) DefaultOrder() string    { return e.defaultOrder }

func TestBuildOrder(t *testing.T) {
	tests := []struct {
		name    string
		req     model.ListRequest
		entity  testEntity
		want    string
		notWant string
	}{
		{
			name:    "user sort with expression is never raw",
			req:     model.ListRequest{Sort: "+id,(select pg_sleep(5))"},
			want:    `then "id,(select pg_sleep(5))" end asc`,
			notWant: "order by +id,(select",
		},
		{
			name: "internal raw sort",
			req:  model.ListRequest{Sort: "case when id > 0 then 0 else 1 end asc,name asc", RawSort: true},
			want: "order by case when id > 0 then 0 else 1 end asc,name asc",
		},
		{
			name:   "raw default order",
			entity: testEntity{defaultOrder: "is_system desc, name asc"},
			want:   "order by is_system desc, name asc",
		},
		{
			name:   "plain default order",
			entity: testEntity{defaultOrder: "+name"},
			want:   `then "name" end asc`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Build(&tt.req, "call_center", "true", tt.entity, map[string]interface{}{})
			if !strings.Contains(got, tt.want) {
				t.Errorf("Build() = %q, want it to contain %q", got, tt.want)
			}

			if tt.notWant != "" && strings.Contains(got, tt.notWant) {
				t.Errorf("Build() = %q, must not contain %q", got, tt.notWant)
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
