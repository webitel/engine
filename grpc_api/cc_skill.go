package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type skill struct {
	*API
	engine.UnsafeSkillServiceServer
}

func NewSkillApi(api *API) *skill {
	return &skill{API: api}
}

func (api *skill) CreateSkill(ctx context.Context, in *engine.CreateSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skill := &model.Skill{
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	}

	skill, err = api.ctrl.CreateSkill(ctx, session, skill)
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) SearchSkill(ctx context.Context, in *engine.SearchSkillRequest) (*engine.ListSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Skill
	var endList bool
	req := &model.SearchSkill{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.ctrl.SearchSkill(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Skill, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkill(v))
	}
	return &engine.ListSkill{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *skill) ReadSkill(ctx context.Context, in *engine.ReadSkillRequest) (*engine.Skill, error) {
	var skill *model.Skill
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skill, err = api.ctrl.ReadSkill(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) UpdateSkill(ctx context.Context, in *engine.UpdateSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var skill *model.Skill

	skill, err = api.ctrl.UpdateSkill(ctx, session, &model.Skill{
		Id:          in.Id,
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) DeleteSkill(ctx context.Context, in *engine.DeleteSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var skill *model.Skill
	skill, err = api.ctrl.DeleteSkill(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) CreateSkillAgent(ctx context.Context, in *engine.CreateSkillAgentRequest) (*engine.CreateSkillAgentResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentSkill

	list, err = api.ctrl.CreateAgentsSkills(ctx, session, &model.AgentsSkills{
		DomainRecord: model.DomainRecord{},
		AgentIds:     LookupsIds(in.Agent),
		SkillIds:     []int64{in.SkillId},
		AgentSkillProps: model.AgentSkillProps{
			Capacity: int(in.Capacity),
			Enabled:  in.Enabled,
		},
	})

	if err != nil {
		return nil, err
	}

	res := &engine.CreateSkillAgentResponse{
		Items: make([]*engine.SkillAgentItem, 0, len(list)),
	}

	for _, v := range list {
		res.Items = append(res.Items, transformSkillAgentItem(v))
	}

	return res, nil

}

func (api *skill) SearchSkillAgent(ctx context.Context, in *engine.SearchSkillAgentRequest) (*engine.ListSkillAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentSkill
	var endList bool
	var existsDisabled bool
	req := &model.SearchAgentSkillList{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		SearchAgentSkill: model.SearchAgentSkill{
			Ids:      in.Id,
			SkillIds: nil,
			AgentIds: in.AgentId,
		},
	}
	list, endList, existsDisabled, err = api.ctrl.GetAgentsSkillBySkill(ctx, session, in.SkillId, req)
	if err != nil {
		return nil, err
	}

	res := &engine.ListSkillAgent{
		Next:  !endList,
		Items: make([]*engine.SkillAgentItem, 0, len(list)),
		Aggs: &engine.ListSkillAgent_ListSkillAgg{
			Enabled: !existsDisabled,
		},
	}

	for _, v := range list {
		res.Items = append(res.Items, transformSkillAgentItem(v))
	}

	return res, nil
}

func (api *skill) PatchSkillAgent(ctx context.Context, in *engine.PatchSkillAgentRequest) (*engine.PatchSkillAgentResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentSkill
	patch := model.AgentSkillPatch{
		UpdatedAt: model.GetMillis(),
		UpdatedBy: model.Lookup{
			Id: int(session.UserId),
		},
	}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "skill.id":
			patch.Skill = &model.Lookup{Id: int(in.GetSkill().GetId())}
		case "capacity":
			patch.Capacity = model.NewInt(int(in.Capacity))
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	list, err = api.ctrl.PatchAgentsSkillBySkill(ctx, session, in.SkillId, model.SearchAgentSkill{
		Ids:      in.Id,
		SkillIds: nil,
		AgentIds: in.AgentId,
	}, patch)

	if err != nil {
		return nil, err
	}

	res := &engine.PatchSkillAgentResponse{
		Items: make([]*engine.SkillAgentItem, 0, len(list)),
	}

	for _, v := range list {
		res.Items = append(res.Items, transformSkillAgentItem(v))
	}

	return res, nil
}

func (api *skill) DeleteSkillAgent(ctx context.Context, in *engine.DeleteSkillAgentRequest) (*engine.DeleteSkillAgentResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentSkill

	list, err = api.ctrl.DeleteAgentsSkill(ctx, session, in.SkillId, model.SearchAgentSkill{
		Ids:      in.Id,
		SkillIds: nil,
		AgentIds: in.AgentId,
	})
	if err != nil {
		return nil, err
	}

	res := &engine.DeleteSkillAgentResponse{
		Items: make([]*engine.SkillAgentItem, 0, len(list)),
	}

	for _, v := range list {
		res.Items = append(res.Items, transformSkillAgentItem(v))
	}

	return res, nil
}

func transformSkillAgentItem(src *model.AgentSkill) *engine.SkillAgentItem {
	return &engine.SkillAgentItem{
		Id:       src.Id,
		Skill:    GetProtoLookup(src.Skill),
		Capacity: int32(src.Capacity),
		Enabled:  src.Enabled,
		Agent:    GetProtoLookup(src.Agent),
		Team:     GetProtoLookup(src.Team),
	}
}

func transformSkill(src *model.Skill) *engine.Skill {
	res := &engine.Skill{
		Id:          src.Id,
		Name:        src.Name,
		Description: src.Description,
	}

	if src.Agents != nil {
		res.Agents = *src.Agents
	}

	return res
}
