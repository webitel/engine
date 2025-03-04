package grpc_api

import (
	"context"

	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"github.com/webitel/engine/model"
)

type quickReply struct {
	*API
	gogrpc.UnsafeQuickRepliesServiceServer
}

func NewQuickReply(api *API) *quickReply {
	return &quickReply{API: api}
}

func (api *quickReply) CreateQuickReply(ctx context.Context, in *engine.CreateQuickReplyRequest) (*engine.QuickReply, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.Name == "" {
		return nil, model.NewBadRequestError("grpc.quickReply.create.name", "field Name is required")
	}

	if in.Text == "" {
		return nil, model.NewBadRequestError("grpc.quickReply.create.text", "field Text is required")
	}

	reply := &model.QuickReply{
		Name:    in.Name,
		Text:    in.Text,
		Article: GetLookup(in.Article),
		Teams:   GetLookups(in.Teams),
		Queues:  GetLookups(in.Queues),
	}

	replyq, err := api.ctrl.CreateQuickReply(ctx, session, reply)
	if err != nil {
		return nil, err
	}

	return toEngineQuickReply(replyq), nil
}

func (api *quickReply) SearchQuickReplies(ctx context.Context, in *engine.SearchQuickRepliesRequest) (*engine.ListQuickReplies, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.QuickReply
	var endList bool
	req := &model.SearchQuickReply{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.GetId(),
	}

	list, endList, err = api.ctrl.SearchQuickReply(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.QuickReply, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineQuickReply(v))
	}
	return &engine.ListQuickReplies{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *quickReply) ReadQuickReply(ctx context.Context, in *engine.ReadQuickReplyRequest) (*engine.QuickReply, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var reply *model.QuickReply
	reply, err = api.ctrl.GetQuickReply(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineQuickReply(reply), nil
}

func (api *quickReply) PatchQuickReply(ctx context.Context, in *engine.PatchQuickReplyRequest) (*engine.QuickReply, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var reply *model.QuickReply
	patch := &model.QuickReplyPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "text":
			patch.Text = model.NewString(in.Text)
		}
	}

	if reply, err = api.ctrl.PatchQuickReply(ctx, session, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineQuickReply(reply), nil
}

func (api *quickReply) UpdateQuickReply(ctx context.Context, in *engine.UpdateQuickReplyRequest) (*engine.QuickReply, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	reply := &model.QuickReply{
		AclRecord: model.AclRecord{},
		Id:        int(in.Id),
		Name:      in.Name,
		Text:      in.Text,
		Article:   GetLookup(in.Article),
		Teams:     GetLookups(in.Teams),
		Queues:    GetLookups(in.Queues),
	}

	reply, err = api.ctrl.UpdateQuickReply(ctx, session, reply)

	if err != nil {
		return nil, err
	}

	return toEngineQuickReply(reply), nil
}

func (api *quickReply) DeleteQuickReply(ctx context.Context, in *engine.DeleteQuickReplyRequest) (*engine.QuickReply, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var reply *model.QuickReply
	reply, err = api.ctrl.DeleteQuickReply(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineQuickReply(reply), nil
}

func toEngineQuickReply(src *model.QuickReply) *engine.QuickReply {
	return &engine.QuickReply{
		Id:        uint32(src.Id),
		CreatedAt: model.TimeToInt64(src.CreatedAt),
		CreatedBy: GetProtoLookup(src.CreatedBy),
		UpdatedAt: model.TimeToInt64(src.UpdatedAt),
		UpdatedBy: GetProtoLookup(src.UpdatedBy),
		Name:      src.Name,
		Text:      src.Text,
		Queues:    GetProtoLookups(src.Queues),
		Teams:     GetProtoLookups(src.Teams),
		Article:   GetProtoLookup(src.Article),
	}
}
