package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateSystemSetting(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	return app.Store.SystemSettings().Create(ctx, domainId, setting)
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

func (a *App) UpdateSystemSetting(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	oldSetting, err := a.GetSystemSetting(ctx, domainId, setting.Id)
	if err != nil {
		return nil, err
	}

	oldSetting.Value = setting.Value

	if err = oldSetting.IsValid(); err != nil {
		return nil, err
	}

	oldSetting, err = a.Store.SystemSettings().Update(ctx, domainId, oldSetting)
	if err != nil {
		return nil, err
	}

	return oldSetting, nil
}

func (a *App) PatchSystemSetting(ctx context.Context, domainId int64, id int32, patch *model.SystemSettingPath) (*model.SystemSetting, model.AppError) {
	oldSetting, err := a.GetSystemSetting(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldSetting.Patch(patch)

	if err = oldSetting.IsValid(); err != nil {
		return nil, err
	}

	oldSetting, err = a.Store.SystemSettings().Update(ctx, domainId, oldSetting)
	if err != nil {
		return nil, err
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

func (a *App) GetAvailableSystemSetting(ctx context.Context, domainId int64) ([]string, model.AppError) {
	list, err := a.Store.SystemSettings().Available(ctx, domainId)
	if err != nil {
		return nil, err
	}
	return list, nil
}
