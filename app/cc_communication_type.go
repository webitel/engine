package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateCommunicationType(ctx context.Context, domainId int64, comm *model.CommunicationType) (*model.CommunicationType, model.AppError) {
	return app.Store.CommunicationType().Create(ctx, domainId, comm)
}

func (app *App) GetCommunicationTypePage(ctx context.Context, domainId int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, bool, model.AppError) {
	list, err := app.Store.CommunicationType().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetCommunicationType(ctx context.Context, id, domainId int64) (*model.CommunicationType, model.AppError) {
	return app.Store.CommunicationType().Get(ctx, domainId, id)
}

func (app *App) UpdateCommunicationType(ctx context.Context, domainId int64, cType *model.CommunicationType) (*model.CommunicationType, model.AppError) {
	oldCType, err := app.Store.CommunicationType().Get(ctx, domainId, cType.Id)

	if err != nil {
		return nil, err
	}

	oldCType.Name = cType.Name
	oldCType.Description = cType.Description
	oldCType.Channel = cType.Channel
	oldCType.Code = cType.Code
	oldCType.Default = cType.Default

	_, err = app.Store.CommunicationType().Update(ctx, domainId, oldCType)
	if err != nil {
		return nil, err
	}

	return oldCType, nil
}

func (app *App) PatchCommunicationType(ctx context.Context, domainId int64, id int64, patch *model.CommunicationTypePatch) (*model.CommunicationType, model.AppError) {
	old, err := app.GetCommunicationType(ctx, id, domainId)
	if err != nil {
		return nil, err
	}

	old.Patch(patch)

	if err = old.IsValid(); err != nil {
		return nil, err
	}

	old, err = app.Store.CommunicationType().Update(ctx, domainId, old)
	if err != nil {
		return nil, err
	}

	return old, nil
}

func (app *App) RemoveCommunicationType(ctx context.Context, domainId, id int64) (*model.CommunicationType, model.AppError) {
	cType, err := app.Store.CommunicationType().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.CommunicationType().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return cType, nil
}
