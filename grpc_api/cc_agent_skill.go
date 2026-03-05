package grpc_api

import (
	"context"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

type agentSkill struct {
	*API
	engine.UnsafeAgentSkillServiceServer
}

func NewAgentSkillApi(api *API) *agentSkill {
	return &agentSkill{API: api}
}

func (api *agentSkill) CreateAgentSkill(ctx context.Context, in *engine.CreateAgentSkillRequest) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	agentSkill := &model.AgentSkill{
		DomainRecord: model.DomainRecord{
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
		Agent: &model.Lookup{
			Id: int(in.GetAgentId()),
		},
		Skill: &model.Lookup{
			Id: int(in.GetSkill().GetId()),
		},
		AgentSkillProps: model.AgentSkillProps{
			Enabled: in.Enabled,
		},
	}

	if in.GetCapacity() != nil {
		agentSkill.AgentSkillProps.Capacity = model.NewInt(int(in.GetCapacity().GetValue()))
	}

	err = agentSkill.IsValid()
	if err != nil {
		return nil, err
	}

	agentSkill, err = api.app.CreateAgentSkill(ctx, agentSkill)
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) SearchAgentSkill(ctx context.Context, in *engine.SearchAgentSkillRequest) (*engine.ListAgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentSkill
	var endList bool
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
			SkillIds: in.SkillId,
			QScopes: []model.SearchAgentSkillQScope{model.SKILL},
		},
	}

	list, endList, err = api.app.GetAgentsSkillPage(ctx, session.Domain(0), in.GetAgentId(), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentSkillItem, 0, len(list))
	for _, v := range list {
		item := &engine.AgentSkillItem{
			Id:      v.Id,
			Skill:   GetProtoLookup(v.Skill),
			Enabled: v.Enabled,
		}

		if v.Capacity != nil {
			item.Capacity = &wrappers.Int32Value{
				Value: int32(*v.Capacity),
			}
		}
		items = append(items, item)
	}
	return &engine.ListAgentSkill{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agentSkill) ReadAgentSkill(ctx context.Context, in *engine.AgentSkillItemRequest) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var agentSkill *model.AgentSkill
	agentSkill, err = api.app.GetAgentsSkillById(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) UpdateAgentSkill(ctx context.Context, in *engine.UpdateAgentSkillRequest) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	agentSkill := &model.AgentSkill{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
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
		Agent: &model.Lookup{
			Id: int(in.GetAgentId()),
		},
		Skill: &model.Lookup{
			Id: int(in.GetSkill().GetId()),
		},
		AgentSkillProps: model.AgentSkillProps{
			Enabled: in.Enabled,
		},
	}

	if in.GetCapacity() != nil {
		agentSkill.AgentSkillProps.Capacity = model.NewInt(int(in.GetCapacity().GetValue()))
	}

	agentSkill, err = api.app.UpdateAgentsSkill(ctx, agentSkill)

	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) PatchAgentSkill(ctx context.Context, in *engine.PatchAgentSkillRequest) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agentSkill *model.AgentSkill
	patch := &model.AgentSkillPatch{
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
			patch.Capacity = model.NewInt(int(in.GetCapacity().GetValue()))
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	agentSkill, err = api.app.PatchAgentSkill(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), in.Id, patch)

	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) PatchAgentSkills(ctx context.Context, in *engine.PatchAgentSkillsRequest) (*engine.ListAgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	patch := model.AgentSkillPatch{
		UpdatedAt: model.GetMillis(),
		UpdatedBy: model.Lookup{
			Id: int(session.UserId),
		},
	}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "capacity":
			patch.Capacity = model.NewInt(int(in.GetCapacity().GetValue()))
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	var list []*model.AgentSkill
	list, err = api.app.UpdateAgentsSkills(ctx, session.Domain(0), in.AgentId, model.SearchAgentSkill{
		Ids:      in.Id,
		SkillIds: in.SkillId,
	}, patch)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentSkillItem, 0, len(list))
	for _, v := range list {
		item := &engine.AgentSkillItem{
			Id:      v.Id,
			Skill:   GetProtoLookup(v.Skill),
			Enabled: v.Enabled,
		}

		if v.Capacity != nil {
			item.Capacity = &wrappers.Int32Value{
				Value: int32(*v.Capacity),
			}
		}

		items = append(items, item)
	}

	return &engine.ListAgentSkill{
		Items: items,
	}, nil
}

