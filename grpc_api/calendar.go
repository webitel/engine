package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
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
		Name: in.Name,
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Description: in.GetDescription(),
	}

	for _, v := range in.Accepts {
		calendar.Accepts = append(calendar.Accepts, model.CalendarAcceptOfDay{
			Day:            int8(v.GetDay()),
			StartTimeOfDay: int16(v.GetStartTimeOfDay()),
			EndTimeOfDay:   int16(v.GetEndTimeOfDay()),
			Disabled:       v.GetDisabled(),
		})
	}

	for _, v := range in.Excepts {
		calendar.Excepts = append(calendar.Excepts, &model.CalendarExceptDate{
			Name:     v.GetName(),
			Repeat:   v.GetRepeat(),
			Date:     v.GetDate(),
			Disabled: v.GetDisabled(),
		})
	}

	if in.StartAt > 0 {
		calendar.StartAt = model.NewInt64(in.StartAt)
	}
	if in.EndAt > 0 {
		calendar.EndAt = model.NewInt64(in.EndAt)
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var calendar *model.Calendar

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	calendar := &model.Calendar{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:    in.Name,
		StartAt: &in.StartAt,
		EndAt:   &in.EndAt,
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Description: in.Description,
	}

	for _, v := range in.Accepts {
		calendar.Accepts = append(calendar.Accepts, model.CalendarAcceptOfDay{
			Day:            int8(v.GetDay()),
			StartTimeOfDay: int16(v.GetStartTimeOfDay()),
			EndTimeOfDay:   int16(v.GetEndTimeOfDay()),
			Disabled:       v.GetDisabled(),
		})
	}

	for _, v := range in.Excepts {
		calendar.Excepts = append(calendar.Excepts, &model.CalendarExceptDate{
			Name:     v.GetName(),
			Repeat:   v.GetRepeat(),
			Date:     v.GetDate(),
			Disabled: v.GetDisabled(),
		})
	}

	if err = calendar.IsValid(); err != nil {
		return nil, err
	}

	calendar, err = api.app.UpdateCalendar(calendar)

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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.CalendarCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
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
		Name:    src.Name,
		StartAt: 0,
		EndAt:   0,
		Timezone: &engine.Lookup{
			Id:   int64(src.Timezone.Id),
			Name: src.Timezone.Name,
		},
		Description: src.Description,
	}

	if len(src.Accepts) > 0 {
		item.Accepts = make([]*engine.AcceptOfDay, 0, len(src.Accepts))
		for _, v := range src.Accepts {
			item.Accepts = append(item.Accepts, transformAcceptOfDay(v))
		}
	}

	if len(src.Excepts) > 0 {
		item.Excepts = make([]*engine.ExceptDate, 0, len(src.Excepts))
		for _, v := range src.Excepts {
			item.Excepts = append(item.Excepts, transformExceptDate(v))
		}
	}

	if src.StartAt != nil {
		item.StartAt = *src.StartAt
	}

	if src.EndAt != nil {
		item.EndAt = *src.EndAt
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

func transformAcceptOfDay(src model.CalendarAcceptOfDay) *engine.AcceptOfDay {
	return &engine.AcceptOfDay{
		Day:            int32(src.Day),
		StartTimeOfDay: int32(src.StartTimeOfDay),
		EndTimeOfDay:   int32(src.EndTimeOfDay),
		Disabled:       src.Disabled,
	}
}

func transformExceptDate(src *model.CalendarExceptDate) *engine.ExceptDate {
	return &engine.ExceptDate{
		Name:     src.Name,
		Date:     src.Date,
		Repeat:   src.Repeat,
		Disabled: src.Disabled,
	}
}
