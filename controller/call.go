package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"time"
)

func (c *Controller) CreateCall(session *auth_manager.Session, req *model.OutboundCallRequest, variables map[string]string) (string, *model.AppError) {
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

	return c.app.CreateOutboundCall(session.DomainId, req, variables)
}

func (c *Controller) SearchCall(session *auth_manager.Session, search *model.SearchCall) ([]*model.Call, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetActiveCallPageByGroups(session.DomainId, session.UserId, session.RoleIds, search)
	}

	return c.app.GetActiveCallPage(session.DomainId, search)
}

func (c *Controller) UserActiveCall(session *auth_manager.Session) ([]*model.Call, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetUserActiveCalls(session.DomainId, session.UserId)
}

func (c *Controller) SearchHistoryCall(session *auth_manager.Session, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetHistoryCallPageByGroups(session.DomainId, session.UserId, session.RoleIds, search)
	}

	return c.app.GetHistoryCallPage(session.Domain(search.DomainId), search)
}

func (c *Controller) AggregateHistoryCall(session *auth_manager.Session, aggs *model.CallAggregate) ([]*model.AggregateResult, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetAggregateHistoryCallPage(session.Domain(aggs.DomainId), aggs)
}

func (c *Controller) GetCall(session *auth_manager.Session, domainId int64, id string) (*model.Call, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCall(session.Domain(domainId), id)
}

func (c *Controller) HangupCall(session *auth_manager.Session, domainId int64, req *model.HangupCall) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanDelete() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.HangupCall(session.Domain(domainId), req)
}

func (c *Controller) HoldCall(session *auth_manager.Session, domainId int64, req *model.UserCallRequest) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.HoldCall(session.Domain(domainId), req)
}

func (c *Controller) UnHoldCall(session *auth_manager.Session, domainId int64, req *model.UserCallRequest) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.UnHoldCall(session.Domain(domainId), req)
}

func (c *Controller) DtmfCall(session *auth_manager.Session, domainId int64, req *model.DtmfCall) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.DtmfCall(session.Domain(domainId), req)
}

func (c *Controller) BlindTransferCall(session *auth_manager.Session, domainId int64, req *model.BlindTransferCall) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.BlindTransferCall(session.Domain(domainId), req)
}
func (c *Controller) EavesdropCall(session *auth_manager.Session, domainId int64, req *model.EavesdropCall, variables map[string]string) (string, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return "", c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.EavesdropCall(session.Domain(domainId), session.UserId, req, variables)
}

func (c *Controller) CreateCallAnnotation(session *auth_manager.Session, annotation *model.CallAnnotation) (*model.CallAnnotation, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	annotation.CreatedBy = model.Lookup{
		Id: int(session.UserId),
	}
	annotation.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}
	annotation.CreatedAt = time.Now()
	annotation.UpdatedAt = annotation.CreatedAt

	if err := annotation.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateCallAnnotation(session.DomainId, annotation)
}

func (c *Controller) UpdateCallAnnotation(session *auth_manager.Session, annotation *model.CallAnnotation) (*model.CallAnnotation, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	annotation.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}
	annotation.UpdatedAt = time.Now()

	if err := annotation.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateCallAnnotation(session.DomainId, annotation)
}

func (c *Controller) DeleteCallAnnotation(session *auth_manager.Session, id int64, callId string) (*model.CallAnnotation, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALL)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.DeleteCallAnnotation(session.DomainId, id, callId)
}