func (api *agentSkill) DeleteAgentSkill(ctx context.Context, in *engine.DeleteAgentSkillRequest) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agentSkill *model.AgentSkill

	agentSkill, err = api.app.RemoveAgentSkill(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), in.Id)
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) DeleteAgentSkills(ctx context.Context, in *engine.DeleteAgentSkillsRequest) (*engine.ListAgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var list []*model.AgentSkill
	list, err = api.app.RemoveAgentSkills(ctx, session.Domain(0), in.AgentId, model.SearchAgentSkill{
		Ids:      in.Id,
		SkillIds: in.SkillId,
	})

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentSkillItem, 0, len(list))
	for _, v := range list {
		item := &engine.AgentSkillItem{
			Id:      v.Id,
			Skill:   GetProtoLookup(v.Skill),
			Enabled: v.Enabled,
		}

		if v.Capacity != nil {
			item.Capacity = &wrappers.Int32Value{
				Value: int32(*v.Capacity),
			}
		}

		items = append(items, item)
	}

	return &engine.ListAgentSkill{
		Items: items,
	}, nil

}

func (api *agentSkill) SearchLookupAgentNotExistsSkill(ctx context.Context, in *engine.SearchLookupAgentNotExistsSkillRequest) (*engine.ListSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.Skill
	var endList bool
	req := &model.SearchSkill{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
		},
		NotExistsAgent: &in.AgentId,
	}

	list, endList, err = api.ctrl.SearchSkill(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Skill, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.Skill{
			Id:   v.Id,
			Name: v.Name,
		})
	}
	return &engine.ListSkill{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agentSkill) CreateAgentSkills(ctx context.Context, in *engine.CreateAgentSkillsRequest) (*engine.CreateAgentSkillsResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	items := make([]*model.AgentSkill, 0, len(in.Items))

	for _, v := range in.Items {
		i := &model.AgentSkill{
			DomainRecord: model.DomainRecord{
				DomainId:  session.Domain(0),
				CreatedAt: model.GetMillis(),
				CreatedBy: &model.Lookup{
					Id: int(session.UserId),
				},
				UpdatedAt: model.GetMillis(),
				UpdatedBy: &model.Lookup{
					Id: int(session.UserId),
				},
			},
			Agent: &model.Lookup{
				Id: int(in.GetAgentId()),
			},
			Skill: &model.Lookup{
				Id: int(v.GetSkill().GetId()),
			},
			AgentSkillProps: model.AgentSkillProps{
				Enabled: v.Enabled,
			},
		}

		if v.GetCapacity() != nil {
			i.AgentSkillProps.Capacity = model.NewInt(int(v.GetCapacity().GetValue()))
		}

		if err = i.IsValid(); err != nil {
			return nil, err
		}

		items = append(items, i)
	}

	res := &engine.CreateAgentSkillsResponse{}
	res.Ids, err = api.app.CreateAgentSkills(ctx, session.Domain(0), in.AgentId, items)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func transformAgentSkill(src *model.AgentSkill) *engine.AgentSkill {
	s := &engine.AgentSkill{
		CreatedAt: src.CreatedAt,
		CreatedBy: GetProtoLookup(src.CreatedBy),
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: GetProtoLookup(src.UpdatedBy),
		Id:        src.Id,
		Agent:     GetProtoLookup(src.Agent),
		Skill:     GetProtoLookup(src.Skill),
		Enabled:   src.Enabled,
	}
	if src.Capacity != nil {
		s.Capacity = &wrappers.Int32Value{
			Value: int32(*src.Capacity),
		}
	}

	return s
}