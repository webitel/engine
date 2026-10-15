package grpc_api

import (
	"context"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type skillGroup struct {
	*API
	engine.UnsafeSkillGroupServiceServer
}

func NewSkillGroupApi(api *API) *skillGroup {
	return &skillGroup{API: api}
}

func (api *skillGroup) CreateSkillGroup(ctx context.Context, in *engine.CreateSkillGroupRequest) (*engine.SkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	group := &model.SkillGroup{
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	}

	group, err = api.ctrl.CreateSkillGroup(ctx, session, group)
	if err != nil {
		return nil, err
	}

	return transformSkillGroup(group), nil
}

func (api *skillGroup) SearchSkillGroup(ctx context.Context, in *engine.SearchSkillGroupRequest) (*engine.ListSkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchSkillGroup{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err := api.ctrl.SearchSkillGroup(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.SkillGroup, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkillGroup(v))
	}
	return &engine.ListSkillGroup{Next: !endList, Items: items}, nil
}

func (api *skillGroup) ReadSkillGroup(ctx context.Context, in *engine.ReadSkillGroupRequest) (*engine.SkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	group, err := api.ctrl.ReadSkillGroup(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}
	return transformSkillGroup(group), nil
}

func (api *skillGroup) UpdateSkillGroup(ctx context.Context, in *engine.UpdateSkillGroupRequest) (*engine.SkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	group, err := api.ctrl.UpdateSkillGroup(ctx, session, &model.SkillGroup{
		Id:          in.Id,
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	})
	if err != nil {
		return nil, err
	}
	return transformSkillGroup(group), nil
}

func (api *skillGroup) DeleteSkillGroup(ctx context.Context, in *engine.DeleteSkillGroupRequest) (*engine.SkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	group, err := api.ctrl.DeleteSkillGroup(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}
	return transformSkillGroup(group), nil
}

func (api *skillGroup) SearchSkillGroupSkill(ctx context.Context, in *engine.SearchSkillGroupSkillRequest) (*engine.ListSkillGroupSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchSkillGroupSkill{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		SkillGroupId: in.SkillGroupId,
		Ids:          in.Id,
	}

	list, endList, err := api.ctrl.SearchSkillGroupSkills(ctx, session, req)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupSkill{Next: !endList, Items: transformSkillGroupSkillItems(list)}, nil
}

func (api *skillGroup) CreateSkillGroupSkill(ctx context.Context, in *engine.CreateSkillGroupSkillRequest) (*engine.ListSkillGroupSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skills := make([]*model.SkillGroupSkill, 0, len(in.Items))
	for _, v := range in.Items {
		skills = append(skills, &model.SkillGroupSkill{
			Skill:    &model.Lookup{Id: int(v.GetSkill().GetId())},
			Capacity: capacityToInt(v.GetCapacity()),
		})
	}

	list, err := api.ctrl.CreateSkillGroupSkills(ctx, session, in.SkillGroupId, skills)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupSkill{Items: transformSkillGroupSkillItems(list)}, nil
}

func (api *skillGroup) UpdateSkillGroupSkill(ctx context.Context, in *engine.UpdateSkillGroupSkillRequest) (*engine.SkillGroupSkillItem, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	item, err := api.ctrl.UpdateSkillGroupSkill(ctx, session, in.SkillGroupId, &model.SkillGroupSkill{
		Id:       in.Id,
		Skill:    &model.Lookup{Id: int(in.GetSkill().GetId())},
		Capacity: capacityToInt(in.GetCapacity()),
	})
	if err != nil {
		return nil, err
	}
	return transformSkillGroupSkillItem(item), nil
}

func (api *skillGroup) DeleteSkillGroupSkill(ctx context.Context, in *engine.DeleteSkillGroupSkillRequest) (*engine.ListSkillGroupSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.DeleteSkillGroupSkills(ctx, session, in.SkillGroupId, in.Id, in.SkillId)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupSkill{Items: transformSkillGroupSkillItems(list)}, nil
}

func (api *skillGroup) SearchSkillGroupAgent(ctx context.Context, in *engine.SearchSkillGroupAgentRequest) (*engine.ListSkillGroupAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchSkillGroupAgent{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		SkillGroupId: in.SkillGroupId,
		Ids:          in.Id,
		AgentIds:     in.AgentId,
	}

	list, endList, err := api.ctrl.SearchSkillGroupAgents(ctx, session, req)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupAgent{Next: !endList, Items: transformSkillGroupAgentItems(list)}, nil
}

func (api *skillGroup) CreateSkillGroupAgent(ctx context.Context, in *engine.CreateSkillGroupAgentRequest) (*engine.ListSkillGroupAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.CreateSkillGroupAgents(ctx, session, []int64{in.SkillGroupId}, LookupsIds(in.Agent))
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupAgent{Items: transformSkillGroupAgentItems(list)}, nil
}

func (api *skillGroup) AssignSkillGroupAgents(ctx context.Context, in *engine.AssignSkillGroupAgentsRequest) (*engine.ListSkillGroupAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.CreateSkillGroupAgents(ctx, session, in.SkillGroupId, in.AgentId)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupAgent{Items: transformSkillGroupAgentItems(list)}, nil
}

func (api *skillGroup) DeleteSkillGroupAgent(ctx context.Context, in *engine.DeleteSkillGroupAgentRequest) (*engine.ListSkillGroupAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.DeleteSkillGroupAgents(ctx, session, in.SkillGroupId, in.Id, in.AgentId)
	if err != nil {
		return nil, err
	}
	return &engine.ListSkillGroupAgent{Items: transformSkillGroupAgentItems(list)}, nil
}

func (api *skillGroup) SearchAgentSkillGroup(ctx context.Context, in *engine.SearchAgentSkillGroupRequest) (*engine.ListAgentSkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.SearchAgentSkillGroup{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		AgentId: in.AgentId,
		Ids:     in.Id,
	}

	list, endList, err := api.ctrl.SearchAgentSkillGroups(ctx, session, req)
	if err != nil {
		return nil, err
	}
	return &engine.ListAgentSkillGroup{Next: !endList, Items: transformAgentSkillGroupItems(list)}, nil
}

func (api *skillGroup) CreateAgentSkillGroup(ctx context.Context, in *engine.CreateAgentSkillGroupRequest) (*engine.ListAgentSkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.CreateAgentSkillGroups(ctx, session, in.AgentId, in.SkillGroupId)
	if err != nil {
		return nil, err
	}
	return &engine.ListAgentSkillGroup{Items: transformAgentSkillGroupItems(list)}, nil
}

func (api *skillGroup) DeleteAgentSkillGroup(ctx context.Context, in *engine.DeleteAgentSkillGroupRequest) (*engine.ListAgentSkillGroup, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := api.ctrl.DeleteAgentSkillGroups(ctx, session, in.AgentId, in.Id, in.SkillGroupId)
	if err != nil {
		return nil, err
	}
	return &engine.ListAgentSkillGroup{Items: transformAgentSkillGroupItems(list)}, nil
}

func transformSkillGroup(src *model.SkillGroup) *engine.SkillGroup {
	res := &engine.SkillGroup{
		Id:          src.Id,
		DomainId:    src.DomainId,
		Name:        src.Name,
		Description: src.Description,
		Skills:      GetProtoLookups(src.Skills),
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		CreatedAt:   model.TimeToInt64(src.CreatedAt),
		UpdatedAt:   model.TimeToInt64(src.UpdatedAt),
	}
	if src.TotalAgents != nil {
		res.TotalAgents = *src.TotalAgents
	}
	if src.ActiveAgents != nil {
		res.ActiveAgents = *src.ActiveAgents
	}
	return res
}

func transformSkillGroupSkillItem(src *model.SkillGroupSkill) *engine.SkillGroupSkillItem {
	res := &engine.SkillGroupSkillItem{
		Id:    src.Id,
		Skill: GetProtoLookup(src.Skill),
	}
	if src.Capacity != nil {
		res.Capacity = &wrappers.Int32Value{Value: int32(*src.Capacity)}
	}
	return res
}

func transformSkillGroupSkillItems(list []*model.SkillGroupSkill) []*engine.SkillGroupSkillItem {
	items := make([]*engine.SkillGroupSkillItem, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkillGroupSkillItem(v))
	}
	return items
}

