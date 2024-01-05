package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateWebHook(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError) {
	hook.Key = model.NewId()
	return app.Store.WebHook().Create(ctx, domainId, hook)
}

func (a *App) SearchWebHook(ctx context.Context, domainId int64, search *model.SearchWebHook) ([]*model.WebHook, bool, model.AppError) {
	list, err := a.Store.WebHook().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetWebHook(ctx context.Context, domainId int64, id int32) (*model.WebHook, model.AppError) {
	return a.Store.WebHook().Get(ctx, domainId, id)
}

func (a *App) UpdateWebHook(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError) {
	oldHook, err := a.GetWebHook(ctx, domainId, hook.Id)
	if err != nil {
		return nil, err
	}

	oldHook.UpdatedBy = hook.UpdatedBy
	oldHook.UpdatedAt = hook.UpdatedAt
	oldHook.Name = hook.Name
	oldHook.Description = hook.Description
	oldHook.Origin = hook.Origin
	oldHook.Schema = hook.Schema
	oldHook.Enabled = hook.Enabled
	oldHook.Authorization = hook.Authorization

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = a.Store.WebHook().Update(ctx, domainId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (a *App) PatchWebHook(ctx context.Context, domainId int64, id int32, patch *model.WebHookPatch) (*model.WebHook, model.AppError) {
	oldHook, err := a.GetWebHook(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldHook.Patch(patch)

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = a.Store.WebHook().Update(ctx, domainId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (a *App) RemoveWebHook(ctx context.Context, domainId int64, id int32) (*model.WebHook, model.AppError) {
	hook, err := a.GetWebHook(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.WebHook().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return hook, nil
}
