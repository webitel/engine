package sqlstore

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"strings"
)

var (
	schemeVariablesSelectFieldsMap = map[string]string{
		model.SchemeVariableFields.Id:      "scheme_variable.id",
		model.SchemeVariableFields.Name:    "scheme_variable.name",
		model.SchemeVariableFields.Encrypt: "scheme_variable.encrypt",
		model.SchemeVariableFields.Value:   "case when not scheme_variable.encrypt then scheme_variable.value else 'null'::jsonb end as value",
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
		case when not encrypt then value else 'null'::jsonb end as value
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

	base, appErr := store.ApplyFiltersToBuilderBulk(s.GetQueryBaseFromSearchOptions(domainId, searchOpts), schemeVariablesFiltersFieldsMap, filters)
	if appErr != nil {
		return nil, appErr
	}
	switch req := base.(type) {
	case squirrel.SelectBuilder:
		query, args, _ = req.ToSql()
	default:
		return nil, model.NewInternalError("store.sql_scheme_variable.get.base_type.wrong", "base of query is of wrong type")
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&vars, query, args...)

	if err != nil {
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
			value = :Value
		where domain_id = :DomainId and id = :Id
		returning *
	)
	select 
		id,
		name,
		encrypt,
		case when not upd.encrypt then upd.value else 'null'::jsonb end as value
	from upd`,
		map[string]interface{}{
			"DomainId": domainId,
			"Value":    variable.Value,
			"Name":     variable.Name,
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

func (c *SqlSchemeVariablesStore) GetQueryBaseFromSearchOptions(domainId int64, opt *model.ListRequest) squirrel.SelectBuilder {
	var fields []string
	if opt == nil {
		// TODO
		return c.GetQueryBase(domainId, c.getFields())
	}
	for _, v := range opt.Fields {
		if columnName, ok := schemeVariablesSelectFieldsMap[v]; ok {
			fields = append(fields, columnName)
		} else {
			fields = append(fields, v)
		}
	}
	if len(fields) == 0 {
		fields = append(fields,
			c.getFields()...)
	}
	base := c.GetQueryBase(domainId, fields)
	if opt.Sort != "" {
		splitted := strings.Split(opt.Sort, ":")
		if len(splitted) == 2 {
			order := splitted[0]
			column := splitted[1]
			if v, ok := schemeVariablesFiltersFieldsMap[column]; ok {
				base = base.OrderBy(fmt.Sprintf("%s %s", v, order))
			}

		}
	}

	if opt.GetQ() != nil {
		base = base.Where("name ilike ?", *opt.GetQ())
	}

	base = base.Limit(uint64(opt.GetLimit()))

	return base.Offset(uint64(opt.GetOffset()))
}

func (c *SqlSchemeVariablesStore) GetQueryBase(domainId int64, fields []string) squirrel.SelectBuilder {
	base := squirrel.Select(fields...).
		From("flow.scheme_variable").
		Where("domain_id = ?", domainId).
		PlaceholderFormat(squirrel.Dollar)

	return base
}

func (c *SqlSchemeVariablesStore) getFields() []string {
	var fields []string
	for _, value := range schemeVariablesSelectFieldsMap {
		fields = append(fields, value)
	}
	return fields
}
