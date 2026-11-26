package sqlstore

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/webitel-go-kit/pkg/sqlkit/order"
)

func IsSortError(err error) bool {
	var oe *order.Error

	return errors.As(err, &oe)
}

func norm(s string) string { return strings.Join(strings.Fields(s), " ") }

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

func lookupSQL(dir string) string {
	return `case when not call_center.cc_is_lookup('cc_queue', 'name') then "name" end ` + dir + `, ` +
		`case when call_center.cc_is_lookup('cc_queue', 'name') then ` +
		`COALESCE((("name"::text)::json->>'name'), (("name"::text)::json->0->>'name')) end ` + dir
}

var testSpec = SortSpec{
	Table:   "cc_queue",
	Columns: []string{"name", "priority"},
	Exprs: map[string]string{
		"agents":   "(select count(*) from x where x.q = t.id)",
		"priority": "t.priority",
	},
	Default: "-name",
}

func TestSortSpec_Expr(t *testing.T) {
	tests := []struct {
		name string
		spec SortSpec
		raw  string
		want string
	}{
		{"column asc by default", testSpec, "name", lookupSQL("ASC")},
		{"column explicit plus", testSpec, "+name", lookupSQL("ASC")},
		{"column desc", testSpec, "-name", lookupSQL("DESC")},
		{"bang is desc", testSpec, "!name", lookupSQL("DESC")},
		{"whitespace is trimmed", testSpec, "  -name  ", lookupSQL("DESC")},
		{"expr keeps direction asc", testSpec, "agents", "(select count(*) from x where x.q = t.id) ASC"},
		{"expr keeps direction desc", testSpec, "-agents", "(select count(*) from x where x.q = t.id) DESC"},
		{"expr overrides column", testSpec, "priority", "t.priority ASC"},
		{"default for empty", testSpec, "", lookupSQL("DESC")},
		{"default for blank", testSpec, "   ", lookupSQL("DESC")},
		{"nothing to sort by", SortSpec{}, "", ""},
		{"nil Exprs does not panic", SortSpec{Table: "cc_queue", Columns: []string{"name"}}, "name", lookupSQL("ASC")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.spec.Expr(tt.raw)
			mustNoErr(t, err)
			mustEqual(t, norm(got), norm(tt.want))
		})
	}
}

func TestSortSpec_Expr_Errors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"unknown field", "password", order.ErrFieldNotAllowed},
		{"prefix only", "-", order.ErrEmptyField},
		{"comma list is not supported", "name,priority", order.ErrFieldNotAllowed},
		{"injection", "name; drop table x", order.ErrFieldNotAllowed},
		{"raw case expression", "case when 1=1 then name end", order.ErrFieldNotAllowed},
		{"subquery", "(select 1)", order.ErrFieldNotAllowed},
		{"quoted identifier", `"name"`, order.ErrFieldNotAllowed},
		{"column from Exprs only with different case", "NAME", order.ErrFieldNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testSpec.Expr(tt.raw)

			if got != "" {
				t.Fatalf("expected empty SQL on error, got %q", got)
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is %v", err, tt.want)
			}

			if extractCodeFromErr(err) != http.StatusBadRequest {
				t.Fatalf("IsSortError(%v) = false, want true", err)
			}
		})
	}
}

func TestSortSpec_InvalidDefaultFails(t *testing.T) {
	sp := SortSpec{Columns: []string{"name"}, Default: "missing"}

	_, err := sp.Expr("")
	if !errors.Is(err, order.ErrFieldNotAllowed) {
		t.Fatalf("error = %v, want ErrFieldNotAllowed", err)
	}
}

func TestSortSpec_Clause(t *testing.T) {
	got, err := testSpec.Clause("-agents")
	mustNoErr(t, err)
	mustEqual(t, got, "order by (select count(*) from x where x.q = t.id) DESC")

	got, err = SortSpec{}.Clause("")
	mustNoErr(t, err)
	mustEqual(t, got, "")
}

func TestIsSortError(t *testing.T) {
	if IsSortError(nil) {
		t.Fatal("nil is not a sort error")
	}

	if IsSortError(errors.New("db is down")) {
		t.Fatal("foreign error must not be a sort error")
	}

	wrapped := errors.Join(errors.New("ctx"), &order.Error{Field: "x", Err: order.ErrFieldNotAllowed})
	if !IsSortError(wrapped) {
		t.Fatal("wrapped order.Error must be detected")
	}
}

