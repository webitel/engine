package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlOutboundResourceGroupStore struct {
	SqlStore
}

func NewSqlOutboundResourceGroupStore(sqlStore SqlStore) store.OutboundResourceGroupStore {
	us := &SqlOutboundResourceGroupStore{sqlStore}
	return us
}

func (s SqlOutboundResourceGroupStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_outbound_resource_group_acl a
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

func (s SqlOutboundResourceGroupStore) Create(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, model.AppError) {
	var out *model.OutboundResourceGroup
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with s as (
    insert into call_center.cc_outbound_resource_group (domain_id, name, strategy, description, communication_id, created_at,
                                        created_by, updated_at, updated_by, time)
values (:DomainId, :Name, :Strategy, :Description, :CommunicationId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :Time)
returning  *
)
select s.id, s.domain_id, s.name, s.strategy, s.description,  call_center.cc_get_lookup(comm.id, comm.name) as communication,
       s.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by, s.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, s.time
from s
    inner join call_center.cc_communication comm on comm.id = s.communication_id
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by`,
		map[string]interface{}{
			"CreatedAt":       group.CreatedAt,
			"CreatedBy":       group.CreatedBy.GetSafeId(),
			"UpdatedAt":       group.UpdatedAt,
			"UpdatedBy":       group.UpdatedBy.GetSafeId(),
			"DomainId":        group.DomainId,
			"Name":            group.Name,
			"Strategy":        group.Strategy,
			"Description":     group.Description,
			"CommunicationId": group.Communication.Id,
			"Time":            model.OutboundResourceGroupTimesToJson(group.Time),
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_out_resource_group.save.app_error", fmt.Sprintf("name=%v, %v", group.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlOutboundResourceGroupStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, model.AppError) {
	var groups []*model.OutboundResourceGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &groups, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.OutboundResourceGroup{}, f)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_out_resource_group.get_all.app_error", fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return groups, nil
	}
}

func (s SqlOutboundResourceGroupStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, model.AppError) {
	var res []*model.OutboundResourceGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &res, search.ListRequest,
		`domain_id = :DomainId
				and exists(select 1
					  from call_center.cc_outbound_resource_group_acl a
					  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.OutboundResourceGroup{}, f)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_out_resource_group.get_all.app_error", fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlOutboundResourceGroupStore) Get(ctx context.Context, domainId int64, id int64) (*model.OutboundResourceGroup, model.AppError) {
	var group *model.OutboundResourceGroup
	if err := s.GetReplica().WithContext(ctx).SelectOne(&group, `
			select s.id, s.domain_id, s.name, s.strategy, s.description,  call_center.cc_get_lookup(comm.id, comm.name) as communication,
				   s.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by, s.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, s.time
			from call_center.cc_outbound_resource_group s
				inner join call_center.cc_communication comm on comm.id = s.communication_id
				left join directory.wbt_user c on c.id = s.created_by
				left join directory.wbt_user u on u.id = s.updated_by
		where s.domain_id = :DomainId and s.id = :Id	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_out_resource_group.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return group, nil
	}
}

func (s SqlOutboundResourceGroupStore) Update(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&group, `with s as (
    update call_center.cc_outbound_resource_group
    set name = :Name,
        strategy = :Strategy,
        description = :Description,
        communication_id = :CommunicationId,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt,
		time = :Time
    where id = :Id and domain_id = :DomainId
	returning *
)
select s.id, s.domain_id, s.name, s.strategy, s.description,  call_center.cc_get_lookup(comm.id, comm.name) as communication,
       s.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by, s.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, s.time
from s
    inner join call_center.cc_communication comm on comm.id = s.communication_id
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by`, map[string]interface{}{
		"Name":            group.Name,
		"Strategy":        group.Strategy,
		"Description":     group.Description,
		"CommunicationId": group.Communication.Id,
		"UpdatedBy":       group.UpdatedBy.GetSafeId(),
		"UpdatedAt":       group.UpdatedAt,
		"Id":              group.Id,
		"DomainId":        group.DomainId,
		"Time":            model.OutboundResourceGroupTimesToJson(group.Time),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_out_resource_group.update.app_error", fmt.Sprintf("Id=%v, %s", group.Id, err.Error()), extractCodeFromErr(err))
	}

	return group, nil
}

func (s SqlOutboundResourceGroupStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_outbound_resource_group c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_out_resource_group.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}
