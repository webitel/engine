package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlOutboundResourceStore struct {
	SqlStore
}

func NewSqlOutboundResourceStore(sqlStore SqlStore) store.OutboundResourceStore {
	us := &SqlOutboundResourceStore{sqlStore}
	return us
}

func (s SqlOutboundResourceStore) Create(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError) {
	var out *model.OutboundCallResource
	if err := s.GetMaster().SelectOne(&out, `with s as (
    insert into cc_outbound_resource ("limit", enabled, updated_at, rps, domain_id, reserve, variables, number,
                                  max_successively_errors, name, error_ids, created_at, created_by, updated_by, gateway_id, description, patterns, failure_dial_delay)
values (:Limit, :Enabled, :UpdatedAt, :Rps, :DomainId, :Reserve , :Variables, :Number,
        :MaxSErrors, :Name, :ErrorIds, :CreatedAt, :CreatedBy, :UpdatedBy, :GatewayId, :Description, :Patterns, :FailureDialDelay)
	returning *
)
select s.id, s."limit", s.enabled, s.updated_at, s.rps, s.domain_id, s.reserve, s.variables, s.number,
      s.max_successively_errors, s.name, s.error_ids, s.last_error_id, s.successively_errors, 
      s.last_error_at, s.created_at, cc_get_lookup(c.id, c.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
	  cc_get_lookup(gw.id, gw.name) as gateway, s.description, s.patterns, s.failure_dial_delay
from s
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by
	left join directory.sip_gateway gw on gw.id = s.gateway_id`,
		map[string]interface{}{
			"Limit":            resource.Limit,
			"Enabled":          resource.Enabled,
			"UpdatedAt":        resource.UpdatedAt,
			"Rps":              resource.RPS,
			"DomainId":         resource.DomainId,
			"Reserve":          resource.Reserve,
			"Variables":        resource.Variables.ToJson(),
			"Number":           resource.Number,
			"MaxSErrors":       resource.MaxSuccessivelyErrors,
			"Name":             resource.Name,
			"ErrorIds":         pq.Array(resource.ErrorIds),
			"CreatedAt":        resource.CreatedAt,
			"CreatedBy":        resource.CreatedBy.Id,
			"UpdatedBy":        resource.UpdatedBy.Id,
			"GatewayId":        resource.GetGatewayId(),
			"Description":      resource.Description,
			"Patterns":         pq.Array(resource.Patterns),
			"FailureDialDelay": resource.FailureDialDelay,
		}); nil != err {
		return nil, model.NewAppError("SqlOutboundResourceStore.Save", "store.sql_out_resource.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", resource.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlOutboundResourceStore) CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from cc_outbound_resource_acl a
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

func (s SqlOutboundResourceStore) GetAllPage(domainId int64, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, *model.AppError) {
	var resources []*model.OutboundCallResource

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&resources, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.OutboundCallResource{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.GetAllPage", "store.sql_out_resource.get_all.app_error", nil,
			fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return resources, nil
	}
}

func (s SqlOutboundResourceStore) GetAllPageByGroups(domainId int64, groups []int, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, *model.AppError) {
	var resources []*model.OutboundCallResource

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(&resources, search.ListRequest,
		`domain_id = :DomainId
				and exists(select 1
				  from cc_outbound_resource_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.OutboundCallResource{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.GetAllPage", "store.sql_out_resource.get_all.app_error", nil,
			fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return resources, nil
	}
}

func (s SqlOutboundResourceStore) Get(domainId int64, id int64) (*model.OutboundCallResource, *model.AppError) {
	var resource *model.OutboundCallResource
	if err := s.GetReplica().SelectOne(&resource, `
			select s.id, s."limit", s.enabled, s.updated_at, s.rps, s.domain_id, s.reserve, s.variables, s.number,
				  s.max_successively_errors, s.name, s.error_ids, s.last_error_id, s.successively_errors, 
				   s.last_error_at, s.created_at, cc_get_lookup(c.id, c.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
				  cc_get_lookup(gw.id, gw.name) as gateway, s.description, s.patterns, s.failure_dial_delay
			from cc_outbound_resource s
				left join directory.wbt_user c on c.id = s.created_by
				left join directory.wbt_user u on u.id = s.updated_by
				left join directory.sip_gateway gw on gw.id = s.gateway_id
		where s.domain_id = :DomainId and s.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.Get", "store.sql_out_resource.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return resource, nil
	}
}

func (s SqlOutboundResourceStore) Update(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError) {

	err := s.GetMaster().SelectOne(&resource, `
with s as (
    update cc_outbound_resource
        set "limit" = :Limit,
            enabled = :Enabled,
            updated_at = :UpdatedAt,
            updated_by = :UpdatedBy,
            rps = :Rps,
            reserve = :Reserve,
            variables = :Variables,
            number = :Number,
            max_successively_errors = :MaxSErrors,
            name = :Name,
            error_ids = :ErrorIds,
			gateway_id = :GatewayId,
			description = :Description,
			patterns = :Patterns,
			failure_dial_delay = :FailureDialDelay
        where id = :Id and domain_id = :DomainId
        returning *
)
select s.id, s."limit", s.enabled, s.updated_at, s.rps, s.domain_id, s.reserve, s.variables, s.number,
      s.max_successively_errors, s.name, s.error_ids, s.last_error_id, s.successively_errors, 
       s.last_error_at, s.created_at, cc_get_lookup(c.id, c.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
		cc_get_lookup(gw.id, gw.name) as gateway, s.description, s.patterns, s.failure_dial_delay
from s
    left join directory.wbt_user c on c.id = s.created_by
    left join directory.wbt_user u on u.id = s.updated_by
	left join directory.sip_gateway gw on gw.id = s.gateway_id`, map[string]interface{}{
		"Limit":            resource.Limit,
		"Enabled":          resource.Enabled,
		"UpdatedAt":        resource.UpdatedAt,
		"UpdatedBy":        resource.UpdatedBy.Id,
		"Rps":              resource.RPS,
		"Reserve":          resource.Reserve,
		"Variables":        resource.Variables.ToJson(),
		"Number":           resource.Number,
		"MaxSErrors":       resource.MaxSuccessivelyErrors,
		"Name":             resource.Name,
		"ErrorIds":         pq.Array(resource.ErrorIds),
		"Id":               resource.Id,
		"DomainId":         resource.DomainId,
		"GatewayId":        resource.GetGatewayId(),
		"Description":      resource.Description,
		"Patterns":         pq.Array(resource.Patterns),
		"FailureDialDelay": resource.FailureDialDelay,
	})

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.Update", "store.sql_out_resource.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", resource.Id, err.Error()), extractCodeFromErr(err))
	}

	return resource, nil
}

func (s SqlOutboundResourceStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_outbound_resource c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlOutboundResourceStore.Delete", "store.sql_out_resource.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlOutboundResourceStore) SaveDisplay(d *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError) {
	var out *model.ResourceDisplay
	err := s.GetMaster().SelectOne(&out, `insert into cc_outbound_resource_display (resource_id, display)
values (:ResourceId, :Display)
returning *`, map[string]interface{}{
		"ResourceId": d.ResourceId,
		"Display":    d.Display,
	})

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.SaveDisplay", "store.sql_out_resource.save_display.app_error", nil,
			fmt.Sprintf("name=%v, %v", d.Display, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlOutboundResourceStore) GetDisplayAllPage(domainId, resourceId int64, search *model.SearchResourceDisplay) ([]*model.ResourceDisplay, *model.AppError) {
	var list []*model.ResourceDisplay

	f := map[string]interface{}{
		"DomainId":   domainId,
		"ResourceId": resourceId,
		"Ids":        pq.Array(search.Ids),
		"Q":          search.GetQ(),
	}

	err := s.ListQuery(&list, search.ListRequest,
		`domain_id = :DomainId
				and resource_id = :ResourceId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (display ilike :Q::varchar ))`,
		model.ResourceDisplay{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.GetDisplayAllPage", "store.sql_out_resource.get_display_all.app_error", nil,
			fmt.Sprintf("ResourceId=%v, %s", resourceId, err.Error()), extractCodeFromErr(err))
	} else {
		return list, nil
	}
}

func (s SqlOutboundResourceStore) GetDisplay(domainId, resourceId, id int64) (*model.ResourceDisplay, *model.AppError) {
	var res *model.ResourceDisplay
	if err := s.GetReplica().SelectOne(&res, `
			select d.id, d.display, d.resource_id
		from cc_outbound_resource_display d
		where d.id = :Id and d.resource_id = :ResourceId and exists (select 1
				from cc_outbound_resource r where r.id = :ResourceId and r.domain_id = :DomainId)	
		`, map[string]interface{}{"Id": id, "DomainId": domainId, "ResourceId": resourceId}); err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.GetDisplay", "store.sql_out_resource.get_display.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlOutboundResourceStore) UpdateDisplay(domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError) {
	err := s.GetMaster().SelectOne(&display, `
		update cc_outbound_resource_display d
set display = :Display 
where d.id = :Id and d.resource_id = :ResourceId 
  and exists(select 1 from cc_outbound_resource r where r.id = d.resource_id and r.domain_id = :DomainId)
returning *`, map[string]interface{}{
		"Display":    display.Display,
		"Id":         display.Id,
		"ResourceId": display.ResourceId,
		"DomainId":   domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceStore.UpdateDisplay", "store.sql_out_resource.update_display.app_error", nil,
			fmt.Sprintf("Id=%v, %s", display.Id, err.Error()), extractCodeFromErr(err))
	}

	return display, nil
}

func (s SqlOutboundResourceStore) DeleteDisplay(domainId, resourceId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_outbound_resource_display d
		where d.id = :Id and d.resource_id = :ResourceId and exists(select 1 from cc_outbound_resource r where r.id = d.resource_id and r.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "ResourceId": resourceId}); err != nil {
		return model.NewAppError("SqlOutboundResourceStore.DeleteDisplay", "store.sql_out_resource.delete_display.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
