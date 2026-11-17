package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"slices"
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

type CustomSorter interface {
	CustomSorts() []string
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

	return strings.Contains(ls, "case ") ||
		strings.Contains(s, "(") ||
		strings.Contains(s, ")") ||
		strings.Contains(ls, "::") ||
		strings.Contains(ls, "json") ||
		strings.Contains(s, ",")
}

var ErrInvalidSortField = errors.New("invalid sort field")

func GetOrderBy(t, s string, e ...Entity) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	if isRawOrder(s) {
		if len(e) == 0 || e[0] == nil {
			return "", fmt.Errorf("%w: %s", ErrInvalidSortField, t)
		}

		customSorter, ok := e[0].(CustomSorter)
		if !ok {
			return "", fmt.Errorf("%w custom not allowed", ErrInvalidSortField, e[0])
		}

		ls := strings.ToLower(s)
		if !slices.ContainsFunc(customSorter.CustomSorts(), func(f string) bool { return strings.ToLower(f) == ls }) {
			return "", fmt.Errorf("%w %s", ErrInvalidSortField, ls)
		}

		return "order by " + s, nil
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
		   case when     call_center.cc_is_lookup(%s, %s) then COALESCE(((%s::text)::json->>'name'), ((%s::text)::json->0->>'name')) end %s %s`,
		QuoteLiteral(t), QuoteLiteral(field), fld, sort, nulls,
		QuoteLiteral(t), QuoteLiteral(field), fld, fld, sort, nulls,
	), nil
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
func Build(req *model.ListRequest, schema string, where string, e Entity, args map[string]any) (string, error) {
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

	order, err := GetOrderBy(e.EntityName(), sort, e)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf(`select %s
	from %s as t
	where %s
	%s
	offset :Offset
	limit :Limit`, strings.Join(s, ", "), t, where, order)

	return query, nil
}

// fixme schema
func (s *SqlSupplier) ListQuery(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
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
func (s *SqlSupplier) ListQueryMaster(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q, err := Build(&req, "call_center", where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetMaster().WithContext(ctx).Select(out, q, params); err != nil {
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
	q, err := Build(&req, "call_center", where, e, params)
	if err != nil {
		return err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*time.Duration(s.QueryTimeout()))
	defer cancel()

	if _, err = s.GetReplica().WithContext(ctxTimeout).Select(out, q, params); err != nil {
		return err
	}

	return nil
}

// todo
func (s *SqlSupplier) ListQueryFromSchema(ctx context.Context, out interface{}, schema string, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q, err := Build(&req, schema, where, e, params)
	if err != nil {
		return err
	}

	if _, err = s.GetReplica().WithContext(ctx).Select(out, q, params); err != nil {
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
