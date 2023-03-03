package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) GetCalendarsPage(ctx context.Context, domainId int64, search *model.SearchCalendar) ([]*model.Calendar, bool, *model.AppError) {
	list, err := a.Store.Calendar().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetCalendarPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchCalendar) ([]*model.Calendar, bool, *model.AppError) {
	list, err := a.Store.Calendar().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateCalendar(ctx context.Context, calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	return a.Store.Calendar().Create(ctx, calendar)
}

func (a *App) UpdateCalendar(ctx context.Context, calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	oldCalendar, err := a.GetCalendarById(ctx, calendar.DomainId, calendar.Id)
	if err != nil {
		return nil, err
	}

	oldCalendar.Timezone.Id = calendar.Timezone.Id
	oldCalendar.Description = calendar.Description
	oldCalendar.Name = calendar.Name
	oldCalendar.StartAt = calendar.StartAt
	oldCalendar.EndAt = calendar.EndAt
	oldCalendar.UpdatedAt = calendar.UpdatedAt
	oldCalendar.UpdatedBy = calendar.UpdatedBy
	oldCalendar.Accepts = calendar.Accepts
	oldCalendar.Excepts = calendar.Excepts

	oldCalendar, err = a.Store.Calendar().Update(ctx, oldCalendar)
	if err != nil {
		return nil, err
	}

	return oldCalendar, nil
}

func (a *App) CalendarCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Calendar().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) GetCalendarById(ctx context.Context, domainId, id int64) (*model.Calendar, *model.AppError) {
	return a.Store.Calendar().Get(ctx, domainId, id)
}

func (a *App) RemoveCalendar(ctx context.Context, domainId, id int64) (*model.Calendar, *model.AppError) {
	calendar, err := a.Store.Calendar().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Calendar().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return calendar, nil
}

func (a *App) GetCalendarTimezoneAllPage(ctx context.Context, search *model.SearchTimezone) ([]*model.Timezone, bool, *model.AppError) {
	list, err := a.Store.Calendar().GetTimezoneAllPage(ctx, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
