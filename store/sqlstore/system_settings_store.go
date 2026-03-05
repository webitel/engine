package sqlstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlSystemSettingsStore struct {
	SqlStore
}

var (
	allSystemSettings = []string{model.SysNameOmnichannel, model.SysNameMemberInsertChunkSize, model.SysNameSchemeVersionLimit,
		model.SysNameAmdCancelNotHuman, model.SysNameTwoFactorAuthorization, model.SysNameExportSettings,
		model.SysNameSearchNumberLength, model.SysNameChatAiConnection, model.SysNamePasswordRegExp,
		model.SysNamePasswordValidationText, model.SysNameAutolinkCallToContact, model.SysNamePeriodToPlaybackRecord,
		model.SysNameIsFulltextSearchEnabled, model.SysNameHideContact, model.SysNameShowFullContact,
		model.SysNameCallEndSoundNotification, model.SysNameCallEndPushNotification, model.SysNameChatEndSoundNotification,
		model.SysNameChatEndPushNotification, model.SysNameTaskEndSoundNotification, model.SysNameTaskEndPushNotification,
		model.SysNamePushNotificationTimeout, model.SysNameLabelsToLimitContacts, model.SysNameAutolinkMailToContact, model.SysNameNewChatSoundNotification,
		model.SysNameNewMessageSoundNotification, model.SysNameScreenshotInterval, model.SysNamePasswordExpiryDays,
		model.SysNamePasswordMinLength, model.SysNamePasswordCategories, model.SysNamePasswordContainsLogin, model.SysNamePasswordWarningDays, model.SysNameDefaultPassword,
	}
)

func NewSqlSystemSettingsStore(sqlStore SqlStore) store.SystemSettingsStore {
	us := &SqlSystemSettingsStore{sqlStore}
	return us
}

func (s SqlSystemSettingsStore) Create(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	var st *model.SystemSetting
	err := s.GetMaster().WithContext(ctx).SelectOne(&st, `with s as (
    insert into call_center.system_settings (domain_id, name, value)
    values (:DomainId::int8, :Name::varchar, :Value::jsonb)
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
	where := `domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar))`
	if len(search.Name) != 0 {
		f["Name"] = pq.Array(search.Name)
		where += " and name = any(:Name::varchar[])"
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		where,
		model.SystemSetting{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_sys_settings.get_all.app_error", err.Error())
	} else {
		return list, nil
	}
}

func (s SqlSystemSettingsStore) Get(ctx context.Context, domainId int64, id int32) (*model.SystemSetting, model.AppError) {
	var ss *model.SystemSetting
	err := s.GetReplica().WithContext(ctx).SelectOne(&ss, `select s.id, s.name, s.value
from call_center.system_settings s
where domain_id = :DomainId::int8 and id = :Id::int4`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_sys_settings.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return ss, nil
}

func (s SqlSystemSettingsStore) Update(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	var ss *model.SystemSetting
	err := s.GetMaster().WithContext(ctx).SelectOne(&ss, `with s as (
    update call_center.system_settings
        set value = :Value::jsonb
    where domain_id = :DomainId::int8 and id = :Id::int4
    returning *
)
select s.id, s.name, s.value
from s;`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       setting.Id,
		"Value":    setting.Value,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_sys_settings.update.app_error", fmt.Sprintf("Id=%v, %s", setting.Id, err.Error()), extractCodeFromErr(err))
	}

	return ss, nil
}

func (s SqlSystemSettingsStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`delete
from call_center.system_settings s
where domain_id = :DomainId::int8 and id = :Id::int4`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_sys_settings.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}

func (s SqlSystemSettingsStore) ValueByName(ctx context.Context, domainId int64, name string) (model.SysValue, model.AppError) {
	var outValue model.SysValue
	err := s.GetReplica().WithContext(ctx).SelectOne(&outValue, `select s.value
from call_center.system_settings s
where domain_id = :DomainId::int8 and name = :Name::varchar`, map[string]interface{}{
		"DomainId": domainId,
		"Name":     name,
	})

	if err != nil && err != sql.ErrNoRows {
		return nil, model.NewCustomCodeError("store.sql_sys_settings.value.app_error", fmt.Sprintf("Name=%v, %s", name, err.Error()), extractCodeFromErr(err))
	}

	return outValue, nil
}

func (s SqlSystemSettingsStore) Available(ctx context.Context, domainId int64, search *model.ListRequest) ([]string, model.AppError) {
	var res []string
	_, err := s.GetReplica().WithContext(ctx).Select(&res, `select t
from unnest(:All::varchar[]) t
where not exists(select 1 from call_center.system_settings ss where ss.domain_id = :DomainId and ss.name = t)
	and (:Q::text isnull or ( t ilike :Q::varchar))`, map[string]interface{}{
		"All":      pq.Array(allSystemSettings),
		"DomainId": domainId,
		"Q":        search.GetQ(),
	})

	if err != nil {
		return nil, model.NewInternalError("store.sql_sys_settings.get_available.app_error", err.Error())
	}

	return res, nil
}
