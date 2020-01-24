package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type resourceTeam struct {
	app *app.App
}

func NewResourceTeamApi(app *app.App) *resourceTeam {
	return &resourceTeam{app: app}
}

func (api *resourceTeam) CreateResourceTeamAgent(ctx context.Context, in *engine.CreateResourceTeamAgentRequest) (*engine.ResourceTeamAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	teamResource := &model.ResourceInTeam{
		TeamId: in.TeamId,
		Agent:  GetLookup(in.Agent),
		Bucket: GetLookup(in.Bucket),
		Lvl:    int(in.Lvl),
	}

	if err = teamResource.IsValid(); err != nil {
		return nil, err
	}

	teamResource, err = api.app.CreateResourceTeamAgent(teamResource)
	if err != nil {
		return nil, err
	}

	return transformResourceTeamAgent(teamResource), nil
}

func (api *resourceTeam) ReadResourceTeamAgent(ctx context.Context, in *engine.ReadResourceTeamAgentRequest) (*engine.ResourceTeamAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var resource *model.ResourceInTeam
	if resource, err = api.app.GetResourceTeamAgent(session.Domain(in.DomainId), in.GetTeamId(), in.GetId()); err != nil {
		return nil, err
	}
	return transformResourceTeamAgent(resource), nil
}

func (api *resourceTeam) SearchResourceTeamAgent(ctx context.Context, in *engine.SearchResourceTeamAgentRequest) (*engine.ListResourceTeamAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.ResourceInTeam
	list, err = api.app.GetResourceTeamAgentPage(session.Domain(in.DomainId), in.GetTeamId(), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.ResourceTeamAgent, 0, len(list))
	for _, v := range list {
		items = append(items, transformResourceTeamAgent(v))
	}
	return &engine.ListResourceTeamAgent{
		Items: items,
	}, nil
}

func (api *resourceTeam) UpdateResourceTeamAgent(ctx context.Context, in *engine.UpdateResourceTeamAgentRequest) (*engine.ResourceTeamAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.UpdateResourceTeamAgent(session.Domain(in.GetDomainId()), &model.ResourceInTeam{
		Id:     in.Id,
		TeamId: in.TeamId,
		Agent:  GetLookup(in.Agent),
		Bucket: GetLookup(in.Bucket),
		Lvl:    int(in.Lvl),
	})

	if err != nil {
		return nil, err
	}

	return transformResourceTeamAgent(resource), nil
}

func (api *resourceTeam) DeleteResourceTeamAgent(ctx context.Context, in *engine.DeleteResourceTeamAgentRequest) (*engine.ResourceTeamAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.RemoveResourceTeamAgent(session.Domain(in.GetDomainId()), in.GetTeamId(), in.GetId())

	if err != nil {
		return nil, err
	}

	return transformResourceTeamAgent(resource), nil
}

//Skill
func (api *resourceTeam) CreateResourceTeamSkill(ctx context.Context, in *engine.CreateResourceTeamSkillRequest) (*engine.ResourceTeamSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	teamResource := &model.ResourceInTeam{
		TeamId:      in.TeamId,
		Skill:       GetLookup(in.Skill),
		Bucket:      GetLookup(in.Bucket),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
	}

	if err = teamResource.IsValid(); err != nil {
		return nil, err
	}

	teamResource, err = api.app.CreateResourceTeamSkill(teamResource)
	if err != nil {
		return nil, err
	}

	return transformResourceTeamSkill(teamResource), nil
}

func (api *resourceTeam) ReadResourceTeamSkill(ctx context.Context, in *engine.ReadResourceTeamSkillRequest) (*engine.ResourceTeamSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var resource *model.ResourceInTeam
	if resource, err = api.app.GetResourceTeamAgent(session.Domain(in.DomainId), in.GetTeamId(), in.GetId()); err != nil {
		return nil, err
	}
	return transformResourceTeamSkill(resource), nil
}

func (api *resourceTeam) SearchResourceTeamSkill(ctx context.Context, in *engine.SearchResourceTeamSkillRequest) (*engine.ListResourceTeamSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.ResourceInTeam
	list, err = api.app.GetResourceTeamSkillPage(session.Domain(in.DomainId), in.GetTeamId(), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.ResourceTeamSkill, 0, len(list))
	for _, v := range list {
		items = append(items, transformResourceTeamSkill(v))
	}
	return &engine.ListResourceTeamSkill{
		Items: items,
	}, nil
}

func (api *resourceTeam) UpdateResourceTeamSkill(ctx context.Context, in *engine.UpdateResourceTeamSkillRequest) (*engine.ResourceTeamSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.UpdateResourceTeamSkill(session.Domain(in.GetDomainId()), &model.ResourceInTeam{
		Id:          in.Id,
		TeamId:      in.TeamId,
		Skill:       GetLookup(in.Skill),
		Bucket:      GetLookup(in.Bucket),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
	})

	if err != nil {
		return nil, err
	}

	return transformResourceTeamSkill(resource), nil
}

func (api *resourceTeam) DeleteResourceTeamSkill(ctx context.Context, in *engine.DeleteResourceTeamSkillRequest) (*engine.ResourceTeamSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetTeamId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.RemoveResourceTeamSkill(session.Domain(in.GetDomainId()), in.GetTeamId(), in.GetId())

	if err != nil {
		return nil, err
	}

	return transformResourceTeamSkill(resource), nil
}

func transformResourceTeamAgent(src *model.ResourceInTeam) *engine.ResourceTeamAgent {
	return &engine.ResourceTeamAgent{
		Id:     src.Id,
		TeamId: src.TeamId,
		Agent:  GetProtoLookup(src.Agent),
		Bucket: GetProtoLookup(src.Bucket),
		Lvl:    int32(src.Lvl),
	}
}

func transformResourceTeamSkill(src *model.ResourceInTeam) *engine.ResourceTeamSkill {
	return &engine.ResourceTeamSkill{
		Id:          src.Id,
		TeamId:      src.TeamId,
		Skill:       GetProtoLookup(src.Skill),
		Bucket:      GetProtoLookup(src.Bucket),
		Lvl:         int32(src.Lvl),
		MinCapacity: int32(src.MinCapacity),
		MaxCapacity: int32(src.MaxCapacity),
	}
}