func transformSkillGroupAgentItem(src *model.SkillGroupAgent) *engine.SkillGroupAgentItem {
	return &engine.SkillGroupAgentItem{
		Id:    src.Id,
		Agent: GetProtoLookup(src.Agent),
		Team:  GetProtoLookup(src.Team),
	}
}

func transformSkillGroupAgentItems(list []*model.SkillGroupAgent) []*engine.SkillGroupAgentItem {
	items := make([]*engine.SkillGroupAgentItem, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkillGroupAgentItem(v))
	}
	return items
}

func transformAgentSkillGroupItem(src *model.AgentSkillGroup) *engine.AgentSkillGroupItem {
	return &engine.AgentSkillGroupItem{
		Id:          src.Id,
		SkillGroup:  GetProtoLookup(src.SkillGroup),
		Description: src.Description,
		Skills:      GetProtoLookups(src.Skills),
	}
}

func transformAgentSkillGroupItems(list []*model.AgentSkillGroup) []*engine.AgentSkillGroupItem {
	items := make([]*engine.AgentSkillGroupItem, 0, len(list))
	for _, v := range list {
		items = append(items, transformAgentSkillGroupItem(v))
	}
	return items
}

func capacityToInt(v *wrappers.Int32Value) *int {
	if v == nil {
		return nil
	}
	return model.NewInt(int(v.GetValue()))
}
