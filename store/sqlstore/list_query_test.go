package sqlstore

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/webitel-go-kit/pkg/sqlkit/order"
)

func IsSortError(err error) bool {
	var oe *order.Error
	return errors.As(err, &oe)
}

func norm(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func mustEqual(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("\n got: %s\nwant: %s", got, want)
	}
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type testEntity struct {
	defaultOrder string
}

func (testEntity) AllowFields() []string {
	return []string{"id", "name", "options"}
}

func (testEntity) DefaultFields() []string {
	return []string{"id", "name"}
}

func (testEntity) EntityName() string {
	return "cc_queue"
}

func (e testEntity) DefaultOrder() string {
	return e.defaultOrder
}

type testEntityExprs struct {
	testEntity
}

func (testEntityExprs) SortExprs() map[string]string {
	return map[string]string{
		"priority": "t.priority",
		"agents":   "(select count(*) from x where x.q = t.id)",
	}
}

func lookupSQL(dir string) string {
	return `case when not call_center.cc_is_lookup('cc_queue', 'name') then "name" end ` + dir + `, ` +
		`case when call_center.cc_is_lookup('cc_queue', 'name') then ` +
		`COALESCE((("name"::text)::json->>'name'), (("name"::text)::json->0->>'name')) end ` + dir
}

func TestSorterFor_Fields(t *testing.T) {
	sp := sorterFor(testEntity{})

	for _, raw := range []string{"id", "-id", "name", "-name", "options", "-options"} {
		t.Run(raw, func(t *testing.T) {
			_, err := sp.SQL([]string{raw})
			mustNoErr(t, err)
		})
	}
}

func TestSorterFor_SortExprs(t *testing.T) {
	sp := sorterFor(testEntityExprs{})

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"expression asc", "priority", "t.priority ASC"},
		{"expression desc", "-priority", "t.priority DESC"},
		{"subquery asc", "agents", "(select count(*) from x where x.q = t.id) ASC"},
		{"subquery desc", "-agents", "(select count(*) from x where x.q = t.id) DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sp.SQL([]string{tt.raw})
			mustNoErr(t, err)
			mustEqual(t, norm(got), norm(tt.want))
		})
	}
}

func TestSorterFor_InvalidFields(t *testing.T) {
	sp := sorterFor(testEntity{})

	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"unknown field", "password", order.ErrFieldNotAllowed},
		{"prefix only", "-", order.ErrEmptyField},
		{"comma list", "id,name", order.ErrFieldNotAllowed},
		{"injection", "id; drop table x", order.ErrFieldNotAllowed},
		{"expression", "case when 1=1 then id end", order.ErrFieldNotAllowed},
		{"subquery", "(select 1)", order.ErrFieldNotAllowed},
		{"quoted identifier", `"id"`, order.ErrFieldNotAllowed},
		{"different case", "ID", order.ErrFieldNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sp.SQL([]string{tt.raw})

			if got != "" {
				t.Fatalf("expected empty SQL on error, got %q", got)
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is %v", err, tt.want)
			}

			if !IsSortError(err) {
				t.Fatalf("expected sort error, got %v", err)
			}
		})
	}
}

func TestSorterFor_LookupClauses(t *testing.T) {
	sp := sorterFor(testEntity{})

	tests := []struct {
		name string
		raw  string
		dir  string
	}{
		{"asc", "name", "ASC"},
		{"desc", "-name", "DESC"},
		{"bang desc", "!name", "DESC"},
		{"plus asc", "+name", "ASC"},
		{"whitespace", "  -name  ", "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sp.SQL([]string{tt.raw})
			mustNoErr(t, err)
			mustEqual(t, norm(got), norm(lookupSQL(tt.dir)))
		})
	}
}

func TestSorterFor_InvalidSortExprs(t *testing.T) {
	sp := sorterFor(testEntityExprs{})

	// SortExprs are intentionally whitelisted.
	for _, raw := range []string{"unknown", "priority; drop table x", "t.priority", "Priority"} {
		t.Run(raw, func(t *testing.T) {
			_, err := sp.SQL([]string{raw})
			if err == nil {
				t.Fatalf("expected error for %q", raw)
			}
		})
	}
}

