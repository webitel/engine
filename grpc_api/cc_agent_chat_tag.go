package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type agentChatTag struct {
	*API
	engine.UnsafeAgentChatTagServiceServer
}

func NewAgentChatTagApi(api *API) *agentChatTag {
	return &agentChatTag{API: api}
}

func (api *agentChatTag) SearchAgentChatTag(ctx context.Context, in *engine.SearchAgentChatTagRequest) (*engine.ListChatTag, error) {
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

	list, endList, err = api.ctrl.SearchAgentChatTag(ctx, session, req)

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
