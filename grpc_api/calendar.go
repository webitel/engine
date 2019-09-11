package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type calendar struct {
	app *app.App
}

func NewCalendarApi(app *app.App) *calendar {
	return &calendar{app}
}

func (api *calendar) Create(ctx context.Context, in *engine.Calendar) (*engine.Calendar, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	calendar := &model.Calendar{
		Name:     in.Name,
		DomainId: session.Domain(in.GetDomainId()),
		Start:    nil, //TODO
		Finish:   nil, //TODO
		Timezone: model.Lookup{
			Id: int(in.Timezone.Id),
		},
		Description: nil, //TODO
	}

	if err = calendar.IsValid(); err != nil {
		return nil, err
	}

	calendar, err = api.app.CreateCalendar(calendar)
	if err != nil {
		return nil, err
	}

	return transformCalendar(calendar), nil
}

func (api *calendar) List(ctx context.Context, in *engine.ListReqeust) (*engine.ListCalendar, error) {

	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var list []*model.Calendar

	if permission.Rbac {
		list, err = api.app.GetCalendarPageByGroups(session.Domain(int64(in.DomainId)), session.RoleIds, int(in.Page), int(in.Size))
	} else {
		list, err = api.app.GetCalendarsPage(session.Domain(int64(in.DomainId)), int(in.Page), int(in.Size))
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Calendar, 0, len(list))
	for _, v := range list {
		items = append(items, transformCalendar(v))
	}
	return &engine.ListCalendar{
		Items: items,
	}, nil
}

func (api *calendar) Get(ctx context.Context, in *engine.ItemRequest) (*engine.Calendar, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var calendar *model.Calendar

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	calendar, err = api.app.GetCalendarById(session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return transformCalendar(calendar), nil
}

func (api *calendar) Update(ctx context.Context, in *engine.Calendar) (*engine.Calendar, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var calendar *model.Calendar

	calendar, err = api.app.UpdateCalendar(&model.Calendar{
		Id:       in.Id,
		Name:     in.Name,
		DomainId: session.Domain(in.GetDomainId()),
		Start:    &in.Start,
		Finish:   &in.Finish,
		Timezone: model.Lookup{
			Id: int(in.Timezone.Id),
		},
		Description: &in.Description,
	})

	if err != nil {
		return nil, err
	}

	return transformCalendar(calendar), nil
}

func (api *calendar) Remove(ctx context.Context, in *engine.ItemRequest) (*engine.Calendar, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_DELETE)
		}
	}

	var calendar *model.Calendar
	calendar, err = api.app.RemoveCalendar(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformCalendar(calendar), nil
}

func (api *calendar) GetTimezones(ctx context.Context, in *engine.ListReqeust) (*engine.ListTimezoneResponse, error) {
	list, err := api.app.GetCalendarTimezoneAllPage(int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Timezone, 0, len(list))
	for _, v := range list {
		items = append(items, transformTimezone(v))
	}

	return &engine.ListTimezoneResponse{
		Items: items,
	}, nil
}

func (api *calendar) GetAcceptOfDay(ctx context.Context, in *engine.AcceptOfDayReqeust) (*engine.ListAcceptOfDay, error) {

	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(0), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.CalendarAcceptOfDay

	list, err = api.app.GetCalendarAcceptOfDay(in.GetCalendarId())
	if err != nil {
		return nil, err
	}

	result := &engine.ListAcceptOfDay{
		Items: make([]*engine.AcceptOfDay, 0, len(list)),
	}

	for _, v := range list {
		result.Items = append(result.Items, transformAcceptOfDay(v))
	}

	return result, nil
}

func transformCalendar(src *model.Calendar) *engine.Calendar {
	item := &engine.Calendar{
		Id:       src.Id,
		Name:     src.Name,
		DomainId: src.DomainId,
		Start:    0,
		Finish:   0,
		Timezone: &engine.Lookup{
			Id:   int64(src.Timezone.Id),
			Name: src.Timezone.Name,
		},
	}

	if src.Description != nil {
		item.Description = *src.Description
	}

	if src.Start != nil {
		item.Start = *src.Start
	}

	if src.Finish != nil {
		item.Finish = *src.Finish
	}

	return item
}

func transformTimezone(src *model.Timezone) *engine.Timezone {
	return &engine.Timezone{
		Id:     src.Id,
		Name:   src.Name,
		Offset: src.Offset,
	}
}

func transformAcceptOfDay(src *model.CalendarAcceptOfDay) *engine.AcceptOfDay {
	return &engine.AcceptOfDay{
		Id:             src.Id,
		WeekDay:        int32(src.Week),
		StartTimeOfDay: int32(src.StartTimeOfDay),
		EndTimeOfDay:   int32(src.EndTimeOfDay),
	}
}
