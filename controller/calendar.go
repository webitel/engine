package controller

import (
	"context"
	"strconv"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateCalendar(ctx context.Context, session *auth_manager.Session, calendar *model.Calendar) (*model.Calendar, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := calendar.IsValid(); err != nil {
		return nil, err
	}

	session.Domain(calendar.DomainId)

	res, err := c.app.CreateCalendar(ctx, calendar)
	if err != nil {
		return nil, err
	}
	c.app.AuditCreate(ctx, session, model.PERMISSION_SCOPE_CALENDAR, strconv.FormatInt(res.Id, 10), res)
	return res, nil
}

func (c *Controller) UpdateCalendar(ctx context.Context, session *auth_manager.Session, calendar *model.Calendar) (*model.Calendar, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		if perm, err := c.app.CalendarCheckAccess(ctx, session.Domain(calendar.DomainId), calendar.Id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, calendar.Id, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	calendar.DomainId = session.Domain(calendar.DomainId)

	if err := calendar.IsValid(); err != nil {
		return nil, err
	}

	res, err := c.app.UpdateCalendar(ctx, calendar)
	if err != nil {
		return nil, err
	}
	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CALENDAR, strconv.FormatInt(res.Id, 10), res)
	return res, nil
}

func (c *Controller) SearchCalendar(ctx context.Context, session *auth_manager.Session, search *model.SearchCalendar) ([]*model.Calendar, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetCalendarPageByGroups(ctx, session.Domain(search.DomainId), session.GetAclRoles(), search)
	} else {
		return c.app.GetCalendarsPage(ctx, session.Domain(search.DomainId), search)
	}
}

func (c *Controller) GetCalendar(ctx context.Context, session *auth_manager.Session, domainId, id int64) (*model.Calendar, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.CalendarCheckAccess(ctx, session.Domain(domainId), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetCalendarById(ctx, session.Domain(domainId), id)
}

func (c *Controller) DeleteCalendar(ctx context.Context, session *auth_manager.Session, domainId, id int64) (*model.Calendar, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		if perm, err := c.app.CalendarCheckAccess(ctx, session.Domain(domainId), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}
	res, err := c.app.RemoveCalendar(ctx, session.Domain(domainId), id)
	if err != nil {
		return nil, err
	}
	c.app.AuditDelete(ctx, session, model.PERMISSION_SCOPE_CALENDAR, strconv.FormatInt(res.Id, 10), res)
	return res, nil
}
