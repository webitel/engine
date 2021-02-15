package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type agentTeam struct {
	app *app.App
}

func NewAgentTeamApi(app *app.App) *agentTeam {
	return &agentTeam{app: app}
}

func (api *agentTeam) CreateAgentTeam(ctx context.Context, in *engine.CreateAgentTeamRequest) (*engine.AgentTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	team := &model.AgentTeam{
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
		Name:              in.Name,
		Description:       in.Description,
		Strategy:          in.Strategy,
		MaxNoAnswer:       int16(in.MaxNoAnswer),
		WrapUpTime:        int16(in.WrapUpTime),
		NoAnswerDelayTime: int16(in.NoAnswerDelayTime),
		CallTimeout:       int16(in.CallTimeout),
		Administrator:     GetLookup(in.Administrator),
	}

	err = team.IsValid()
	if err != nil {
		return nil, err
	}

	team, err = api.app.CreateAgentTeam(team)
	if err != nil {
		return nil, err
	}

	return transformAgentTeam(team), nil
}

func (api *agentTeam) SearchAgentTeam(ctx context.Context, in *engine.SearchAgentTeamRequest) (*engine.ListAgentTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.AgentTeam
	var endList bool
	req := &model.SearchAgentTeam{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.Sort,
			Fields:   in.Fields,
		},
		Ids:              in.Id,
		Strategy:         in.Strategy,
		AdministratorIds: in.AdministratorId,
	}

	if permission.Rbac {
		list, endList, err = api.app.GetAgentTeamsPageByGroups(session.Domain(in.DomainId), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetAgentTeamsPage(session.Domain(in.DomainId), req)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentTeam, 0, len(list))
	for _, v := range list {
		items = append(items, transformAgentTeam(v))
	}

	return &engine.ListAgentTeam{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agentTeam) ReadAgentTeam(ctx context.Context, in *engine.ReadAgentTeamRequest) (*engine.AgentTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var team *model.AgentTeam

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	team, err = api.app.GetAgentTeamById(session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return transformAgentTeam(team), nil
}

func (api *agentTeam) UpdateAgentTeam(ctx context.Context, in *engine.UpdateAgentTeamRequest) (*engine.AgentTeam, error) {
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
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var team *model.AgentTeam

	team, err = api.app.UpdateAgentTeam(session.Domain(in.GetDomainId()), &model.AgentTeam{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:              in.Name,
		Description:       in.Description,
		Strategy:          in.Strategy,
		MaxNoAnswer:       int16(in.MaxNoAnswer),
		WrapUpTime:        int16(in.WrapUpTime),
		NoAnswerDelayTime: int16(in.NoAnswerDelayTime),
		CallTimeout:       int16(in.CallTimeout),
		Administrator:     GetLookup(in.Administrator),
	})

	if err != nil {
		return nil, err
	}

	return transformAgentTeam(team), nil
}

func (api *agentTeam) DeleteAgentTeam(ctx context.Context, in *engine.DeleteAgentTeamRequest) (*engine.AgentTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var team *model.AgentTeam
	team, err = api.app.RemoveAgentTeam(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformAgentTeam(team), nil
}

func transformAgentTeam(src *model.AgentTeam) *engine.AgentTeam {
	return &engine.AgentTeam{
		Id:                src.Id,
		DomainId:          src.DomainId,
		Name:              src.Name,
		Description:       src.Description,
		Strategy:          src.Strategy,
		MaxNoAnswer:       int32(src.MaxNoAnswer),
		WrapUpTime:        int32(src.WrapUpTime),
		NoAnswerDelayTime: int32(src.NoAnswerDelayTime),
		CallTimeout:       int32(src.CallTimeout),
		Administrator:     GetProtoLookup(src.Administrator),
	}
}
