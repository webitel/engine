package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"github.com/webitel/webitel-go-kit/pkg/sqlkit/order"
)

var schemeVariableColumns = []struct{ name, col string }{
	{model.SchemeVariableFields.Id, "scheme_variable.id"},
	{model.SchemeVariableFields.Name, "scheme_variable.name"},
	{model.SchemeVariableFields.Encrypt, "scheme_variable.encrypt"},
	{model.SchemeVariableFields.Value, "scheme_variable.value"},
}

var schemeVariableSorter = order.New(
	map[string]string{
		model.SchemeVariableFields.Id:      "scheme_variable.id",
		model.SchemeVariableFields.Name:    "scheme_variable.name",
		model.SchemeVariableFields.Encrypt: "scheme_variable.encrypt",
	},
	order.WithDefaults(order.Field{Name: model.SchemeVariableFields.Id, Dir: order.Asc}),
)

var (
	schemeVariablesSelectFieldsMap = map[string]string{
		model.SchemeVariableFields.Id:      "scheme_variable.id",
		model.SchemeVariableFields.Name:    "scheme_variable.name",
		model.SchemeVariableFields.Encrypt: "scheme_variable.encrypt",
		model.SchemeVariableFields.Value:   "scheme_variable.value",
	}
	schemeVariablesFiltersFieldsMap = map[string]string{
		model.SchemeVariableFields.Id:      "scheme_variable.id",
		model.SchemeVariableFields.Name:    "scheme_variable.name",
		model.SchemeVariableFields.Encrypt: "scheme_variable.encrypt",
		model.SchemeVariableFields.Value:   "scheme_variable.value",
	}
)

type SqlSchemeVariablesStore struct {
	SqlStore
}

func NewSqlSqlSchemeVariablesStore(sqlStore SqlStore) store.SchemeVariablesStore {
	us := &SqlSchemeVariablesStore{sqlStore}
	return us
}

func (s *SqlSchemeVariablesStore) Create(ctx context.Context, domainId int64, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	var out *model.SchemeVariable
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with ins as (
		insert into flow.scheme_variable (domain_id, value, name, encrypt)
		values (:DomainId, :Value::text::json, :Name, :Encrypt)
		returning *
	)
	select
		id,
		name,
		encrypt,
		value
	from ins`,
		map[string]interface{}{
			"DomainId": domainId,
			"Value":    variable.Value,
			"Name":     variable.Name,
			"Encrypt":  variable.Encrypt,
		}); err != nil {
		return nil, model.NewInternalError("store.sql_scheme_variable.save.app_error", fmt.Sprintf("name=%v, %v", variable.Name, err.Error()))
	} else {
		return out, nil
	}
}

func (s *SqlSchemeVariablesStore) Search(ctx context.Context, domainId int64, searchOpts *model.ListRequest, filters any) ([]*model.SchemeVariable, model.AppError) {
	var (
		vars  []*model.SchemeVariable
		query string
		args  []any
	)

	base, err := s.GetQueryBaseFromSearchOptions(domainId, searchOpts)
	if err != nil {
		return nil, model.NewBadRequestError("sqlstore.schema_variables.search.sort", err.Error())
	}

	baseQ, appErr := store.ApplyFiltersToBuilderBulk(base, schemeVariablesFiltersFieldsMap, filters)
	if appErr != nil {
		return nil, appErr
	}

	q, ok := baseQ.(squirrel.SelectBuilder)
	if !ok {
		return nil, model.NewInternalError("store.sql_scheme_variable.get.base_type.wrong", "base of query is of wrong type")
	}

	if query, args, err = q.ToSql(); err != nil {
		return nil, model.NewInternalError("store.sql_scheme_variable.search.preparing", err.Error())
	}

	if _, err = s.GetReplica().WithContext(ctx).Select(&vars, query, args...); err != nil {
		return nil, model.NewCustomCodeError("store.sql_scheme_variable.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return vars, nil
}

func (s *SqlSchemeVariablesStore) Get(ctx context.Context, domainId int64, id int32) (*model.SchemeVariable, model.AppError) {
	var (
		variable *model.SchemeVariable
		query    string
		args     []any
	)

	base, appErr := store.ApplyFiltersToBuilderBulk(s.GetQueryBase(domainId, s.getFields()), schemeVariablesFiltersFieldsMap, &model.Filter{
		Column:         model.SchemeVariableFields.Id,
		Value:          id,
		ComparisonType: model.Equal,
	})
	if appErr != nil {
		return nil, appErr
	}

	switch req := base.(type) {
	case squirrel.SelectBuilder:
		query, args, _ = req.ToSql()
	default:
		return nil, model.NewInternalError("store.sql_scheme_variable.get.base_type.wrong", "base of query is of wrong type")
	}

	err := s.GetReplica().WithContext(ctx).SelectOne(&variable, query, args...)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_scheme_variable.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return variable, nil
}

func (s *SqlSchemeVariablesStore) Update(ctx context.Context, domainId int64, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	var out *model.SchemeVariable
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with upd as (
		update flow.scheme_variable
		set name = :Name,
			value = :Value,
			encrypt = :Encrypt
		where domain_id = :DomainId and id = :Id
		returning *
	)
	select
		id,
		name,
		encrypt,
		upd.value
	from upd`,
		map[string]interface{}{
			"DomainId": domainId,
			"Value":    variable.Value,
			"Name":     variable.Name,
			"Encrypt":  variable.Encrypt,
			"Id":       variable.Id,
		}); err != nil {
		return nil, model.NewInternalError("store.sql_scheme_variable.update.app_error", fmt.Sprintf("id=%v, %v", variable.Id, err.Error()))
	} else {
		return out, nil
	}
}

