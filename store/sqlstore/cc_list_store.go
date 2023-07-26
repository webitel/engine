package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlListStore struct {
	SqlStore
}

func NewSqlListStore(sqlStore SqlStore) store.ListStore {
	us := &SqlListStore{sqlStore}
	return us
}

func (s SqlListStore) Create(ctx context.Context, list *model.List) (*model.List, model.AppError) {
	var out *model.List
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with i as (
    insert into call_center.cc_list (name, description, domain_id, created_at, created_by, updated_at, updated_by)
    values (:Name, :Description, :DomainId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy)
    returning *
)
select
       i.id,
       i.name,
       i.description,
       i.domain_id,
       i.created_at,
       call_center.cc_get_lookup(uc.id, uc.name) as created_by,
       i.updated_at,
       call_center.cc_get_lookup(u.id, u.name) as updated_by

from i
    left join directory.wbt_user uc on uc.id = i.created_by
    left join directory.wbt_user u on u.id = i.updated_by`,
		map[string]interface{}{
			"Name":        list.Name,
			"Description": list.Description,
			"DomainId":    list.DomainId,
			"CreatedAt":   list.CreatedAt,
			"CreatedBy":   list.CreatedBy.GetSafeId(),
			"UpdatedAt":   list.UpdatedAt,
			"UpdatedBy":   list.UpdatedBy.GetSafeId(),
		}); err != nil {
		return nil, model.NewInternalError("store.sql_list.save.app_error", fmt.Sprintf("name=%v, %v", list.Name, err.Error()))
	} else {
		return out, nil
	}
}

func (s SqlListStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_list_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlListStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchList) ([]*model.List, model.AppError) {
	var list []*model.List

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.List{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_list.get_all.app_error", err.Error())
	} else {
		return list, nil
	}
}

func (s SqlListStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchList) ([]*model.List, model.AppError) {
	var list []*model.List

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))
			    and exists(select 1
				  from call_center.cc_list_acl a
 				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access
				)`,
		model.List{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_list.get_all.app_error", err.Error())
	} else {
		return list, nil
	}
}

