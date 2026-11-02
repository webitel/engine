package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type teamChatTag struct {
	*API
	engine.UnsafeTeamChatTagServiceServer
}

func NewTeamChatTagApi(api *API) *teamChatTag {
	return &teamChatTag{API: api}
}

func (api *teamChatTag) CreateChatTag(ctx context.Context, in *engine.CreateChatTagRequest) (*engine.ChatTag, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	tag := &model.TeamChatTag{
		Tag: in.Tag,
	}

	tag, err = api.ctrl.CreateChatTag(ctx, session, in.TeamId, tag)
	if err != nil {
		return nil, err
	}

	return toEngineChatTag(tag), nil
}

func (api *teamChatTag) SearchChatTag(ctx context.Context, in *engine.SearchChatTagRequest) (*engine.ListChatTag, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.TeamChatTag
	var endList bool
	req := &model.SearchTeamChatTag{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.GetId(),
	}

	list, endList, err = api.ctrl.SearchChatTag(ctx, session, in.TeamId, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.ChatTag, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineChatTag(v))
	}
	return &engine.ListChatTag{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *teamChatTag) UpdateChatTag(ctx context.Context, in *engine.UpdateChatTagRequest) (*engine.ChatTag, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	tag := &model.TeamChatTag{
		Id:  in.Id,
		Tag: in.Tag,
	}

	tag, err = api.ctrl.UpdateChatTag(ctx, session, in.TeamId, tag)
	if err != nil {
		return nil, err
	}

	return toEngineChatTag(tag), nil
}

func (api *teamChatTag) DeleteChatTag(ctx context.Context, in *engine.DeleteChatTagRequest) (*engine.ChatTag, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	tag, err := api.ctrl.DeleteChatTag(ctx, session, in.TeamId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineChatTag(tag), nil
}

func toEngineChatTag(src *model.TeamChatTag) *engine.ChatTag {
	if src == nil {
		return nil
	}
	return &engine.ChatTag{
		Id:  src.Id,
		Tag: src.Tag,
	}
}