func (s *SqlSchemeVariablesStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {

	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from flow.scheme_variable
where domain_id = :DomainId and id = :Id`,
		map[string]interface{}{
			"DomainId": domainId,
			"Id":       id,
		}); err != nil {
		return model.NewInternalError("store.sql_scheme_variable.delete.app_error", fmt.Sprintf("id=%v, %v", id, err.Error()))
	} else {
		return nil
	}
}

func (c *SqlSchemeVariablesStore) GetQueryBaseFromSearchOptions(domainId int64, opt *model.ListRequest) (squirrel.SelectBuilder, error) {
	if opt == nil {
		return c.GetQueryBase(domainId, c.getFields()), nil
	}

	base := c.GetQueryBase(domainId, c.selectFields(opt.Fields))

	clauses, err := schemeVariableSorter.Resolve(parseSort(opt.Sort))
	if err != nil {
		return base, err
	}

	base = base.OrderBy(clauses.SQL())

	if q := opt.GetQ(); q != nil {
		base = base.Where("name ilike ?", *q)
	}

	return base.Limit(uint64(opt.GetLimit())).Offset(uint64(opt.GetOffset())), nil
}

func (c *SqlSchemeVariablesStore) GetQueryBase(domainId int64, fields []string) squirrel.SelectBuilder {
	base := squirrel.Select(fields...).
		From("flow.scheme_variable").
		Where("domain_id = ?", domainId).
		PlaceholderFormat(squirrel.Dollar)

	return base
}

func (c *SqlSchemeVariablesStore) getFields() []string {
	fields := make([]string, 0, len(schemeVariableColumns))
	for _, f := range schemeVariableColumns {
		fields = append(fields, f.col)
	}

	return fields
}

func (c *SqlSchemeVariablesStore) selectFields(requested []string) []string {
	fields := make([]string, 0, len(requested))

	for _, v := range requested {
		if col, ok := schemeVariablesSelectFieldsMap[v]; ok {
			fields = append(fields, col)
		}
	}

	if len(fields) == 0 {
		return c.getFields()
	}

	return fields
}

func parseSort(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	if dir, col, ok := strings.Cut(s, ":"); ok {
		switch strings.ToLower(strings.TrimSpace(dir)) {
		case "desc":
			return []string{"-" + col}
		case "asc":
			return []string{col}
		}
	}

	return []string{s}
}
