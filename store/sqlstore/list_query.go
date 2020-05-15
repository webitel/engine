package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"strings"
)

type Filter map[string][]interface{}

type Entity interface {
	AllowFields() []string
	DefaultFields() []string
	EntityName() string
}

func GetFields(f []string, e Entity) []string {

	if f == nil || len(f) < 1 {
		//TODO add cache
		f = e.DefaultFields()
	}

	res := make([]string, 0, len(f))

	for _, v := range f {
		if containsString(e.AllowFields(), v) {
			res = append(res, pq.QuoteIdentifier(v))
		}
	}

	return res
}

func GetOrderBy(s string) string {
	if s != "" {
		if s[0] == '+' {
			return "order by " + s[1:] + " asc"
		} else if s[0] == '-' {
			return "order by " + s[1:] + " desc"
		} else {
			return "order by " + s
		}
	}

	return "" //TODO
}

//TODO filter
func Build(req *model.ListRequest, where string, e Entity, args map[string]interface{}) string {
	s := GetFields(req.Fields, e)

	args["Offset"] = req.GetOffset()
	args["Limit"] = req.GetLimit()

	query := fmt.Sprintf(`select %s 
	from %s as t
	where %s
	%s
	offset :Offset
	limit :Limit`, strings.Join(s, ", "), pq.QuoteIdentifier(e.EntityName()), where, GetOrderBy(req.Sort))

	return query
}

func (s *SqlSupplier) ListQuery(out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error {
	q := Build(&req, where, e, params)
	_, err := s.GetReplica().Select(out, q, params)
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
