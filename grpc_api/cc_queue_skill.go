package grpc_api

import (
	"context"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type queueSkill struct {
	*API
	engine.UnsafeQueueSkillServiceServer
}

func NewQueueSkill(api *API) *queueSkill {
	return &queueSkill{API: api}
}

func (api *queueSkill) CreateQueueSkill(ctx context.Context, in *engine.CreateQueueSkillRequest) (*engine.QueueSkill, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	qs := &model.QueueSkill{
		QueueId:     in.QueueId,
		Buckets:     GetLookups(in.Buckets),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.GetMinCapacity().GetValue()),
		MaxCapacity: int(in.GetMaxCapacity().GetValue()),
		Enabled:     in.Enabled,
	}

	if in.Skill != nil {
		qs.Skill.Id = int(in.Skill.Id)
	}

	qs, err = api.ctrl.CreateQueueSkill(ctx, session, qs)
	if err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func (api *queueSkill) SearchQueueSkill(ctx context.Context, in *engine.SearchQueueSkillRequest) (*engine.ListQueueSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.QueueSkill
	var endList bool
	req := &model.SearchQueueSkill{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		QueueId:     in.QueueId,
		Ids:         in.GetId(),
		SkillIds:    in.SkillId,
		BucketIds:   in.BucketId,
		Lvl:         in.Lvl,
		MinCapacity: in.MinCapacity,
		MaxCapacity: in.MaxCapacity,
	}

	if in.Enabled {
		req.Enabled = &in.Enabled
	}

	list, endList, err = api.ctrl.SearchQueueSkill(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.QueueSkill, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineQueueSkill(v))
	}
	return &engine.ListQueueSkill{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *queueSkill) ReadQueueSkill(ctx context.Context, in *engine.ReadQueueSkillRequest) (*engine.QueueSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var qs *model.QueueSkill
	qs, err = api.ctrl.GetQueueSkill(ctx, session, in.QueueId, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func (api *queueSkill) UpdateQueueSkill(ctx context.Context, in *engine.UpdateQueueSkillRequest) (*engine.QueueSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	qs := &model.QueueSkill{
		Id:          in.Id,
		QueueId:     in.QueueId,
		Skill:       model.Lookup{},
		Buckets:     GetLookups(in.Buckets),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.GetMinCapacity().GetValue()),
		MaxCapacity: int(in.GetMaxCapacity().GetValue()),
		Enabled:     in.Enabled,
	}

	if in.Skill != nil {
		qs.Skill.Id = int(in.Skill.Id)
	}

	qs, err = api.ctrl.UpdateQueueSkill(ctx, session, qs)

	if err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func (api *queueSkill) PatchQueueSkill(ctx context.Context, in *engine.PatchQueueSkillRequest) (*engine.QueueSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var qs *model.QueueSkill
	patch := &model.QueueSkillPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "skill.id":
			patch.Skill = &model.Lookup{Id: int(in.GetSkill().GetId())}
		case "buckets":
			patch.Buckets = GetLookups(in.GetBuckets())
		case "lvl":
			patch.Lvl = model.NewInt(int(in.Lvl))
		case "min_capacity":
			patch.MinCapacity = model.NewInt(int(in.GetMinCapacity().GetValue()))
		case "max_capacity":
			patch.MaxCapacity = model.NewInt(int(in.GetMaxCapacity().GetValue()))
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	if qs, err = api.ctrl.PatchQueueSkill(ctx, session, in.QueueId, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func (api *queueSkill) DeleteQueueSkill(ctx context.Context, in *engine.DeleteQueueSkillRequest) (*engine.QueueSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var qs *model.QueueSkill
	qs, err = api.ctrl.DeleteQueueSkill(ctx, session, in.QueueId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func toEngineQueueSkill(src *model.QueueSkill) *engine.QueueSkill {
	return &engine.QueueSkill{
		Id:      src.Id,
		Skill:   GetProtoLookup(&src.Skill),
		Buckets: GetProtoLookups(src.Buckets),
		Lvl:     int32(src.Lvl),
		MinCapacity: &wrappers.Int32Value{
			Value: int32(src.MinCapacity),
		},
		MaxCapacity: &wrappers.Int32Value{
			Value: int32(src.MaxCapacity),
		},
		Enabled: src.Enabled,
	}
}
