package controller

import (
	"context"
	"time"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateCall(ctx context.Context, session *auth_manager.Session, req *model.OutboundCallRequest, variables map[string]string) (string, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanCreate() {
		return "", c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	req.CreatedAt = model.GetMillis()
	req.CreatedById = session.UserId

	if req.From == nil {
		req.From = &model.EndpointRequest{
			UserId: model.NewInt64(session.UserId),
		}
	}

	return c.app.CreateOutboundCall(ctx, session.DomainId, req, variables)
}

func (c *Controller) RedialCall(ctx context.Context, session *auth_manager.Session, callId string) (string, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanCreate() {
		return "", c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	return c.app.RedialCall(ctx, session.DomainId, session.UserId, callId)
}

func (c *Controller) SearchCall(ctx context.Context, session *auth_manager.Session, search *model.SearchCall) ([]*model.Call, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetActiveCallPageByGroups(ctx, session.DomainId, session.UserId, session.RoleIds, search)
	}

	return c.app.GetActiveCallPage(ctx, session.DomainId, search)
}

func (c *Controller) UserActiveCall(ctx context.Context, session *auth_manager.Session) ([]*model.Call, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetUserActiveCalls(ctx, session.DomainId, session.UserId)
}

func (c *Controller) SearchHistoryCall(ctx context.Context, session *auth_manager.Session, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetHistoryCallPageByGroups(ctx, session.DomainId, session.UserId, session.RoleIds, search)
	}

	return c.app.GetHistoryCallPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) AggregateHistoryCall(ctx context.Context, session *auth_manager.Session, aggs *model.CallAggregate) ([]*model.AggregateResult, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetAggregateHistoryCallPage(ctx, session.Domain(aggs.DomainId), aggs)
}

func (c *Controller) GetCall(ctx context.Context, session *auth_manager.Session, domainId int64, id string) (*model.Call, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCall(ctx, session.Domain(domainId), id)
}

func (c *Controller) HangupCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.HangupCall) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanDelete() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.HangupCall(ctx, session.Domain(domainId), req)
}

func (c *Controller) HoldCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.UserCallRequest) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.HoldCall(ctx, session.Domain(domainId), req)
}

func (c *Controller) UnHoldCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.UserCallRequest) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.UnHoldCall(ctx, session.Domain(domainId), req)
}

func (c *Controller) DtmfCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.DtmfCall) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.DtmfCall(ctx, session.Domain(domainId), req)
}

// BlindTransferCall todo deprecated
func (c *Controller) BlindTransferCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.BlindTransferCall) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.BlindTransferCall(ctx, session.Domain(domainId), req)
}

func (c *Controller) BlindTransferCallExt(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.BlindTransferCall) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.BlindTransferCallExt(ctx, session.Domain(domainId), req)
}

func (c *Controller) EavesdropCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.EavesdropCall, variables map[string]string) (string, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return "", c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.EavesdropCall(ctx, session.Domain(domainId), session.UserId, req, variables)
}

func (c *Controller) EavesdropStateCall(ctx context.Context, session *auth_manager.Session, domainId int64, req *model.EavesdropCall) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.EavesdropCallState(ctx, session.Domain(domainId), session.UserId, req)
}

func (c *Controller) CreateCallAnnotation(ctx context.Context, session *auth_manager.Session, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	annotation.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	annotation.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	annotation.CreatedAt = time.Now()
	annotation.UpdatedAt = annotation.CreatedAt

	if err := annotation.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateCallAnnotation(ctx, session.DomainId, annotation)
}

func (c *Controller) UpdateCallAnnotation(ctx context.Context, session *auth_manager.Session, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	annotation.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	annotation.UpdatedAt = time.Now()

	if err := annotation.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateCallAnnotation(ctx, session.DomainId, annotation)
}

func (c *Controller) DeleteCallAnnotation(ctx context.Context, session *auth_manager.Session, id int64, callId string) (*model.CallAnnotation, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.DeleteCallAnnotation(ctx, session.DomainId, id, callId)
}

func (c *Controller) ConfirmPushCall(session *auth_manager.Session, callId string) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.ConfirmPushCall(session.DomainId, callId)
}

func (c *Controller) SetCallVariables(ctx context.Context, session *auth_manager.Session, callId string, vars map[string]string) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.SetCallVariables(ctx, session.Domain(0), callId, vars)
}

func (c *Controller) UpdateCallHistory(ctx context.Context, session *auth_manager.Session, id string, upd *model.HistoryCallPatch) (*model.HistoryCall, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	// TODO RBAC ?

	return c.app.UpdateHistoryCall(ctx, session.Domain(0), id, upd)
}

func (c *Controller) SetContactCall(ctx context.Context, session *auth_manager.Session, id string, contactId int64) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	// TODO RBAC ?

	return c.app.SetCallContactId(ctx, session.Domain(0), session.UserId, id, contactId)
}
