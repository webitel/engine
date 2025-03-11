package sqlstore

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlPresetQueryStore struct {
	SqlStore
}

func (s SqlPresetQueryStore) Create(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&preset, `with p as (
    insert into call_center.cc_preset_query (domain_id, user_id, name, created_at, updated_at, preset, description, section)
    select :DomainId, :UserId, :Name, :CreatedAt, :UpdatedAt, :Preset, :Description, :Section
    where exists(select 1 from directory.wbt_user u where u.dc = :DomainId and u.id = :UserId)
    returning *
)
select
    p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    p.section,
    p.preset
from p`, map[string]interface{}{
		"DomainId":    domainId,
		"UserId":      userId,
		"Name":        preset.Name,
		"CreatedAt":   preset.CreatedAt,
		"UpdatedAt":   preset.UpdatedAt,
		"Preset":      preset.Preset.ToSafeBytes(),
		"Description": preset.Description,
		"Section":     preset.Section,
	})

	if err != nil {
		switch e := err.(type) {
		case *pq.Error: // TODO
			if e.Constraint == "cc_preset_query_user_id_name_uindex" {
				return nil, model.NewCustomCodeError("store.sql_preset_query.save.name", err.Error(), http.StatusConflict)
			}
		}
		return nil, model.NewCustomCodeError("store.sql_preset_query.save.app_error", err.Error(), extractCodeFromErr(err))
	}

	return preset, nil
}

func (s SqlPresetQueryStore) GetAllPage(ctx context.Context, domainId, userId int64, search *model.SearchPresetQuery) ([]*model.PresetQuery, model.AppError) {
	var list []*model.PresetQuery

	f := map[string]interface{}{
		"DomainId": domainId,
		"UserId":   userId,
		"Ids":      pq.Array(search.Ids),
		"Sections": pq.Array(search.Section),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and user_id = :UserId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Sections::varchar[] isnull or section = any(:Sections))
				and (:Q::varchar isnull or (name ilike :Q::varchar))
	`,
		model.PresetQuery{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_preset_query.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlPresetQueryStore) Get(ctx context.Context, domainId, userId int64, id int32) (*model.PresetQuery, model.AppError) {
	var preset *model.PresetQuery
	f := map[string]interface{}{
		"DomainId": domainId,
		"UserId":   userId,
		"Id":       id,
	}

	err := s.One(ctx, &preset,
		`domain_id = :DomainId and id = :Id and user_id = :UserId`,
		model.PresetQuery{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_preset_query.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return preset, nil
}

func (s SqlPresetQueryStore) Update(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&preset, `with p as (
    update call_center.cc_preset_query
	set updated_at = :UpdatedAt,
		name = :Name,
		preset = :Preset,
		description = :Description,
		section = :Section
	where id = :Id and domain_id  = :DomainId and user_id = :UserId
	returning *
)
select
    p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    p.section,
    p.preset
from p`, map[string]interface{}{
		"Id":          preset.Id,
		"DomainId":    domainId,
		"UserId":      userId,
		"Name":        preset.Name,
		"CreatedAt":   preset.CreatedAt,
		"UpdatedAt":   preset.UpdatedAt,
		"Preset":      preset.Preset.ToSafeBytes(),
		"Description": preset.Description,
		"Section":     preset.Section,
	})

	if err != nil {
		switch e := err.(type) {
		case *pq.Error: // TODO
			if e.Constraint == "cc_preset_query_user_id_name_uindex" {
				return nil, model.NewCustomCodeError("store.sql_preset_query.update.name", err.Error(), http.StatusConflict)
			}
		}
		return nil, model.NewCustomCodeError("store.sql_preset_query.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return preset, nil
}

func (s SqlPresetQueryStore) Delete(ctx context.Context, domainId, userId int64, id int32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_preset_query c where c.id=:Id and c.domain_id = :DomainId and user_id = :UserId`,
		map[string]interface{}{
			"Id":       id,
			"UserId":   userId,
			"DomainId": domainId,
		}); err != nil {
		return model.NewCustomCodeError("store.sql_preset_query.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func NewSqlPresetQueryStore(sqlStore SqlStore) store.PresetQueryStore {
	us := &SqlPresetQueryStore{sqlStore}
	return us
}
