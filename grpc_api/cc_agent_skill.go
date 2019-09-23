package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type agentSkill struct {
	app *app.App
}

func NewAgentSkillApi(app *app.App) *agentSkill {
	return &agentSkill{app: app}
}

func (api *agentSkill) Create(ctx context.Context, in *engine.AgentSkill) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgent().GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgent().GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	agentSkill := &model.AgentSkill{
		DomainRecord: model.DomainRecord{
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Agent: model.Lookup{
			Id: int(in.GetAgent().GetId()),
		},
		Skill: model.Lookup{
			Id: int(in.GetSkill().GetId()),
		},
		Capacity: int(in.Capacity),
	}

	err = agentSkill.IsValid()
	if err != nil {
		return nil, err
	}

	agentSkill, err = api.app.CreateAgentSkill(agentSkill)
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) List(ctx context.Context, in *engine.ListForItemRequest) (*engine.ListAgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetItemId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetItemId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentSkill
	list, err = api.app.GetAgentsSkillPage(session.Domain(int64(in.DomainId)), in.GetItemId(), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentSkillItem, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.AgentSkillItem{
			Id: v.Id,
			Skill: &engine.Lookup{
				Id:   int64(v.Skill.Id),
				Name: v.Skill.Name,
			},
			Capacity: int32(v.Capacity),
		})
	}
	return &engine.ListAgentSkill{
		Items: items,
	}, nil
}

func (api *agentSkill) Get(ctx context.Context, in *engine.AgentSkillItemReqeust) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgentId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var agentSkill *model.AgentSkill
	agentSkill, err = api.app.GetAgentsSkillById(session.Domain(in.GetDomainId()), in.GetAgentId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) Update(ctx context.Context, in *engine.AgentSkill) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgent().GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgent().GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agentSkill *model.AgentSkill

	agentSkill, err = api.app.UpdateAgentsSkill(&model.AgentSkill{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Agent: model.Lookup{
			Id: int(in.GetAgent().GetId()),
		},
		Skill: model.Lookup{
			Id: int(in.GetSkill().GetId()),
		},
		Capacity: int(in.Capacity),
	})

	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func (api *agentSkill) Remove(ctx context.Context, in *engine.AgentSkill) (*engine.AgentSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgent().GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgent().GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agentSkill *model.AgentSkill

	agentSkill, err = api.app.RemoveAgentSkill(session.Domain(in.GetDomainId()), in.GetAgent().GetId(), in.Id)
	if err != nil {
		return nil, err
	}

	return transformAgentSkill(agentSkill), nil
}

func transformAgentSkill(src *model.AgentSkill) *engine.AgentSkill {
	return &engine.AgentSkill{
		Id:        src.Id,
		CreatedAt: src.CreatedAt,
		CreatedBy: &engine.Lookup{
			Id:   int64(src.CreatedBy.Id),
			Name: src.CreatedBy.Name,
		},
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: &engine.Lookup{
			Id:   int64(src.UpdatedBy.Id),
			Name: src.UpdatedBy.Name,
		},
		Agent: &engine.Lookup{
			Id:   int64(src.Agent.Id),
			Name: src.Agent.Name,
		},
		Skill: &engine.Lookup{
			Id:   int64(src.Skill.Id),
			Name: src.Skill.Name,
		},
		Capacity: int32(src.Capacity),
	}
}
