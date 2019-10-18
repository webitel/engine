package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlOutboundResourceInGroupStore struct {
	SqlStore
}

func NewSqlOutboundResourceInGroupStore(sqlStore SqlStore) store.OutboundResourceInGroupStore {
	us := &SqlOutboundResourceInGroupStore{sqlStore}
	return us
}

func (s SqlOutboundResourceInGroupStore) Create(domainId, resourceId, groupId int64) (*model.OutboundResourceInGroup, *model.AppError) {
	var out *model.OutboundResourceInGroup
	if err := s.GetMaster().SelectOne(&out, `with s as (
    insert into cc_outbound_resource_in_group (resource_id, group_id)
    select :ResourceId, :GroupId
    where exists(select 1 from cc_outbound_resource_group where domain_id = :DomainId)
    returning *
)
select s.id, s.group_id, cc_get_lookup(cor.id, cor.name) as resource
from s
    inner join cc_outbound_resource cor on s.resource_id = cor.id`,
		map[string]interface{}{
			"DomainId":   domainId,
			"ResourceId": resourceId,
			"GroupId":    groupId,
		}); nil != err {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Save", "store.sql_out_resource_in_group.save.app_error", nil,
			fmt.Sprintf("GroupId=%v, %v", groupId, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlOutboundResourceInGroupStore) GetAllPage(domainId, groupId int64, offset, limit int) ([]*model.OutboundResourceInGroup, *model.AppError) {
	var groups []*model.OutboundResourceInGroup
	if _, err := s.GetReplica().Select(&groups, `
			select s.id, s.group_id, cc_get_lookup(cor.id, cor.name) as resource
from cc_outbound_resource_in_group s
    inner join cc_outbound_resource cor on s.resource_id = cor.id
    inner join cc_outbound_resource_group corg on s.group_id = corg.id
where s.group_id = :GroupId and cor.domain_id = :DomainId and corg.domain_id = :DomainId
		order by s.id
		limit :Limit
		offset :Offset
		`, map[string]interface{}{"DomainId": domainId, "GroupId": groupId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.GetAllPage", "store.sql_out_resource_in_group.get_all.app_error", nil,
			fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return groups, nil
	}
}

func (s SqlOutboundResourceInGroupStore) Get(domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	var res *model.OutboundResourceInGroup
	if err := s.GetReplica().SelectOne(&res, `
			select s.id, s.group_id, cc_get_lookup(cor.id, cor.name) as resource
			from cc_outbound_resource_in_group s
				inner join cc_outbound_resource cor on s.resource_id = cor.id
				inner join cc_outbound_resource_group corg on s.group_id = corg.id
			where s.group_id = :GroupId and cor.domain_id = :DomainId and corg.domain_id = :DomainId and s.id = :Id	
		`, map[string]interface{}{"Id": id, "DomainId": domainId, "GroupId": groupId}); err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Get", "store.sql_out_resource_in_group.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlOutboundResourceInGroupStore) Update(domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {

	err := s.GetMaster().SelectOne(&res, `with s as (
    update cc_outbound_resource_in_group 
        set resource_id  = :ResourceId  
    where id = :Id and group_id = :GroupId and exists(select 1 from cc_outbound_resource_group where domain_id = :DomainId)
    returning *
)
select s.id, s.group_id, cc_get_lookup(cor.id, cor.name) as resource
from s
    inner join cc_outbound_resource cor on s.resource_id = cor.id`, map[string]interface{}{
		"ResourceId": res.Resource.Id,
		"GroupId":    res.GroupId,
		"Id":         res.Id,
		"DomainId":   domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Update", "store.sql_out_resource_in_group.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", res.Id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlOutboundResourceInGroupStore) Delete(domainId, groupId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_outbound_resource_in_group c 
			where id = :Id and group_id = :GroupId and exists(select 1 from cc_outbound_resource_group where domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "GroupId": groupId}); err != nil {
		return model.NewAppError("SqlOutboundResourceGroupStore.Delete", "store.sql_out_resource_group.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
