package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateCommunicationType(ctx context.Context, comm *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	return app.Store.CommunicationType().Create(ctx, comm)
}

func (app *App) GetCommunicationTypePage(ctx context.Context, domainId int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, bool, *model.AppError) {
	list, err := app.Store.CommunicationType().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetCommunicationType(ctx context.Context, id, domainId int64) (*model.CommunicationType, *model.AppError) {
	return app.Store.CommunicationType().Get(ctx, domainId, id)
}

func (app *App) UpdateCommunicationType(ctx context.Context, cType *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	oldCType, err := app.Store.CommunicationType().Get(ctx, cType.DomainId, cType.Id)

	if err != nil {
		return nil, err
	}

	oldCType.Name = cType.Name
	oldCType.Description = cType.Description
	oldCType.Type = cType.Type
	oldCType.Code = cType.Code

	_, err = app.Store.CommunicationType().Update(ctx, oldCType)
	if err != nil {
		return nil, err
	}

	return oldCType, nil
}

func (app *App) RemoveCommunicationType(ctx context.Context, domainId, id int64) (*model.CommunicationType, *model.AppError) {
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
