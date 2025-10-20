package sqlstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/webitel/engine/model"
)

type Filter map[string][]interface{}

type Entity interface {
	AllowFields() []string
	DefaultFields() []string
	EntityName() string
	DefaultOrder() string
}

func GetFields(f []string, e Entity) []string {

	if f == nil || len(f) < 1 {
		//TODO add cache
		f = e.DefaultFields()
	}

	res := make([]string, 0, len(f))

	jsonb := make(map[string][]string)

	for _, v := range f {
		if containsString(e.AllowFields(), v) {
			res = append(res, pq.QuoteIdentifier(v))
		} else {
			i := strings.Index(v, ".")
			if i > 0 {
				jsonb[v[:i]] = append(jsonb[v[:i]], pq.QuoteLiteral(v[i+1:]))
			}
		}
	}

	for k, v := range jsonb {
		if containsString(e.AllowFields(), k) && len(v) != 0 {
			res = append(res, fmt.Sprintf(`call_center.cc_jsonb_show_fields(%s, array[%s]) as %s`, pq.QuoteIdentifier(k),
				strings.Join(v, ","), pq.QuoteIdentifier(k)))
		}
	}

	return res
}

func QuoteIdentifier(name string) string {
	return pq.QuoteIdentifier(name)
}

func QuoteLiteral(name string) string {
	return pq.QuoteLiteral(name)
}

func isRawOrder(s string) bool {
	ls := strings.ToLower(s)
	// якщо це складний вираз або кілька полів — віддаємо як є
	return strings.Contains(ls, "case ") ||
		strings.Contains(s, "(") ||
		strings.Contains(s, ")") ||
		strings.Contains(ls, "::") ||
		strings.Contains(ls, "json") ||
		strings.Contains(s, ",")
}

func GetOrderBy(t, s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	if isRawOrder(s) {
		return "order by " + s
	}

	sort, field := orderBy(s)

	fld := pq.QuoteIdentifier(field)

	nulls := "NULLS LAST"
	if strings.EqualFold(sort, "desc") {
		nulls = "NULLS FIRST"
	}

	return fmt.Sprintf(
		`order by 
           case when not call_center.cc_is_lookup(%s, %s) then %s end %s %s,
           case when     call_center.cc_is_lookup(%s, %s) then (%s::text)::json->>'name' end %s %s`,
		QuoteLiteral(t), QuoteLiteral(field), fld, sort, nulls,
		QuoteLiteral(t), QuoteLiteral(field), fld, sort, nulls,
	)
}

func orderBy(s string) (sort string, field string) {
	if len(s) == 0 {
		return "", ""
	}

	if s[0] == '+' || s[0] == 32 {
		sort = "asc"
		field = s[1:]
	} else if s[0] == '-' {
		sort = "desc"
		field = s[1:]
	} else {
		field = s
	}

	return
}

// TODO filter
func Build(req *model.ListRequest, schema string, where string, e Entity, args map[string]interface{}) string {
	s := GetFields(req.Fields, e)
	sort := ""

	if req.Sort != "" {
		sort = req.Sort
	} else if e.DefaultOrder() != "" {
		sort = e.DefaultOrder()
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
	%s
	offset :Offset
	limit :Limit`, strings.Join(s, ", "), t, where, GetOrderBy(e.EntityName(), sort))

	return query
}

// fixme schema
func (s *SqlSupplier) ListQuery(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q := Build(&req, "call_center", where, e, params)
	_, err := s.GetReplica().WithContext(ctx).Select(out, q, params)
	if err != nil {
		return err
	}

	return nil
}

// fixme
func (s *SqlSupplier) ListQueryMaster(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q := Build(&req, "call_center", where, e, params)
	_, err := s.GetMaster().WithContext(ctx).Select(out, q, params)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqlSupplier) One(ctx context.Context, out interface{}, where string, e Entity, params map[string]interface{}) error {
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

func (s *SqlSupplier) ListQueryTimeout(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	ctxTimeout, _ := context.WithTimeout(ctx, (time.Second * time.Duration(s.QueryTimeout())))
	q := Build(&req, "call_center", where, e, params)
	_, err := s.GetReplica().WithContext(ctxTimeout).Select(out, q, params)
	if err != nil {
		return err
	}

	return nil
}

// todo
func (s *SqlSupplier) ListQueryFromSchema(ctx context.Context, out interface{}, schema string, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q := Build(&req, schema, where, e, params)
	_, err := s.GetReplica().WithContext(ctx).Select(out, q, params)
	if err != nil {
		return err
	}

	return nil
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