func TestGetFields_DefaultFields(t *testing.T) {
	got := GetFields(nil, testEntity{})

	want := []string{`"id"`, `"name"`}

	if !slices.Equal(got, want) {
		t.Fatalf("GetFields() = %#v, want %#v", got, want)
	}
}

func TestGetFields_AllowedFields(t *testing.T) {
	got := GetFields([]string{"name", "id"}, testEntity{})

	want := []string{`"name"`, `"id"`}

	if !slices.Equal(got, want) {
		t.Fatalf("GetFields() = %#v, want %#v", got, want)
	}
}

func TestGetFields_JSONB(t *testing.T) {
	got := GetFields([]string{"options.foo"}, testEntity{})

	want := []string{
		`call_center.cc_jsonb_show_fields("options", array['foo']) as "options"`,
	}

	if !slices.Equal(got, want) {
		t.Fatalf("GetFields() = %#v, want %#v", got, want)
	}
}

func TestGetFields_JSONB_MultipleKeys(t *testing.T) {
	got := GetFields([]string{"options.foo", "options.bar"}, testEntity{})

	if len(got) != 1 {
		t.Fatalf("expected one JSONB expression, got %#v", got)
	}

	want := `call_center.cc_jsonb_show_fields("options", array['foo','bar']) as "options"`

	mustEqual(t, norm(got[0]), norm(want))
}

func TestGetFields_UnknownFields(t *testing.T) {
	got := GetFields([]string{"unknown", "unknown.foo"}, testEntity{})

	if len(got) != 0 {
		t.Fatalf("unknown fields must be ignored, got %#v", got)
	}
}

func TestGetFields_MixedFields(t *testing.T) {
	got := GetFields([]string{"id", "options.foo", "name"}, testEntity{})

	if len(got) != 3 {
		t.Fatalf("expected 3 fields, got %#v", got)
	}

	if !slices.Contains(got, `"id"`) {
		t.Errorf("id field is missing: %#v", got)
	}

	if !slices.Contains(got, `"name"`) {
		t.Errorf("name field is missing: %#v", got)
	}

	if !slices.Contains(got, `call_center.cc_jsonb_show_fields("options", array['foo']) as "options"`) {
		t.Errorf("JSONB field is missing: %#v", got)
	}
}

func TestBuild_GoldenQuery(t *testing.T) {
	req := model.ListRequest{Sort: "-id"}
	args := map[string]any{}

	q, err := Build(&req, "call_center", "true", testEntity{}, args)
	mustNoErr(t, err)

	mustEqual(
		t,
		norm(q),
		`select "id", "name" from "call_center"."cc_queue" as t where true order by case when not call_center.cc_is_lookup('cc_queue', 'id') then "id" end DESC, case when call_center.cc_is_lookup('cc_queue', 'id') then COALESCE((("id"::text)::json->>'name'), (("id"::text)::json->0->>'name')) end DESC offset :Offset limit :Limit`,
	)
}

func TestBuild_InvalidSort(t *testing.T) {
	for _, raw := range []string{"id; drop table x", "unknown", "-", "id,name"} {
		t.Run(raw, func(t *testing.T) {
			args := map[string]any{"Keep": 1}

			q, err := Build(
				&model.ListRequest{Sort: raw},
				"call_center",
				"true",
				testEntity{},
				args,
			)

			if q != "" {
				t.Fatalf("query must be empty on error, got %q", q)
			}

			if !IsSortError(err) {
				t.Fatalf("want sort error, got %v", err)
			}

			if _, ok := args["Offset"]; ok {
				t.Error("Offset must not be set when Build fails")
			}

			if _, ok := args["Limit"]; ok {
				t.Error("Limit must not be set when Build fails")
			}

			if args["Keep"] != 1 {
				t.Error("existing args must be untouched")
			}
		})
	}
}

