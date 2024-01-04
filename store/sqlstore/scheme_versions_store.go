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
	schemeVersionsSelectFieldsMap = map[string]string{
		model.SchemeVersionFields.SchemeId:  "scheme_version.id",
		model.SchemeVersionFields.CreatedBy: "call_center.cc_get_lookup(scheme_version.created_by, wbt_user.name::text) as created_by",
		model.SchemeVersionFields.CreatedAt: "scheme_version.created_at",
		model.SchemeVersionFields.SchemeId:  "scheme_version.scheme_id",
		model.SchemeVersionFields.Scheme:    "scheme_version.scheme",
		model.SchemeVersionFields.Payload:   "scheme_version.payload",
		model.SchemeVersionFields.Version:   "scheme_version.version",
		model.SchemeVersionFields.Note:      "scheme_version.note",
	}
	schemeVersionsFiltersFieldsMap = map[string]string{
		model.SchemeVersionFields.SchemeId:  "scheme_version.id",
		model.SchemeVersionFields.CreatedBy: "scheme_version.created_by",
		model.SchemeVersionFields.CreatedAt: "scheme_version.created_at",
		model.SchemeVersionFields.SchemeId:  "scheme_version.scheme_id",
		model.SchemeVersionFields.Scheme:    "scheme_version.scheme",
		model.SchemeVersionFields.Payload:   "scheme_version.payload",
		model.SchemeVersionFields.Version:   "scheme_version.version",
		model.SchemeVersionFields.Note:      "scheme_version.note",
	}
)

type SqlSchemeVersionsStore struct {
	SqlStore
}

func NewSqlSchemeVersionsStore(sqlStore SqlStore) store.SchemeVersionsStore {
	us := &SqlSchemeVersionsStore{sqlStore}
	return us
}

func (s *SqlSchemeVersionsStore) Search(ctx context.Context, searchOpts *model.ListRequest, filters any) ([]*model.SchemeVersion, model.AppError) {
	var (
		versions []*model.SchemeVersion
		query    string
		args     []any
	)

	base := store.ApplyFiltersToBuilderBulk(s.GetQueryBaseFromSearchOptions(searchOpts), schemeVersionsFiltersFieldsMap, filters)
	switch req := base.(type) {
	case squirrel.SelectBuilder:
		query, args, _ = req.ToSql()
	default:
		return nil, model.NewInternalError("store.sql_scheme_version.get.base_type.wrong", "base of query is of wrong type")
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&versions, query, args...)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_scheme_version.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return versions, nil
}

func (c *SqlSchemeVersionsStore) GetQueryBaseFromSearchOptions(opt *model.ListRequest) squirrel.SelectBuilder {
	var fields []string
	if opt == nil {
		return c.GetQueryBase(c.getFields())
	}
	for _, v := range opt.Fields {
		if columnName, ok := schemeVersionsSelectFieldsMap[v]; ok {
			fields = append(fields, columnName)
		} else {
			fields = append(fields, v)
		}
	}
	if len(fields) == 0 {
		fields = append(fields,
			c.getFields()...)
	}
	base := c.GetQueryBase(fields)
	if opt.Q != "" {
		base = base.Where(squirrel.Like{"user_ip": opt.Q + "%"})
	}
	if opt.Sort != "" {
		splitted := strings.Split(opt.Sort, ":")
		if len(splitted) == 2 {
			order := splitted[0]
			column := splitted[1]
			if column == "user" {
				column = "user_name"
			}
			base = base.OrderBy(fmt.Sprintf("%s %s", column, order))
		}

	}
	offset := (opt.Page - 1) * opt.PerPage
	if offset < 0 {
		offset = 0
	}
	if opt.PerPage != 0 {
		base = base.Limit(uint64(opt.PerPage + 1))
	}
	return base.Offset(uint64(offset))
}

func (c *SqlSchemeVersionsStore) GetQueryBase(fields []string) squirrel.SelectBuilder {
	base := squirrel.Select(fields...).
		From("flow.scheme_version").
		JoinClause("LEFT JOIN directory.wbt_user ON wbt_user.id = scheme_version.created_by").
		PlaceholderFormat(squirrel.Dollar)

	return base
}

func (c *SqlSchemeVersionsStore) getFields() []string {
	var fields []string
	for _, value := range schemeVersionsSelectFieldsMap {
		fields = append(fields, value)
	}
	return fields
}
