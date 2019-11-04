package app

import "github.com/webitel/engine/model"

func (a *App) GetCalendarsPage(domainId int64, page, perPage int) ([]*model.Calendar, *model.AppError) {
	return a.Store.Calendar().GetAllPage(domainId, page*perPage, perPage)
}

func (a *App) CreateCalendar(calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	return a.Store.Calendar().Create(calendar)
}

func (a *App) UpdateCalendar(calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	oldCalendar, err := a.GetCalendarById(calendar.DomainId, calendar.Id)
	if err != nil {
		return nil, err
	}

	oldCalendar.Timezone.Id = calendar.Timezone.Id
	oldCalendar.Description = calendar.Description
	oldCalendar.Name = calendar.Name
	oldCalendar.StartAt = calendar.StartAt
	oldCalendar.EndAt = calendar.EndAt
	oldCalendar.UpdatedAt = calendar.UpdatedAt
	oldCalendar.UpdatedBy.Id = calendar.UpdatedBy.Id

	oldCalendar, err = a.Store.Calendar().Update(oldCalendar)
	if err != nil {
		return nil, err
	}

	return oldCalendar, nil
}

func (a *App) CalendarCheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Calendar().CheckAccess(domainId, id, groups, access)
}

func (a *App) GetCalendarPageByGroups(domainId int64, groups []int, page, perPage int) ([]*model.Calendar, *model.AppError) {
	return a.Store.Calendar().GetAllPageByGroups(domainId, groups, page*perPage, perPage)
}

func (a *App) GetCalendarById(domainId, id int64) (*model.Calendar, *model.AppError) {
	return a.Store.Calendar().Get(domainId, id)
}

func (a *App) CreateCalendarAcceptOfDay(domainId, calendarId int64, accept *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError) {
	return a.Store.Calendar().CreateAcceptOfDay(domainId, calendarId, accept)
}

func (a *App) GetCalendarAcceptOfDayAllPage(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError) {
	return a.Store.Calendar().GetAcceptOfDayAllPage(calendarId)
}

func (a *App) GetCalendarAcceptOfDayById(domainId, calendarId, id int64) (*model.CalendarAcceptOfDay, *model.AppError) {
	return a.Store.Calendar().GetAcceptOfDayById(domainId, calendarId, id)
}

func (a *App) UpdateCalendarAcceptOfDay(domainId, calendarId int64, timeRange *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError) {
	oldAccept, err := a.GetCalendarAcceptOfDayById(domainId, calendarId, timeRange.Id)
	if err != nil {
		return nil, err
	}

	oldAccept.Day = timeRange.Day
	oldAccept.StartTimeOfDay = timeRange.StartTimeOfDay
	oldAccept.EndTimeOfDay = timeRange.EndTimeOfDay
	oldAccept.Disabled = timeRange.Disabled

	oldAccept, err = a.Store.Calendar().UpdateAcceptOfDay(calendarId, oldAccept)
	if err != nil {
		return nil, err
	}

	return oldAccept, nil
}

func (a *App) RemoveCalendarAcceptOfDay(domainId, calendarId, id int64) (*model.CalendarAcceptOfDay, *model.AppError) {
	accept, err := a.Store.Calendar().GetAcceptOfDayById(domainId, calendarId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Calendar().DeleteAcceptOfDay(domainId, calendarId, id)
	if err != nil {
		return nil, err
	}
	return accept, nil
}

func (a *App) RemoveCalendar(domainId, id int64) (*model.Calendar, *model.AppError) {
	calendar, err := a.Store.Calendar().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Calendar().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return calendar, nil
}

func (a *App) GetCalendarTimezoneAllPage(page, perPage int) ([]*model.Timezone, *model.AppError) {
	return a.Store.Calendar().GetTimezoneAllPage(page*perPage, perPage)
}

func (a *App) CreateCalendarExceptDate(domainId, calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError) {
	return a.Store.Calendar().CreateExcept(domainId, calendarId, except)
}

func (a *App) GetCalendarExceptDateById(domainId, calendarId, id int64) (*model.CalendarExceptDate, *model.AppError) {
	return a.Store.Calendar().GetExceptById(domainId, calendarId, id)
}

func (a *App) CalendarExceptDateAllPage(calendarId int64) ([]*model.CalendarExceptDate, *model.AppError) {
	return a.Store.Calendar().GetExceptAllPage(calendarId)
}

func (a *App) UpdateCalendarExceptDate(domainId, calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError) {
	oldExcept, err := a.GetCalendarExceptDateById(domainId, calendarId, except.Id)
	if err != nil {
		return nil, err
	}

	oldExcept.Name = except.Name
	oldExcept.Date = except.Date
	oldExcept.Repeat = except.Repeat
	oldExcept.Disabled = except.Disabled

	oldExcept, err = a.Store.Calendar().UpdateExceptDate(calendarId, oldExcept)
	if err != nil {
		return nil, err
	}

	return oldExcept, nil
}

func (a *App) RemoveCalendarExceptDate(domainId, calendarId, id int64) (*model.CalendarExceptDate, *model.AppError) {
	except, err := a.Store.Calendar().GetExceptById(domainId, calendarId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Calendar().DeleteExceptDate(domainId, calendarId, id)
	if err != nil {
		return nil, err
	}
	return except, nil
}