func (s SqlListStore) Get(ctx context.Context, domainId int64, id int64) (*model.List, model.AppError) {
	var list *model.List
	if err := s.GetReplica().WithContext(ctx).SelectOne(&list, `
			select
			   i.id,
			   i.name,
			   i.description,
			   i.domain_id,
			   i.created_at,
			   call_center.cc_get_lookup(uc.id, uc.name) as created_by,
			   i.updated_at,
			   call_center.cc_get_lookup(u.id, u.name) as updated_by,
			   coalesce(cls.count, 0) count
		from call_center.cc_list i
			left join directory.wbt_user uc on uc.id = i.created_by
			left join directory.wbt_user u on u.id = i.updated_by
			left join call_center.cc_list_statistics cls on i.id = cls.list_id
		where i.domain_id = :DomainId and i.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_list.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return list, nil
	}
}

func (s SqlListStore) Update(ctx context.Context, list *model.List) (*model.List, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&list, `with i as (
    update call_center.cc_list
        set name = :Name,
            description = :Description,
            updated_at = :UpdatedAt,
            updated_by = :UpdatedBy
    where id = :Id and domain_id = :DomainId
    returning *
)
select
       i.id,
       i.name,
       i.description,
       i.domain_id,
       i.created_at,
       call_center.cc_get_lookup(uc.id, uc.name) as created_by,
       i.updated_at,
       call_center.cc_get_lookup(u.id, u.name) as updated_by,
       coalesce(cls.count, 0) count
from i
    left join directory.wbt_user uc on uc.id = i.created_by
    left join directory.wbt_user u on u.id = i.updated_by
    left join call_center.cc_list_statistics cls on i.id = cls.list_id`, map[string]interface{}{
		"Name":        list.Name,
		"Description": list.Description,
		"UpdatedAt":   list.UpdatedAt,
		"UpdatedBy":   list.UpdatedBy.GetSafeId(),
		"Id":          list.Id,
		"DomainId":    list.DomainId,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_list.update.app_error", fmt.Sprintf("Id=%v, %s", list.Id, err.Error()), extractCodeFromErr(err))
	}
	return list, nil
}

func (s SqlListStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_list c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_list.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}

// Communications
func (s SqlListStore) CreateCommunication(ctx context.Context, comm *model.ListCommunication) (*model.ListCommunication, model.AppError) {
	var out *model.ListCommunication
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `insert into call_center.cc_list_communications (list_id, number, description, expire_at)
values (:ListId, :Number, :Description, :ExpireAt)
returning id, list_id, number, description, expire_at`,
		map[string]interface{}{
			"ListId":      comm.ListId,
			"Number":      comm.Number,
			"Description": comm.Description,
			"ExpireAt":    comm.ExpireAt,
		}); err != nil {
		return nil, model.NewInternalError("store.sql_list.save_communication.app_error", fmt.Sprintf("number=%v, %v", comm.Number, err.Error()))
	} else {
		return out, nil
	}
}

func (s SqlListStore) GetAllPageCommunication(ctx context.Context, domainId, listId int64, search *model.SearchListCommunication) ([]*model.ListCommunication, model.AppError) {
	var communication []*model.ListCommunication

	f := map[string]interface{}{
		"DomainId":   domainId,
		"ListId":     listId,
		"Ids":        pq.Array(search.Ids),
		"Q":          search.GetQ(),
		"ExpireFrom": model.GetBetweenFromTime(search.ExpireAt),
		"ExpireTo":   model.GetBetweenToTime(search.ExpireAt),
	}

	err := s.ListQuery(ctx, &communication, search.ListRequest,
		`domain_id = :DomainId
				and list_id = :ListId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (number ilike :Q::varchar ))
				and ( :ExpireFrom::timestamptz isnull or expire_at >= :ExpireFrom::timestamptz )
				and ( :ExpireTo::timestamptz isnull or expire_at <= :ExpireTo::timestamptz )
			`,
		model.ListCommunication{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_list.get_all_communication.app_error", err.Error())
	} else {
		return communication, nil
	}
}

func (s SqlListStore) GetCommunication(ctx context.Context, domainId, listId int64, id int64) (*model.ListCommunication, model.AppError) {
	var communication *model.ListCommunication
	if err := s.GetReplica().WithContext(ctx).SelectOne(&communication, `
			select i.id, i.number, i.description, i.list_id, i.expire_at
from call_center.cc_list_communications i
where i.id = :Id and i.list_id = :ListId  and exists(select 1 from call_center.cc_list l where l.id = i.list_id and l.domain_id = :DomainId)	
		`, map[string]interface{}{"ListId": listId, "Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_list.get_communication.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return communication, nil
	}
}

func (s SqlListStore) UpdateCommunication(ctx context.Context, domainId int64, communication *model.ListCommunication) (*model.ListCommunication, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&communication, `update call_center.cc_list_communications i
set number = :Number,
    description = :Description,
	expire_at = :ExpireAt
where list_id = :ListId and id = :Id and exists(select 1 from call_center.cc_list l where l.id = i.list_id and l.domain_id = :DomainId)
returning *`, map[string]interface{}{
		"Number":      communication.Number,
		"Description": communication.Description,
		"ListId":      communication.ListId,
		"Id":          communication.Id,
		"DomainId":    domainId,
		"ExpireAt":    communication.ExpireAt,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_list.update_communication.app_error", fmt.Sprintf("Id=%v, %s", communication.Id, err.Error()), extractCodeFromErr(err))
	}
	return communication, nil
}

func (s SqlListStore) DeleteCommunication(ctx context.Context, domainId, listId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_list_communications i where i.id=:Id and i.list_id = :ListId
    and exists(select 1 from call_center.cc_list l where l.id = i.list_id and l.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "ListId": listId}); err != nil {
		return model.NewInternalError("store.sql_list.delete_communication.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}
