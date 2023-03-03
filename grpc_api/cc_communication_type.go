package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type communicationType struct {
	*API
	engine.UnsafeCommunicationTypeServiceServer
}

//aaaa

func NewCommunicationTypeApi(api *API) *communicationType {
	return &communicationType{API: api}
}

func (api *communicationType) CreateCommunicationType(ctx context.Context, in *engine.CommunicationTypeRequest) (*engine.CommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cType := &model.CommunicationType{
		DomainId:    session.Domain(in.GetDomainId()),
		Name:        in.Name,
		Code:        in.GetCode(),
		Type:        in.GetType(),
		Description: in.Description,
	}

	cType, err = api.ctrl.CreateCommunicationType(ctx, session, cType)
	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(cType), nil
}

func (api *communicationType) SearchCommunicationType(ctx context.Context, in *engine.SearchCommunicationTypeRequest) (*engine.ListCommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.CommunicationType
	var endList bool
	req := &model.SearchCommunicationType{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.ctrl.GetCommunicationTypePage(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.CommunicationType, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineCommunicationType(v))
	}
	return &engine.ListCommunicationType{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *communicationType) ReadCommunicationType(ctx context.Context, in *engine.ReadCommunicationTypeRequest) (*engine.CommunicationType, error) {
	var cType *model.CommunicationType
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cType, err = api.ctrl.ReadCommunicationType(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(cType), nil
}

func (api *communicationType) UpdateCommunicationType(ctx context.Context, in *engine.UpdateCommunicationTypeRequest) (*engine.CommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cType := &model.CommunicationType{
		Id:          in.GetId(),
		DomainId:    session.Domain(in.GetDomainId()),
		Name:        in.GetName(),
		Code:        in.GetCode(),
		Type:        in.GetType(),
		Description: in.GetDescription(),
	}

	cType, err = api.ctrl.UpdateCommunicationType(ctx, session, cType)

	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(cType), nil
}

func (api *communicationType) DeleteCommunicationType(ctx context.Context, in *engine.DeleteCommunicationTypeRequest) (*engine.CommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var cType *model.CommunicationType
	cType, err = api.ctrl.RemoveCommunicationType(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(cType), nil
}

func toEngineCommunicationType(src *model.CommunicationType) *engine.CommunicationType {
	return &engine.CommunicationType{
		Id:          src.Id,
		DomainId:    src.DomainId,
		Name:        src.Name,
		Code:        src.Code,
		Type:        src.Type,
		Description: src.Description,
	}
}
