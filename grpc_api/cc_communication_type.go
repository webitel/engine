package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type communicationType struct {
	*API
	engine.UnsafeCommunicationTypeServiceServer
}

func NewCommunicationTypeApi(api *API) *communicationType {
	return &communicationType{API: api}
}

func (api *communicationType) CreateCommunicationType(ctx context.Context, in *engine.CommunicationTypeRequest) (*engine.CommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cType := &model.CommunicationType{
		Name:        in.Name,
		Code:        in.GetCode(),
		Channel:     in.GetChannel().String(),
		Description: in.Description,
		Default:     in.Default,
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

	if in.Default {
		req.Default = &in.Default
	}

	if len(in.Channel) != 0 {
		req.Channels = make([]string, 0, len(in.Channel))
		for _, v := range in.Channel {
			req.Channels = append(req.Channels, v.String())
		}
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
		Name:        in.GetName(),
		Code:        in.GetCode(),
		Channel:     in.GetChannel().String(),
		Description: in.GetDescription(),
		Default:     in.Default,
	}

	cType, err = api.ctrl.UpdateCommunicationType(ctx, session, cType)

	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(cType), nil
}

func (api *communicationType) PatchCommunicationType(ctx context.Context, in *engine.PatchCommunicationTypeRequest) (*engine.CommunicationType, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var ct *model.CommunicationType
	patch := &model.CommunicationTypePatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "default":
			patch.Default = &in.Default
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		case "code":
			patch.Code = &in.Code
		case "channel":
			patch.Channel = model.NewString(in.GetChannel().String())
		}
	}

	ct, err = api.ctrl.PatchCommunicationType(ctx, session, in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return toEngineCommunicationType(ct), nil
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
		Name:        src.Name,
		Code:        src.Code,
		Channel:     getChannelEnum(src.Channel),
		Description: src.Description,
		Default:     src.Default,
	}
}

func getChannelEnum(c string) engine.CommunicationChannels {
	switch c {
	case engine.CommunicationChannels_Phone.String():
		return engine.CommunicationChannels_Phone
	case engine.CommunicationChannels_Messaging.String():
		return engine.CommunicationChannels_Messaging
	case engine.CommunicationChannels_Email.String():
		return engine.CommunicationChannels_Email
	default:
		return engine.CommunicationChannels_Undefined
	}
}
