package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type list struct {
	app *app.App
}

func NewListApi(app *app.App) *list {
	return &list{app: app}
}

func (api *list) CreateList(ctx context.Context, in *engine.CreateListRequest) (*engine.List, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	list := &model.List{
		DomainRecord: model.DomainRecord{
			Id:        0,
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.GetDescription(),
	}

	list, err = api.app.CreateList(list)
	if err != nil {
		return nil, err
	}

	return toEngineList(list), nil
}

func (api *list) SearchList(ctx context.Context, in *engine.SearchListRequest) (*engine.ListOfList, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.List
	var endList bool
	req := &model.SearchList{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.GetListPageByGroups(session.Domain(0), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetListPage(session.Domain(0), req)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.List, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineList(v))
	}
	return &engine.ListOfList{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *list) ReadList(ctx context.Context, in *engine.ReadListRequest) (*engine.List, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list *model.List

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	list, err = api.app.GetListById(session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineList(list), nil
}

func (api *list) UpdateList(ctx context.Context, in *engine.UpdateListRequest) (*engine.List, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var list *model.List

	list, err = api.app.UpdateList(&model.List{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return toEngineList(list), nil
}

func (api *list) DeleteList(ctx context.Context, in *engine.DeleteListRequest) (*engine.List, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var list *model.List
	list, err = api.app.RemoveList(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineList(list), nil
}

func (api *list) CreateListCommunication(ctx context.Context, in *engine.CreateListCommunicationRequest) (*engine.ListCommunication, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetListId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetListId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	communication := &model.ListCommunication{
		ListId:      in.GetListId(),
		Number:      in.GetNumber(),
		Description: in.GetDescription(),
	}

	if err = communication.IsValid(); err != nil {
		return nil, err
	}

	communication, err = api.app.CreateListCommunication(communication)

	if err != nil {
		return nil, err
	}

	return toEngineListCommunication(communication), nil
}

func (api *list) SearchListCommunication(ctx context.Context, in *engine.SearchListCommunicationRequest) (*engine.ListOfListCommunication, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(0), in.GetListId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetListId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var communication []*model.ListCommunication
	var endList bool
	req := &model.SearchListCommunication{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	communication, endList, err = api.app.GetListCommunicationPage(session.Domain(0), in.GetListId(), req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.ListCommunication, 0, len(communication))
	for _, v := range communication {
		items = append(items, toEngineListCommunication(v))
	}
	return &engine.ListOfListCommunication{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *list) ReadListCommunication(ctx context.Context, in *engine.ReadListCommunicationRequest) (*engine.ListCommunication, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetListId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetListId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var communication *model.ListCommunication
	communication, err = api.app.GetListCommunicationById(session.Domain(in.GetDomainId()), in.GetListId(), in.GetId())

	if err != nil {
		return nil, err
	} else {
		return toEngineListCommunication(communication), nil
	}
}

func (api *list) UpdateListCommunication(ctx context.Context, in *engine.UpdateListCommunicationRequest) (*engine.ListCommunication, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetListId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetListId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	communication := &model.ListCommunication{
		Id:          in.GetId(),
		ListId:      in.GetListId(),
		Number:      in.GetNumber(),
		Description: in.GetDescription(),
	}

	if err = communication.IsValid(); err != nil {
		return nil, err
	}

	communication, err = api.app.UpdateListCommunication(session.Domain(in.GetDomainId()), communication)

	if err != nil {
		return nil, err
	}

	return toEngineListCommunication(communication), nil
}

func (api *list) DeleteListCommunication(ctx context.Context, in *engine.DeleteListCommunicationRequest) (*engine.ListCommunication, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_LIST)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.ListCheckAccess(session.Domain(in.GetDomainId()), in.GetListId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetListId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var communication *model.ListCommunication
	communication, err = api.app.RemoveListCommunication(session.Domain(in.GetDomainId()), in.GetListId(), in.GetId())

	if err != nil {
		return nil, err
	} else {
		return toEngineListCommunication(communication), nil
	}
}

func toEngineList(src *model.List) *engine.List {
	item := &engine.List{
		Id:          src.Id,
		DomainId:    src.DomainId,
		CreatedAt:   src.CreatedAt,
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedAt:   src.UpdatedAt,
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		Name:        src.Name,
		Description: src.Description,
		Count:       src.Count,
	}

	return item
}

func toEngineListCommunication(src *model.ListCommunication) *engine.ListCommunication {
	item := &engine.ListCommunication{
		Id:          src.Id,
		ListId:      src.ListId,
		Number:      src.Number,
		Description: src.Description,
	}

	return item
}
