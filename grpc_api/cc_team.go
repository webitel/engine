package grpc_api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/webitel/engine/gen/engine"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

type agentTeam struct {
	app *app.App
	engine.UnsafeAgentTeamServiceServer
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
			CreatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
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
		Admin:             GetLookups(in.Admin),
		InviteChatTimeout: int16(in.InviteChatTimeout),
		TaskAcceptTimeout: int16(in.TaskAcceptTimeout),
	}

	if in.ForecastCalculation != nil {
		if !session.HasLicense(auth_manager.LicenseWFM) {
			return nil, model.NewForbiddenError("grpc_api.cc_team.forecast_calculation", fmt.Sprintf("license %s required to proceed with forecast calculation", auth_manager.LicenseWFM))
		}

		team.ForecastCalculation = GetLookup(in.ForecastCalculation)
	}

	err = team.IsValid()
	if err != nil {
		return nil, err
	}

	team, err = api.app.CreateAgentTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	res := transformAgentTeam(team)
	api.app.AuditCreate(ctx, session, model.PERMISSION_SCOPE_CC_TEAM, strconv.FormatInt(res.Id, 10), res)

	return res, nil
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
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Sort:    in.Sort,
			Fields:  in.Fields,
		},
		Ids:      in.Id,
		Strategy: in.Strategy,
		AdminIds: in.AdminId,
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.GetAgentTeamsPageByGroups(ctx, session.Domain(0), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetAgentTeamsPage(ctx, session.Domain(0), req)
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	team, err = api.app.GetAgentTeamById(ctx, session.Domain(in.DomainId), in.Id)

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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	team := &model.AgentTeam{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:                in.Name,
		Description:         in.Description,
		Strategy:            in.Strategy,
		MaxNoAnswer:         int16(in.MaxNoAnswer),
		WrapUpTime:          int16(in.WrapUpTime),
		NoAnswerDelayTime:   int16(in.NoAnswerDelayTime),
		CallTimeout:         int16(in.CallTimeout),
		Admin:               GetLookups(in.Admin),
		InviteChatTimeout:   int16(in.InviteChatTimeout),
		TaskAcceptTimeout:   int16(in.TaskAcceptTimeout),
		ForecastCalculation: GetLookup(in.ForecastCalculation),
	}

	out, err := api.app.UpdateAgentTeam(ctx, session.Domain(in.GetDomainId()), team)
	if err != nil {
		return nil, err
	}

	res := transformAgentTeam(out)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_TEAM, strconv.FormatInt(res.Id, 10), res)

	return res, nil
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = api.app.AgentTeamCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var team *model.AgentTeam
	team, err = api.app.RemoveAgentTeam(ctx, session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	res := transformAgentTeam(team)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_TEAM, strconv.FormatInt(res.Id, 10), res)

	return res, nil
}

func transformAgentTeam(src *model.AgentTeam) *engine.AgentTeam {
	return &engine.AgentTeam{
		Id:                  src.Id,
		DomainId:            src.DomainId,
		Name:                src.Name,
		Description:         src.Description,
		Strategy:            src.Strategy,
		MaxNoAnswer:         int32(src.MaxNoAnswer),
		WrapUpTime:          int32(src.WrapUpTime),
		NoAnswerDelayTime:   int32(src.NoAnswerDelayTime),
		CallTimeout:         int32(src.CallTimeout),
		Admin:               GetProtoLookups(src.Admin),
		InviteChatTimeout:   int32(src.InviteChatTimeout),
		TaskAcceptTimeout:   int32(src.TaskAcceptTimeout),
		ForecastCalculation: GetProtoLookup(src.ForecastCalculation),
	}
}
