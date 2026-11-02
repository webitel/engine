package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) SearchChatTag(ctx context.Context, session *auth_manager.Session, teamId int64, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchChatTag(ctx, session.Domain(0), teamId, search)
}

func (c *Controller) GetChatTag(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamChatTag, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetChatTag(ctx, session.Domain(0), teamId, id)
}

func (c *Controller) CreateChatTag(ctx context.Context, session *auth_manager.Session, teamId int64, tag *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	tag.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	tag.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	tag.CreatedAt = model.GetTime()
	tag.UpdatedAt = tag.CreatedAt

	if err := tag.IsValid(); err != nil {
		return nil, err
	}

	tag, err = c.app.CreateChatTag(ctx, session.Domain(0), teamId, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (c *Controller) UpdateChatTag(ctx context.Context, session *auth_manager.Session, teamId int64, tag *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	tag.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	tag.UpdatedAt = model.GetTime()

	if err := tag.IsValid(); err != nil {
		return nil, err
	}

	tag, err = c.app.UpdateChatTag(ctx, session.Domain(0), teamId, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (c *Controller) DeleteChatTag(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamChatTag, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	tag, err := c.app.RemoveChatTag(ctx, session.Domain(0), teamId, id)
	if err != nil {
		return nil, err
	}

	return tag, nil
}
