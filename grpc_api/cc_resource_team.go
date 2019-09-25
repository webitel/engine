package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type resourceTeam struct {
	app *app.App
}

func NewResourceTeamApi(app *app.App) *resourceTeam {
	return &resourceTeam{app: app}
}

func (api *resourceTeam) Create(ctx context.Context, in *engine.ResourceTeam) (*engine.ResourceTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	teamResource := &model.ResourceInTeam{
		TeamId:      in.TeamId,
		Agent:       GetLookup(in.Agent),
		Skill:       GetLookup(in.Skill),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
	}

	if err = teamResource.IsValid(); err != nil {
		return nil, err
	}

	teamResource, err = api.app.CreateResourceTeam(teamResource)
	if err != nil {
		return nil, err
	}

	return transformResourceTeam(teamResource), nil
}

func (api *resourceTeam) Get(ctx context.Context, in *engine.ResourceTeamItemReqeust) (*engine.ResourceTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var resource *model.ResourceInTeam
	if resource, err = api.app.GetResourceTeam(session.Domain(in.DomainId), in.GetTeamId(), in.GetId()); err != nil {
		return nil, err
	}
	return transformResourceTeam(resource), nil
}

func (api *resourceTeam) List(ctx context.Context, in *engine.ListForItemRequest) (*engine.ListResourceTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetItemId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetItemId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.ResourceInTeam
	list, err = api.app.GetResourceTeamPage(session.Domain(int64(in.DomainId)), in.GetItemId(), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.ResourceTeam, 0, len(list))
	for _, v := range list {
		items = append(items, transformResourceTeam(v))
	}
	return &engine.ListResourceTeam{
		Items: items,
	}, nil
}

func (api *resourceTeam) Update(ctx context.Context, in *engine.ResourceTeam) (*engine.ResourceTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.UpdateResourceTeam(session.Domain(in.GetDomainId()), &model.ResourceInTeam{
		Id:          in.Id,
		TeamId:      in.TeamId,
		Agent:       GetLookup(in.Agent),
		Skill:       GetLookup(in.Skill),
		Lvl:         int(in.Lvl),
		MinCapacity: int(in.MinCapacity),
		MaxCapacity: int(in.MaxCapacity),
	})

	if err != nil {
		return nil, err
	}

	return transformResourceTeam(resource), nil
}

func (api *resourceTeam) Remove(ctx context.Context, in *engine.ResourceTeamItemReqeust) (*engine.ResourceTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetTeamId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var resource *model.ResourceInTeam

	resource, err = api.app.RemoveResourceTeam(session.Domain(in.GetDomainId()), in.GetTeamId(), in.GetId())

	if err != nil {
		return nil, err
	}

	return transformResourceTeam(resource), nil
}

func transformResourceTeam(src *model.ResourceInTeam) *engine.ResourceTeam {
	return &engine.ResourceTeam{
		Id:          src.Id,
		TeamId:      src.TeamId,
		Agent:       GetProtoLookup(src.Agent),
		Skill:       GetProtoLookup(src.Skill),
		Lvl:         int32(src.Lvl),
		MinCapacity: int32(src.MinCapacity),
		MaxCapacity: int32(src.MaxCapacity),
	}
}
