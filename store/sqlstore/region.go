package sqlstore

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlRegionStore struct {
	SqlStore
}

func NewSqlRegionStore(sqlStore SqlStore) store.RegionStore {
	us := &SqlRegionStore{sqlStore}
	return us
}

func (s SqlRegionStore) Create(ctx context.Context, domainId int64, region *model.Region) (*model.Region, *model.AppError) {
	var r *model.Region
	err := s.GetMaster().WithContext(ctx).SelectOne(&r, `with r as (
    insert into flow.region (domain_id, name, description, timezone_id)
    values (:DomainId, :Name, :Description, :TimezoneId)
    returning *
)
select r.id, r.name, r.description, call_center.cc_get_lookup(t.id, t.name) timezone
from  r
    left join flow.calendar_timezones t on t.id = r.timezone_id`, map[string]interface{}{
		"DomainId":    domainId,
		"Name":        region.Name,
		"Description": region.Description,
		"TimezoneId":  region.Timezone.Id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlRegionStore.Create", "store.sql_region.create.app_error", nil,
			fmt.Sprintf("name=%v, %v", region.Name, err.Error()), extractCodeFromErr(err))
	}

	return r, nil
}

func (s SqlRegionStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchRegion) ([]*model.Region, *model.AppError) {
	var region []*model.Region

	f := map[string]interface{}{
		"DomainId":    domainId,
		"Q":           search.GetQ(),
		"Ids":         pq.Array(search.Ids),
		"Name":        search.Name,
		"Description": search.Description,
		"TimezoneIds": pq.Array(search.TimezoneIds),
	}

	err := s.ListQueryFromSchema(ctx, &region, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:TimezoneIds::int4[] isnull or timezone_id = any(:TimezoneIds))
				and (:Name::text isnull or name = :Name)
				and (:Description::text isnull or description = :Description)
			`,
		model.Region{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlRegionStore.GetAllPage", "store.sql_region.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return region, nil
}

func (s SqlRegionStore) Get(ctx context.Context, domainId int64, id int64) (*model.Region, *model.AppError) {
	var region *model.Region
	err := s.GetReplica().WithContext(ctx).SelectOne(&region, `select r.id, r.name, r.description, call_center.cc_get_lookup(t.id, t.name) timezone
from flow.region r
          left join flow.calendar_timezones t on t.id = r.timezone_id
where r.domain_id = :DomainId and r.id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlRegionStore.Get", "store.sql_region.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return region, nil
}

func (s SqlRegionStore) Update(ctx context.Context, domainId int64, region *model.Region) (*model.Region, *model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&region, `with r as (
    update flow.region
        set name = :Name,
            description = :Description,
            timezone_id = :TimezoneId
    where domain_id = :DomainId and id = :Id
    returning *
)
select r.id, r.name, r.description, call_center.cc_get_lookup(t.id, t.name) timezone
from r
         left join flow.calendar_timezones t on t.id = r.timezone_id`, map[string]interface{}{
		"DomainId":    domainId,
		"Id":          region.Id,
		"Name":        region.Name,
		"Description": region.Description,
		"TimezoneId":  region.Timezone.Id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlRegionStore.Update", "store.sql_region.update.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return region, nil

}

func (s SqlRegionStore) Delete(ctx context.Context, domainId int64, id int64) *model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from flow.region c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRegionStore.Delete", "store.sql_region.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