func FuzzSortSpec_OutputIsWhitelisted(f *testing.F) {
	for _, s := range []string{"", "name", "-name", "!agents", "+priority", "a,b", "name;--", "-", "--name", "\x00", "name\n"} {
		f.Add(s)
	}

	allowed := map[string]struct{}{"": {}}

	for _, dir := range []string{"ASC", "DESC"} {
		allowed[norm(lookupSQL(dir))] = struct{}{}

		for _, e := range testSpec.Exprs {
			allowed[e+" "+dir] = struct{}{}
		}
	}

	f.Fuzz(func(t *testing.T, raw string) {
		out, err := testSpec.Expr(raw)
		if err != nil {
			return
		}

		if _, ok := allowed[norm(out)]; !ok {
			t.Fatalf("input %q produced non-whitelisted SQL: %s", raw, out)
		}
	})
}

type testEntity struct{ defaultOrder string }

func (testEntity) AllowFields() []string   { return []string{"id", "name"} }
func (testEntity) DefaultFields() []string { return []string{"id", "name"} }
func (testEntity) EntityName() string      { return "cc_queue" }
func (e testEntity) DefaultOrder() string  { return e.defaultOrder }

type testEntityExprs struct{ testEntity }

func (testEntityExprs) SortExprs() map[string]string { return map[string]string{"id": "t.id"} }

func TestSpecFor(t *testing.T) {
	t.Run("plain entity", func(t *testing.T) {
		sp := specFor(testEntity{defaultOrder: "name"})

		if sp.Table != "cc_queue" || sp.Default != "name" || len(sp.Columns) != 2 {
			t.Fatalf("unexpected spec: %+v", sp)
		}

		if sp.Exprs != nil {
			t.Fatalf("Exprs must be nil, got %v", sp.Exprs)
		}
	})

	t.Run("entity with SortExprs", func(t *testing.T) {
		got, err := specFor(testEntityExprs{}).Expr("-id")
		mustNoErr(t, err)
		mustEqual(t, got, "t.id DESC")
	})
}

func buildQuery(t *testing.T, e Entity, req model.ListRequest, args map[string]any) (string, error) {
	t.Helper()

	return Build(&req, "call_center", "true", e, args)
}

func TestBuild_GoldenQuery(t *testing.T) {
	q, err := buildQuery(t, testEntityExprs{testEntity{defaultOrder: "-id"}}, model.ListRequest{}, map[string]any{})
	mustNoErr(t, err)

	mustEqual(t, norm(q),
		`select "id", "name" from "call_center"."cc_queue" as t where true `+
			`order by t.id DESC offset :Offset limit :Limit`)
}

func TestBuild_Order(t *testing.T) {
	tests := []struct {
		name     string
		entity   Entity
		sort     string
		contains []string
		absent   []string
	}{
		{
			name:     "client sort wins over default",
			entity:   testEntity{defaultOrder: "-id"},
			sort:     "name",
			contains: []string{"order by case when not", `then "name" end ASC`},
		},
		{
			name:     "default used when sort is empty",
			entity:   testEntityExprs{testEntity{defaultOrder: "-id"}},
			sort:     "",
			contains: []string{"order by t.id DESC"},
		},
		{
			name:   "no order by when nothing to sort",
			entity: testEntity{},
			sort:   "",
			absent: []string{"order by"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := buildQuery(t, tt.entity, model.ListRequest{Sort: tt.sort}, map[string]any{})
			mustNoErr(t, err)

			for _, s := range tt.contains {
				if !strings.Contains(q, s) {
					t.Errorf("query must contain %q:\n%s", s, q)
				}
			}

			for _, s := range tt.absent {
				if strings.Contains(q, s) {
					t.Errorf("query must not contain %q:\n%s", s, q)
				}
			}
		})
	}
}

func TestBuild_InvalidSort(t *testing.T) {
	for _, raw := range []string{"id; drop table x", "unknown", "-", "id,name"} {
		t.Run(raw, func(t *testing.T) {
			args := map[string]any{"Keep": 1}

			q, err := buildQuery(t, testEntity{}, model.ListRequest{Sort: raw}, args)
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
	_, err := buildQuery(t, testEntity{defaultOrder: "missing"}, model.ListRequest{}, map[string]any{})
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
		model.AgentCallStatistics{},
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
			sp := specFor(e)

			if _, err := sp.Expr(""); err != nil {
				t.Errorf("DefaultOrder() %q is rejected: %v", e.DefaultOrder(), err)
			}

			for _, f := range e.AllowFields() {
				for _, prefix := range []string{"", "-"} {
					if _, err := sp.Expr(prefix + f); err != nil {
						t.Errorf("sort %q is rejected: %v", prefix+f, err)
					}
				}
			}
		})
	}
}
