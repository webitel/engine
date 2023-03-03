package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) CreateList(ctx context.Context, list *model.List) (*model.List, *model.AppError) {
	return a.Store.List().Create(ctx, list)
}

func (a *App) ListCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.List().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) GetListPage(ctx context.Context, domainId int64, search *model.SearchList) ([]*model.List, bool, *model.AppError) {
	list, err := a.Store.List().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetListPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchList) ([]*model.List, bool, *model.AppError) {
	list, err := a.Store.List().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetListById(ctx context.Context, domainId, id int64) (*model.List, *model.AppError) {
	return a.Store.List().Get(ctx, domainId, id)
}

func (a *App) UpdateList(ctx context.Context, list *model.List) (*model.List, *model.AppError) {
	oldList, err := a.GetListById(ctx, list.DomainId, list.Id)
	if err != nil {
		return nil, err
	}

	oldList.Description = list.Description
	oldList.Name = list.Name
	oldList.UpdatedAt = list.UpdatedAt
	oldList.UpdatedBy = list.UpdatedBy

	oldList, err = a.Store.List().Update(ctx, oldList)
	if err != nil {
		return nil, err
	}

	return oldList, nil
}

func (a *App) RemoveList(ctx context.Context, domainId, id int64) (*model.List, *model.AppError) {
	list, err := a.Store.List().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.List().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (a *App) CreateListCommunication(ctx context.Context, comm *model.ListCommunication) (*model.ListCommunication, *model.AppError) {
	return a.Store.List().CreateCommunication(ctx, comm)
}

func (a *App) GetListCommunicationPage(ctx context.Context, domainId, listId int64, search *model.SearchListCommunication) ([]*model.ListCommunication, bool, *model.AppError) {
	list, err := a.Store.List().GetAllPageCommunication(ctx, domainId, listId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetListCommunicationById(ctx context.Context, domainId, listId, id int64) (*model.ListCommunication, *model.AppError) {
	return a.Store.List().GetCommunication(ctx, domainId, listId, id)
}

func (a *App) UpdateListCommunication(ctx context.Context, domainId int64, communication *model.ListCommunication) (*model.ListCommunication, *model.AppError) {
	oldComm, err := a.GetListCommunicationById(ctx, domainId, communication.ListId, communication.Id)
	if err != nil {
		return nil, err
	}

	oldComm.Description = communication.Description
	oldComm.Number = communication.Number
	oldComm.ExpireAt = communication.ExpireAt

	oldComm, err = a.Store.List().UpdateCommunication(ctx, domainId, oldComm)
	if err != nil {
		return nil, err
	}

	return oldComm, nil
}

func (a *App) RemoveListCommunication(ctx context.Context, domainId, listId, id int64) (*model.ListCommunication, *model.AppError) {
	communication, err := a.Store.List().GetCommunication(ctx, domainId, listId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.List().DeleteCommunication(ctx, domainId, listId, id)
	if err != nil {
		return nil, err
	}
	return communication, nil
}
