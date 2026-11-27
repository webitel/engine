package sqlstore

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/webitel/engine/model"
	"github.com/webitel/webitel-go-kit/pkg/sqlkit/order"
)

type SortExpr interface{ SortExprs() map[string]string }

type Entity interface {
	AllowFields() []string
	DefaultFields() []string
	EntityName() string
	DefaultOrder() string
}

func sorterFor(e Entity) *order.Sorter {
	entityAllowedFields := e.AllowFields()
	allowed := make(map[string]order.ExprFunc, len(entityAllowedFields))

	for _, f := range entityAllowedFields {
		allowed[f] = func(d order.Direction) order.Clauses {
			return lookupClauses(e.EntityName(), order.Field{Name: f, Dir: d})
		}
	}

	if m, ok := e.(SortExpr); ok {
		for name, expr := range m.SortExprs() {
			allowed[name] = order.Expr(expr)
		}
	}

	return order.New(allowed)
}

func GetFields(f []string, e Entity) []string {
	if len(f) == 0 { //TODO add cache
		f = e.DefaultFields()
	}

	res := make([]string, 0, len(f))
	jsonb := make(map[string][]string)

	for _, v := range f {
		if slices.Contains(e.AllowFields(), v) {
			res = append(res, pq.QuoteIdentifier(v))

			continue
		}

		if i := strings.Index(v, "."); i > 0 {
			jsonb[v[:i]] = append(jsonb[v[:i]], pq.QuoteLiteral(v[i+1:]))
		}
	}

	for k, v := range jsonb {
		if !slices.Contains(e.AllowFields(), k) || len(v) == 0 {
			continue
		}

		res = append(res, fmt.Sprintf(
			`call_center.cc_jsonb_show_fields(%s, array[%s]) as %s`,
			pq.QuoteIdentifier(k),
			strings.Join(v, ","),
			pq.QuoteIdentifier(k),
		))
	}

	return res
}

func QuoteIdentifier(name string) string { return pq.QuoteIdentifier(name) }
func QuoteLiteral(name string) string    { return pq.QuoteLiteral(name) }

func lookupClauses(table string, f order.Field) order.Clauses {
	return order.Clauses{
		{
			Expr: fmt.Sprintf(
				"case when not call_center.cc_is_lookup(%s, %s) then %s end",
				QuoteLiteral(table),
				QuoteLiteral(f.Name),
				QuoteIdentifier(f.Name),
			),
			Dir: f.Dir,
		}, {
			Expr: fmt.Sprintf(
				"case when call_center.cc_is_lookup(%s, %s) then COALESCE(((%s::text)::json->>'name'), ((%s::text)::json->0->>'name')) end",
				QuoteLiteral(table),
				QuoteLiteral(f.Name),
				QuoteIdentifier(f.Name),
				QuoteIdentifier(f.Name),
			),
			Dir: f.Dir,
		},
	}
}

// TODO filter
func Build(req *model.ListRequest, schema string, where string, e Entity, args map[string]any) (string, error) {
	s := GetFields(req.Fields, e)

	orderBy, err := sorterFor(e).SQL([]string{cmp.Or(req.Sort, e.DefaultOrder())})
	if err != nil {
		return "", err
	}

	args["Offset"] = req.GetOffset()
	args["Limit"] = req.GetLimit()

	t := pq.QuoteIdentifier(e.EntityName())

	if schema != "" {
		t = pq.QuoteIdentifier(schema) + "." + t
	}

	query := fmt.Sprintf(`select %s
	from %s as t
	where %s
	order by %s
	offset :Offset
	limit :Limit`, strings.Join(s, ", "), t, where, orderBy)

	return query, nil
}

// fixme schema
func (s *SqlSupplier) ListQuery(ctx context.Context, out any, req model.ListRequest, where string, e Entity, params map[string]any) error {
	q, err := Build(&req, "call_center", where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetReplica().WithContext(ctx).Select(out, q, params); err != nil {
		return err
	}

	return nil
}

// fixme
func (s *SqlSupplier) ListQueryMaster(ctx context.Context, out any, req model.ListRequest, where string, e Entity, params map[string]any) error {
	q, err := Build(&req, "call_center", where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetMaster().WithContext(ctx).Select(out, q, params); err != nil {
		return err
	}

	return nil
}

func (s *SqlSupplier) One(ctx context.Context, out any, where string, e Entity, params map[string]any) error {
	fields := make([]string, 0, len(e.AllowFields()))

	for _, v := range e.AllowFields() {
		fields = append(fields, pq.QuoteIdentifier(v))
	}

	t := pq.QuoteIdentifier(e.EntityName())

	query := fmt.Sprintf(`select %s
	from call_center.%s as t
	where %s
	limit 1`, strings.Join(fields, ", "), t, where)

	err := s.GetReplica().WithContext(ctx).SelectOne(out, query, params)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqlSupplier) ListQueryTimeout(ctx context.Context, out any, req model.ListRequest, where string, e Entity, params map[string]any) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, (time.Second * time.Duration(s.QueryTimeout())))
	defer cancel()

	q, err := Build(&req, "call_center", where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetReplica().WithContext(ctxTimeout).Select(out, q, params); err != nil {
		return err
	}

	return nil
}

// todo
func (s *SqlSupplier) ListQueryFromSchema(ctx context.Context, out any, schema string, req model.ListRequest, where string, e Entity, params map[string]any) error {
	q, err := Build(&req, schema, where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetReplica().WithContext(ctx).Select(out, q, params); err != nil {
		return err
	}

	return nil
}
