package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlRoutingSchemaStore struct {
	SqlStore
}

func NewSqlRoutingSchemaStore(sqlStore SqlStore) store.RoutingSchemaStore {
	us := &SqlRoutingSchemaStore{sqlStore}
	return us
}

func (s SqlRoutingSchemaStore) Create(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
	var out *model.RoutingSchema
	if err := s.GetMaster().SelectOne(&out, `with s as (
    insert into flow.acr_routing_scheme (domain_id, name, scheme, payload, type, created_at, created_by, updated_at, updated_by, debug, editor, tags)
    values (:DomainId, :Name, :Scheme, :Payload, :Type, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :Debug, :Editor, :Tags)
    returning *
)
select s.id, s.domain_id, s.name, s.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by,
    s.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, s.scheme as schema, s.payload, debug, s.type
from s
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by`,
		map[string]interface{}{
			"DomainId":  scheme.DomainId,
			"Name":      scheme.Name,
			"Scheme":    scheme.Schema,
			"Payload":   scheme.Payload,
			"Type":      scheme.Type,
			"CreatedAt": scheme.CreatedAt,
			"CreatedBy": scheme.CreatedBy.GetSafeId(),
			"UpdatedAt": scheme.UpdatedAt,
			"UpdatedBy": scheme.UpdatedBy.GetSafeId(),
			"Debug":     scheme.Debug,
			"Editor":    scheme.Editor,
			"Tags":      pq.Array(scheme.Tags),
		}); err != nil {
		return nil, model.NewAppError("SqlRoutingSchemaStore.Save", "store.sql_routing_schema.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", scheme.Name, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlRoutingSchemaStore) GetAllPage(domainId int64, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, *model.AppError) {
	var schemes []*model.RoutingSchema

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
		"Types":    pq.Array(search.Type),
		"Editor":   search.Editor,
		"Tags":     pq.Array(search.Tags),
	}

	err := s.ListQueryFromSchema(&schemes, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar  ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:Types::varchar[] isnull or "type" = any(:Types::varchar[]))
				and (not :Editor::bool is true or editor = true)
				and (:Name::text isnull or name = :Name)
				and (:Tags::varchar[] isnull or tags && :Tags::varchar[])
			`,
		model.RoutingSchema{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlRoutingSchemaStore.GetAllPage", "store.sql_routing_schema.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return schemes, nil
	}
}

func (s SqlRoutingSchemaStore) Get(domainId int64, id int64) (*model.RoutingSchema, *model.AppError) {
	var rScheme *model.RoutingSchema
	if err := s.GetReplica().SelectOne(&rScheme, `select s.id,
       s.domain_id,
       s.name,
       s.created_at,
       call_center.cc_get_lookup(c.id, c.name) as created_by,
       s.updated_at,
       call_center.cc_get_lookup(u.id, u.name) as updated_by,
       s.scheme                                as schema,
       s.payload,
       debug,
       editor,
       s.type,
       s.tags
from flow.acr_routing_scheme s
         left join directory.wbt_user c on c.id = s.created_by
         left join directory.wbt_user u on u.id = s.updated_by
where s.id = :Id
  and s.domain_id = :DomainId
order by s.id`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlRoutingSchemaStore.Get", "store.sql_routing_schema.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return rScheme, nil
	}
}

func (s SqlRoutingSchemaStore) Update(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
	err := s.GetMaster().SelectOne(&scheme, `with s as (
    update flow.acr_routing_scheme s
    set name = :Name,
        scheme = :Scheme,
        payload = :Payload,
        type = :Type,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
		description = :Description,
		debug = :Debug,
		editor = :Editor,
		tags = :Tags
    where s.id = :Id and s.domain_id = :Domain
    returning *
)
select s.id, s.domain_id, s.description, s.name, s.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by,
    s.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, s.scheme as schema, s.payload, s.debug, s.type, s.tags
from s
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by`, map[string]interface{}{
		"Name":        scheme.Name,
		"Scheme":      scheme.Schema,
		"Payload":     scheme.Payload,
		"Type":        scheme.Type,
		"UpdatedAt":   scheme.UpdatedAt,
		"UpdatedBy":   scheme.UpdatedBy.GetSafeId(),
		"Id":          scheme.Id,
		"Domain":      scheme.DomainId,
		"Description": scheme.Description,
		"Debug":       scheme.Debug,
		"Editor":      scheme.Editor,
		"Tags":        pq.Array(scheme.Tags),
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingSchemaStore.Update", "store.sql_routing_schema.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", scheme.Id, err.Error()), code)
	}
	return scheme, nil
}

func (s SqlRoutingSchemaStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from flow.acr_routing_scheme c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingSchemaStore.Delete", "store.sql_routing_schema.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

// todo
func (s SqlRoutingSchemaStore) ListTags(domainId int64, search *model.SearchRoutingSchemaTag) ([]*model.RoutingSchemaTag, *model.AppError) {
	var res []*model.RoutingSchemaTag
	if search.Sort == "" {
		search.Sort = "name"
	}
	st, f := orderBy(search.Sort)
	sort := fmt.Sprintf("order by %s %s", QuoteIdentifier(f), st)

	q := `with tags as (
    select tag as name,
       count(*) count
    from flow.acr_routing_scheme s,
         unnest(s.tags) tag
    where s.domain_id = :DomainId
        and (:Q::varchar isnull or tag ilike :Q::varchar)
        and (:Type::varchar[] isnull or s.type = any(:Type::varchar[]))
    group by 1
)
select *
from tags
%s
limit :Limit
offset :Offset`

	_, err := s.GetReplica().Select(&res, fmt.Sprintf(q, sort), map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"Type":     pq.Array(search.Type),
	})

	if err != nil {
		return nil, model.NewAppError("SqlRoutingSchemaStore.ListTags", "store.sql_routing_schema.tag_list.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}
