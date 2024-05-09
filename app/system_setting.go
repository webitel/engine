package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/webitel/engine/model"
	"strconv"
)

const (
	MqSysSettingObjectName = "system_settings"
)

func (a *App) CreateSystemSetting(ctx context.Context, userId, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {

	setting, err := a.Store.SystemSettings().Create(ctx, domainId, setting)
	if err != nil {
		return nil, err
	}
	// publish event
	err = a.PublishSysSettingEventContext(ctx, setting, nil, EventCreateAction, strconv.FormatInt(domainId, 10), strconv.FormatInt(userId, 10))
	if err != nil {
		// event generation error
		return nil, model.NewInternalError("app.system_settings.patch_system_setting.generate_regeneration_event.error", err.Error())
	}
	return setting, nil
}

func (a *App) GetSystemSettingPage(ctx context.Context, domainId int64, search *model.SearchSystemSetting) ([]*model.SystemSetting, bool, model.AppError) {
	list, err := a.Store.SystemSettings().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetSystemSetting(ctx context.Context, domainId int64, id int32) (*model.SystemSetting, model.AppError) {
	return a.Store.SystemSettings().Get(ctx, domainId, id)
}

func (a *App) UpdateSystemSetting(ctx context.Context, userId, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	oldSetting, appErr := a.GetSystemSetting(ctx, domainId, setting.Id)
	if appErr != nil {
		return nil, appErr
	}
	oldSettingCopy := *oldSetting

	oldSetting.Value = setting.Value

	if appErr = oldSetting.IsValid(); appErr != nil {
		return nil, appErr
	}

	oldSetting, appErr = a.Store.SystemSettings().Update(ctx, domainId, oldSetting)
	if appErr != nil {
		return nil, appErr
	}
	// publish event
	appErr = a.PublishSysSettingEventContext(ctx, oldSetting, &oldSettingCopy, EventUpdateAction, strconv.FormatInt(domainId, 10), strconv.FormatInt(userId, 10))
	if appErr != nil {
		// event generation error
		return nil, model.NewInternalError("app.system_settings.patch_system_setting.generate_regeneration_event.error", appErr.Error())
	}
	return oldSetting, nil
}

func (a *App) PatchSystemSetting(ctx context.Context, userId, domainId int64, id int32, patch *model.SystemSettingPath) (*model.SystemSetting, model.AppError) {
	oldSetting, err := a.GetSystemSetting(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	oldSettingCopy := *oldSetting

	oldSetting.Patch(patch)

	if err = oldSetting.IsValid(); err != nil {
		return nil, err
	}

	oldSetting, err = a.Store.SystemSettings().Update(ctx, domainId, oldSetting)
	if err != nil {
		return nil, err
	}
	// publish event
	err = a.PublishSysSettingEventContext(ctx, oldSetting, &oldSettingCopy, EventUpdateAction, strconv.FormatInt(domainId, 10), strconv.FormatInt(userId, 10))
	if err != nil {
		// event generation error
		return nil, model.NewInternalError("app.system_settings.patch_system_setting.generate_regeneration_event.error", err.Error())
	}
	return oldSetting, nil
}

func (a *App) RemoveSystemSetting(ctx context.Context, domainId int64, id int32) (*model.SystemSetting, model.AppError) {
	setting, err := a.GetSystemSetting(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.SystemSettings().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (a *App) GetAvailableSystemSetting(ctx context.Context, domainId int64, search *model.ListRequest) ([]string, model.AppError) {
	list, err := a.Store.SystemSettings().Available(ctx, domainId, search)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PublishSysSettingEventContext handles the publishing system setting change/create/delete event to the broker, pass old setting as nil for a creation action and old = nil and new = nil for deletion.
//
// keys parameter sets the additional nodes to the message's routing key
func (a *App) PublishSysSettingEventContext(ctx context.Context, new *model.SystemSetting, old *model.SystemSetting, action string, keys ...string) model.AppError {

	// validation
	switch action {
	case EventUpdateAction:
		if old == nil || new == nil {
			return model.NewInternalError("app.system_setting.setting_event_context.args_check.bad_arg", fmt.Sprintf("[%s] action requires old and new setting copies", action))
		}
		switch new.Name {
		case model.SysNameTwoFactorAuthorization:
			oldParsed, newParsed := model.SysValue(old.Value), model.SysValue(new.Value)
			oldValue, newValue := oldParsed.Bool(), newParsed.Bool()
			if *oldValue == *newValue { // value didn't changed -- ignore
				return nil
			}
		default:
			// system setting change doesn't need an event -- ignore
			return nil
		}
	case EventCreateAction, EventDeleteAction:
		if new == nil {
			return model.NewInternalError("app.system_setting.setting_event_context.args_check.bad_arg", fmt.Sprintf("[%s] action requires new value", action))
		}
	default:
		return model.NewInternalError("app.system_setting.publish_setting_event_context.args_check.unknown_action", fmt.Sprintf("[%s] unknown action", action))
	}
	var newKeys []string
	// construct
	newKeys = append(newKeys, new.Name, action)
	newKeys = append(newKeys, keys...)
	body, err := json.Marshal(new)
	if err != nil {
		return model.NewInternalError("app.system_setting.publish_setting_event_context.update_marshal.error", err.Error())
	}
	appErr := a.PublishEventContext(ctx, body, MqSysSettingObjectName, newKeys...)
	if appErr != nil {
		return appErr
	}
	return nil
}
