package app

import "github.com/webitel/engine/model"

func (app *App) CreateCommunicationType(comm *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	return app.Store.CommunicationType().Create(comm)
}

func (app *App) GetCommunicationTypePage(domainId int64, page, perPage int) ([]*model.CommunicationType, *model.AppError) {
	return app.Store.CommunicationType().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) GetCommunicationType(id, domainId int64) (*model.CommunicationType, *model.AppError) {
	return app.Store.CommunicationType().Get(domainId, id)
}

func (app *App) UpdateCommunicationType(cType *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	oldCType, err := app.Store.CommunicationType().Get(cType.DomainId, cType.Id)

	if err != nil {
		return nil, err
	}

	oldCType.Name = cType.Name
	oldCType.Description = cType.Description
	oldCType.Type = cType.Type
	oldCType.Code = cType.Code

	_, err = app.Store.CommunicationType().Update(oldCType)
	if err != nil {
		return nil, err
	}

	return oldCType, nil
}

func (app *App) RemoveCommunicationType(domainId, id int64) (*model.CommunicationType, *model.AppError) {
	cType, err := app.Store.CommunicationType().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.CommunicationType().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return cType, nil
}
