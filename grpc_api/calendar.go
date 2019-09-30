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

func (api *calendar) CreateCalendar(ctx context.Context, in *engine.CreateCalendarRequest) (*engine.Calendar, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CALENDAR)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	calendar := &model.Calendar{
		DomainRecord: model.DomainRecord{
			Id:        0,
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
		Name:   in.Name,
		Start:  nil, //TODO
		Finish: nil, //TODO
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
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

func (api *calendar) SearchCalendar(ctx context.Context, in *engine.SearchCalendarRequest) (*engine.ListCalendar, error) {

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
		list, err = api.app.GetCalendarPageByGroups(session.Domain(in.DomainId), session.RoleIds, int(in.Page), int(in.Size))
	} else {
		list, err = api.app.GetCalendarsPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))
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

func (api *calendar) ReadCalendar(ctx context.Context, in *engine.ReadCalendarRequest) (*engine.Calendar, error) {
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

func (api *calendar) UpdateCalendar(ctx context.Context, in *engine.UpdateCalendarRequest) (*engine.Calendar, error) {
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
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:   in.Name,
		Start:  &in.Start,
		Finish: &in.Finish,
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Description: &in.Description,
	})

	if err != nil {
		return nil, err
	}

	return transformCalendar(calendar), nil
}

func (api *calendar) DeleteCalendar(ctx context.Context, in *engine.DeleteCalendarRequest) (*engine.Calendar, error) {
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

func (api *calendar) SearchTimezones(ctx context.Context, in *engine.SearchTimezonesRequest) (*engine.ListTimezoneResponse, error) {
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

func (api *calendar) CreateAcceptOfDay(ctx context.Context, in *engine.CreateAcceptOfDayRequest) (*engine.AcceptOfDay, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(0), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var accept *model.CalendarAcceptOfDay

	accept, err = api.app.CreateCalendarAcceptOfDay(session.Domain(in.GetDomainId()), in.GetCalendarId(), &model.CalendarAcceptOfDay{
		Week:           int8(in.GetWeekDay()),
		StartTimeOfDay: int16(in.GetStartTimeOfDay()),
		EndTimeOfDay:   int16(in.GetEndTimeOfDay()),
		Disabled:       in.GetDisabled(),
	})

	if err != nil {
		return nil, err
	}

	return transformAcceptOfDay(accept), nil
}

func (api *calendar) SearchAcceptOfDay(ctx context.Context, in *engine.AcceptOfDayRequest) (*engine.ListAcceptOfDay, error) {

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

	list, err = api.app.GetCalendarAcceptOfDayAllPage(in.GetCalendarId())
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

func (api *calendar) ReadAcceptOfDay(ctx context.Context, in *engine.ReadAcceptOfDayRequest) (*engine.AcceptOfDay, error) {
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

	var accept *model.CalendarAcceptOfDay
	accept, err = api.app.GetCalendarAcceptOfDayById(session.Domain(in.GetDomainId()), in.GetCalendarId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformAcceptOfDay(accept), nil
}

func (api *calendar) UpdateAcceptOfDay(ctx context.Context, in *engine.UpdateAcceptOfDayRequest) (*engine.AcceptOfDay, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var accept = &model.CalendarAcceptOfDay{
		Id:             in.GetId(),
		Week:           int8(in.GetWeekDay()),
		StartTimeOfDay: int16(in.GetStartTimeOfDay()),
		EndTimeOfDay:   int16(in.GetEndTimeOfDay()),
		Disabled:       in.GetDisabled(),
	}

	if err = accept.IsValid(); err != nil {
		return nil, err
	}

	accept, err = api.app.UpdateCalendarAcceptOfDay(session.Domain(in.GetDomainId()), in.GetCalendarId(), accept)
	if err != nil {
		return nil, err
	}
	return transformAcceptOfDay(accept), nil
}

func (api *calendar) DeleteAcceptOfDay(ctx context.Context, in *engine.DeleteAcceptOfDayRequest) (*engine.AcceptOfDay, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var accept *model.CalendarAcceptOfDay
	accept, err = api.app.RemoveCalendarAcceptOfDay(session.Domain(in.GetDomainId()), in.GetCalendarId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformAcceptOfDay(accept), nil
}

func (api *calendar) CreateExceptDate(ctx context.Context, in *engine.CreateExceptDateRequest) (*engine.ExceptDate, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(0), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var except *model.CalendarExceptDate

	except, err = api.app.CreateCalendarExceptDate(session.Domain(in.GetDomainId()), in.GetCalendarId(), &model.CalendarExceptDate{
		CalendarId: in.GetCalendarId(),
		Name:       in.GetName(),
		Repeat:     int8(in.GetRepeat()),
		Date:       in.GetDate(),
		Disabled:   in.GetDisabled(),
	})

	if err != nil {
		return nil, err
	}

	return transformExceptDate(except), nil
}

func (api *calendar) SearchExceptDate(ctx context.Context, in *engine.SearchExceptDateRequest) (*engine.ListExceptDate, error) {
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

	var list []*model.CalendarExceptDate

	list, err = api.app.CalendarExceptDateAllPage(in.GetCalendarId())
	if err != nil {
		return nil, err
	}

	result := &engine.ListExceptDate{
		Items: make([]*engine.ExceptDate, 0, len(list)),
	}

	for _, v := range list {
		result.Items = append(result.Items, transformExceptDate(v))
	}

	return result, nil
}

func (api *calendar) ReadExceptDate(ctx context.Context, in *engine.ReadExceptDateRequest) (*engine.ExceptDate, error) {
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

	var except *model.CalendarExceptDate
	except, err = api.app.GetCalendarExceptDateById(session.Domain(in.GetDomainId()), in.GetCalendarId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformExceptDate(except), nil
}

func (api *calendar) UpdateExceptDate(ctx context.Context, in *engine.UpdateExceptDateRequest) (*engine.ExceptDate, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var except = &model.CalendarExceptDate{
		Id:         in.GetId(),
		CalendarId: in.GetCalendarId(),
		Name:       in.GetName(),
		Repeat:     int8(in.GetRepeat()),
		Date:       in.GetDate(),
		Disabled:   in.GetDisabled(),
	}

	if err = except.IsValid(); err != nil {
		return nil, err
	}

	except, err = api.app.UpdateCalendarExceptDate(session.Domain(in.GetDomainId()), in.GetCalendarId(), except)
	if err != nil {
		return nil, err
	}
	return transformExceptDate(except), nil
}

func (api *calendar) DeleteExceptDate(ctx context.Context, in *engine.DeleteExceptDateRequest) (*engine.ExceptDate, error) {
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
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetCalendarId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetCalendarId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var except *model.CalendarExceptDate
	except, err = api.app.RemoveCalendarExceptDate(session.Domain(in.GetDomainId()), in.GetCalendarId(), in.GetId())
	if err != nil {
		return nil, err
	}

	return transformExceptDate(except), nil
}

func transformCalendar(src *model.Calendar) *engine.Calendar {
	item := &engine.Calendar{
		Id:        src.Id,
		DomainId:  src.DomainId,
		CreatedAt: src.CreatedAt,
		CreatedBy: &engine.Lookup{
			Id:   int64(src.CreatedBy.Id),
			Name: src.CreatedBy.Name,
		},
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: &engine.Lookup{
			Id:   int64(src.UpdatedBy.Id),
			Name: src.UpdatedBy.Name,
		},
		Name:   src.Name,
		Start:  0,
		Finish: 0,
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
		Disabled:       src.Disabled,
	}
}

func transformExceptDate(src *model.CalendarExceptDate) *engine.ExceptDate {
	return &engine.ExceptDate{
		Id:         src.Id,
		CalendarId: src.CalendarId,
		Name:       src.Name,
		Date:       int64(src.Date),
		Repeat:     int32(src.Repeat),
		Disabled:   src.Disabled,
	}
}
