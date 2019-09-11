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
	oldCalendar.Finish = calendar.Finish
	oldCalendar.Start = calendar.Start

	a.Store.Calendar().Update(oldCalendar)

	return oldCalendar, nil //TODO
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

func (a *App) GetCalendarAcceptOfDay(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError) {
	return a.Store.Calendar().GetAcceptOfDay(calendarId)
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
