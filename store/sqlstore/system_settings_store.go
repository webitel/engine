package sqlstore

import (
	"context"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlSystemSettingsStore struct {
	SqlStore
}

func NewSqlSystemSettingsStore(sqlStore SqlStore) store.SystemSettingsStore {
	us := &SqlSystemSettingsStore{sqlStore}
	return us
}

func (s SqlSystemSettingsStore) Create(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	var st *model.SystemSetting
	err := s.GetMaster().SelectOne(&st, `with s as (
    insert into call_center.system_settings (domain_id, name, value)
    values (:DomainId, :Name, :Value)
    returning *
)
select s.id, s.name, s.value
from s;`, map[string]interface{}{
		"DomainId": domainId,
		"Name":     setting.Name,
		"Value":    setting.Value,
	})

	if err != nil {
		return nil, model.NewInternalError("store.sql_sys_settings.save.app_error", fmt.Sprintf("name=%v, %v", setting.Name, err.Error()))
	}

	return st, nil
}

func (s SqlSystemSettingsStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchSystemSetting) ([]*model.SystemSetting, model.AppError) {
	var list []*model.SystemSetting

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar))`,
		model.SystemSetting{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_sys_settings.get_all.app_error", err.Error())
	} else {
		return list, nil
	}
}

func (s SqlSystemSettingsStore) Get(ctx context.Context, domainId int64, id int32) (*model.SystemSetting, model.AppError) {
	//TODO implement me
	panic("implement me")
}

func (s SqlSystemSettingsStore) Update(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	//TODO implement me
	panic("implement me")
}

func (s SqlSystemSettingsStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {
	//TODO implement me
	panic("implement me")
}
