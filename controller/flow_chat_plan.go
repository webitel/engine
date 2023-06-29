package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchChatPlan(ctx context.Context, session *auth_manager.Session, search *model.SearchChatPlan) ([]*model.ChatPlan, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetChatPlanPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) CreateChatPlan(ctx context.Context, session *auth_manager.Session, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := plan.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateChatPlan(ctx, session.Domain(0), plan)
}

func (c *Controller) GetChatPlan(ctx context.Context, session *auth_manager.Session, id int32) (*model.ChatPlan, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetChatPlan(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateChatPlan(ctx context.Context, session *auth_manager.Session, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := plan.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateChatPlan(ctx, session.DomainId, plan)
}

func (c *Controller) PatchChatPlan(ctx context.Context, session *auth_manager.Session, id int32, patch *model.PatchChatPlan) (*model.ChatPlan, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.PatchChatPlan(ctx, session.DomainId, id, patch)
}

func (c *Controller) DeleteChatPlan(ctx context.Context, session *auth_manager.Session, id int32) (*model.ChatPlan, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_CHAT_PLAN)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveChatPlan(ctx, session.Domain(0), id)
}