func TestBuild_InvalidDefaultOrderFails(t *testing.T) {
	_, err := Build(
		&model.ListRequest{},
		"call_center",
		"true",
		testEntity{defaultOrder: "missing"},
		map[string]any{},
	)

	if !errors.Is(err, order.ErrFieldNotAllowed) {
		t.Fatalf("error = %v, want ErrFieldNotAllowed", err)
	}
}

func TestBuild_Args(t *testing.T) {
	req := model.ListRequest{}
	args := map[string]any{"DomainId": int64(7)}

	_, err := Build(&req, "call_center", "true", testEntity{}, args)
	mustNoErr(t, err)

	if args["DomainId"] != int64(7) {
		t.Error("existing args must be preserved")
	}

	if args["Offset"] != req.GetOffset() {
		t.Errorf("Offset = %v, want %v", args["Offset"], req.GetOffset())
	}

	if args["Limit"] != req.GetLimit() {
		t.Errorf("Limit = %v, want %v", args["Limit"], req.GetLimit())
	}
}

func TestBuild_DoesNotMutateRequest(t *testing.T) {
	req := model.ListRequest{Sort: "-name"}

	_, err := Build(&req, "call_center", "true", testEntity{}, map[string]any{})
	mustNoErr(t, err)

	if req.Sort != "-name" {
		t.Fatalf("req.Sort mutated to %q", req.Sort)
	}
}

func TestBuild_Schema(t *testing.T) {
	req := model.ListRequest{}

	q, err := Build(&req, "", "true", testEntity{}, map[string]any{})
	mustNoErr(t, err)

	if !strings.Contains(q, `from "cc_queue" as t`) {
		t.Fatalf("empty schema must give unqualified table:\n%s", q)
	}

	q, err = Build(&req, "other", "true", testEntity{}, map[string]any{})
	mustNoErr(t, err)

	if !strings.Contains(q, `from "other"."cc_queue" as t`) {
		t.Fatalf("schema must be quoted and qualified:\n%s", q)
	}
}

func allEntities() []Entity {
	return []Entity{
		model.AgentCallStatistics{},
		model.Calendar{},
		model.Timezone{},
		model.Call{},
		model.HistoryCall{},
		model.Agent{},
		model.UserStatus{},
		model.AgentState{},
		model.AgentInQueue{},
		model.AgentSkill{},
		model.AuditForm{},
		model.AuditRate{},
		model.Bucket{},
		model.QueueBucket{},
		model.CommunicationType{},
		model.List{},
		model.ListCommunication{},
		model.Member{},
		model.AttemptHistory{},
		model.Attempt{},
		model.PauseCause{},
		model.PresetQuery{},
		model.Queue{},
		model.QueueHook{},
		model.QueueSkill{},
		model.QuickReply{},
		model.OutboundCallResource{},
		model.ResourceDisplay{},
		model.OutboundResourceGroup{},
		model.OutboundResourceInGroup{},
		model.Skill{},
		model.AgentTeam{},
		model.TeamChatTag{},
		model.TeamHook{},
		model.TeamTrigger{},
		model.Trigger{},
		model.TriggerWithDomainID{},
		model.TriggerJob{},
		model.ChatPlan{},
		model.EmailProfile{},
		model.RoutingSchema{},
		model.RoutingOutboundCall{},
		model.OnlineSkills{},
		model.QueueResourceGroup{},
		model.Region{},
		model.SocketSessionView{},
		model.SystemSetting{},
		model.WebHook{},
	}
}

func TestEntities_SortConfigIsValid(t *testing.T) {
	for _, e := range allEntities() {
		t.Run(e.EntityName(), func(t *testing.T) {
			sp := sorterFor(e)

			if _, err := sp.SQL([]string{e.DefaultOrder()}); err != nil {
				t.Errorf("DefaultOrder() %q is rejected: %v", e.DefaultOrder(), err)
			}

			for _, f := range e.AllowFields() {
				for _, prefix := range []string{"", "-"} {
					if _, err := sp.SQL([]string{prefix + f}); err != nil {
						t.Errorf("sort %q is rejected: %v", prefix+f, err)
					}
				}
			}
		})
	}
}
