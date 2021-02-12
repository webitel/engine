package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type queueSkill struct {
	*API
}

func NewQueueSkill(api *API) *queueSkill {
	return &queueSkill{api}
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
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
		Disabled:    in.Disabled,
	}

	if in.Skill != nil {
		qs.Skill.Id = int(in.Skill.Id)
	}

	qs, err = api.ctrl.CreateQueueSkill(session, qs)
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
		},
		QueueId:     in.QueueId,
		Ids:         in.GetId(),
		SkillIds:    in.SkillId,
		BucketIds:   in.BucketId,
		Lvl:         in.Lvl,
		MinCapacity: in.MinCapacity,
		MaxCapacity: in.MaxCapacity,
	}

	if in.Disabled {
		req.Disabled = &in.Disabled
	}

	list, endList, err = api.ctrl.SearchQueueSkill(session, req)

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
	qs, err = api.ctrl.GetQueueSkill(session, in.QueueId, in.Id)

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
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
		Disabled:    in.Disabled,
	}

	if in.Skill != nil {
		qs.Skill.Id = int(in.Skill.Id)
	}

	qs, err = api.ctrl.UpdateQueueSkill(session, qs)

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
			patch.MinCapacity = model.NewInt(int(in.MinCapacity))
		case "max_capacity":
			patch.MaxCapacity = model.NewInt(int(in.MaxCapacity))
		case "disabled":
			patch.Disabled = &in.Disabled
		}
	}

	if qs, err = api.ctrl.PatchQueueSkill(session, in.QueueId, in.Id, patch); err != nil {
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
	qs, err = api.ctrl.DeleteQueueSkill(session, in.QueueId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineQueueSkill(qs), nil
}

func toEngineQueueSkill(src *model.QueueSkill) *engine.QueueSkill {
	return &engine.QueueSkill{
		Id:          src.Id,
		Skill:       GetProtoLookup(&src.Skill),
		Buckets:     GetProtoLookups(src.Buckets),
		Lvl:         int32(src.Lvl),
		MinCapacity: int32(src.MinCapacity),
		MaxCapacity: int32(src.MaxCapacity),
		Disabled:    src.Disabled,
	}
}
