package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlCalendarStore struct {
	SqlStore
}

func NewSqlCalendarStore(sqlStore SqlStore) store.CalendarStore {
	us := &SqlCalendarStore{sqlStore}
	return us
}

func (s *SqlCalendarStore) CreateTableIfNotExists() {
}

func (s SqlCalendarStore) Create(calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	var out *model.Calendar
	if err := s.GetMaster().SelectOne(&out, `with c as (
		  insert into calendar (name,  domain_id, start, finish,description, timezone_id, created_at, created_by, updated_at, updated_by)
		  values (:Name, :DomainId, :Start, :Finish, :Description, :TimezoneId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy)
		  returning *
		)
		select
			c.id,
			c.name, 
			c.start,
			c.finish, 
			c.description, 
			c.domain_id, 
			cc_get_lookup(ct.id, ct.name) as timezone,
		    c.created_at,
		    cc_get_lookup(uc.id, uc.name) as created_by,
		    c.updated_at,
		    cc_get_lookup(u.id, u.name) as updated_by
		from c
		  inner join calendar_timezones ct on ct.id = c.timezone_id
	      left join directory.wbt_user uc on uc.id = c.created_by
	      left join directory.wbt_user u on u.id = c.updated_by`,
		map[string]interface{}{
			"Name":        calendar.Name,
			"DomainId":    calendar.DomainId,
			"Start":       calendar.Start,
			"Finish":      calendar.Finish,
			"Description": calendar.Description,
			"TimezoneId":  calendar.Timezone.Id,
			"CreatedAt":   calendar.CreatedAt,
			"CreatedBy":   calendar.CreatedBy.Id,
			"UpdatedAt":   calendar.UpdatedAt,
			"UpdatedBy":   calendar.UpdatedBy.Id,
		}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.Save", "store.sql_calendar.save.app_error", nil,
			fmt.Sprintf("id=%v, %v", calendar.Id, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlCalendarStore) GetAllPage(domainId int64, offset, limit int) ([]*model.Calendar, *model.AppError) {
	var calendars []*model.Calendar

	if _, err := s.GetReplica().Select(&calendars,
		`select c.id,
       c.name,
       c.start,
       c.finish,
       c.description,
	   c.domain_id,
       cc_get_lookup(ct.id, ct.name) as timezone,
	   c.created_at,
	   cc_get_lookup(uc.id, uc.name) as created_by,
       c.updated_at,
       cc_get_lookup(u.id, u.name) as updated_by
	from calendar c
       left join calendar_timezones ct on c.timezone_id = ct.id
	   left join directory.wbt_user uc on uc.id = c.created_by
	   left join directory.wbt_user u on u.id = c.updated_by
where c.domain_id = :DomainId
order by id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAllPage", "store.sql_calendar.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return calendars, nil
	}
}

func (s SqlCalendarStore) CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from calendar_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlCalendarStore) GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Calendar, *model.AppError) {
	var calendars []*model.Calendar

	if _, err := s.GetReplica().Select(&calendars,
		`select c.id,
       c.name,
       c.start,
       c.finish,
       c.description,
	   c.domain_id,
       cc_get_lookup(ct.id, ct.name) as timezone,
	   c.created_at,
	   cc_get_lookup(uc.id, uc.name) as created_by,
       c.updated_at,
       cc_get_lookup(u.id, u.name) as updated_by
from calendar c
       left join calendar_timezones ct on c.timezone_id = ct.id
	   left join directory.wbt_user uc on uc.id = c.created_by
	   left join directory.wbt_user u on u.id = c.updated_by
where c.domain_id = :DomainId
  and (
    exists(select 1
      from calendar_acl a
      where a.dc = c.domain_id and a.object = c.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
  )
order by id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset, "Groups": pq.Array(groups), "Access": model.PERMISSION_ACCESS_READ.Value()}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAllPage", "store.sql_calendar.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return calendars, nil
	}
}

func (s SqlCalendarStore) Get(domainId int64, id int64) (*model.Calendar, *model.AppError) {
	var calendar *model.Calendar
	if err := s.GetReplica().SelectOne(&calendar, `
			select c.id,
			   c.name,
			   c.start,
			   c.finish,
			   c.description,
			   c.domain_id,
			   cc_get_lookup(ct.id, ct.name) as timezone,
			   c.created_at,
			   cc_get_lookup(uc.id, uc.name) as created_by,
			   c.updated_at,
			   cc_get_lookup(u.id, u.name) as updated_by
		from calendar c
			   left join calendar_timezones ct on c.timezone_id = ct.id
			   left join directory.wbt_user uc on uc.id = c.created_by
			   left join directory.wbt_user u on u.id = c.updated_by
		where c.domain_id = :DomainId and c.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlCalendarStore.Get", "store.sql_calendar.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlCalendarStore.Get", "store.sql_calendar.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
		}
	} else {
		return calendar, nil
	}
}

func (s SqlCalendarStore) Update(calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	err := s.GetMaster().SelectOne(&calendar, `with c as (
    update calendar
	set name = :Name,
    timezone_id = :TimezoneId,
    description = :Description,
    finish = :Finish,
    start = :Start,
    updated_at = :UpdatedAt,
	updated_by = :UpdatedBy
where id = :Id and domain_id = :DomainId
    returning *
)
select c.id,
       c.name,
       c.start,
       c.finish,
       c.description,
       c.domain_id,
       cc_get_lookup(ct.id, ct.name) as timezone,
       c.created_at,
       cc_get_lookup(uc.id, uc.name) as created_by,
       c.updated_at,
       cc_get_lookup(u.id, u.name) as updated_by
from c
       left join calendar_timezones ct on c.timezone_id = ct.id
       left join directory.wbt_user uc on uc.id = c.created_by
       left join directory.wbt_user u on u.id = c.updated_by`, map[string]interface{}{
		"Name":        calendar.Name,
		"TimezoneId":  calendar.Timezone.Id,
		"Description": calendar.Description,
		"Finish":      calendar.Finish,
		"Start":       calendar.Start,
		"Id":          calendar.Id,
		"DomainId":    calendar.DomainId,
		"UpdatedAt":   calendar.UpdatedAt,
		"UpdatedBy":   calendar.UpdatedBy.Id,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.Update", "store.sql_calendar.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", calendar.Id, err.Error()), extractCodeFromErr(err))
	}
	return calendar, nil
}

func (s SqlCalendarStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from calendar c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlCalendarStore.Delete", "store.sql_calendar.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlCalendarStore) GetTimezoneAllPage(offset, limit int) ([]*model.Timezone, *model.AppError) {
	var timezones []*model.Timezone

	if _, err := s.GetReplica().Select(&timezones, `select id, name, utc_offset::text as "offset" from calendar_timezones 
		order by name limit :Limit offset :Offset`, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetTimezoneAllPage", "store.sql_calendar_timezone.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return timezones, nil
	}
}

func (s SqlCalendarStore) GetAcceptOfDayAllPage(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError) {
	var list []*model.CalendarAcceptOfDay

	if _, err := s.GetReplica().Select(&list, `select id, week_day, start_time_of_day, end_time_of_day, disabled
		from calendar_accept_of_day a
		where a.calendar_id = :CalendarId
		order by a.week_day`, map[string]interface{}{"CalendarId": calendarId}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAcceptOfDay", "store.sql_calendar_accept.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return list, nil
	}
}

func (s SqlCalendarStore) CreateAcceptOfDay(domainId, calendarId int64, timeRange *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError) {
	var out *model.CalendarAcceptOfDay
	err := s.GetMaster().SelectOne(&out, `insert into calendar_accept_of_day (calendar_id, week_day, start_time_of_day, end_time_of_day, disabled)
select c.id, :WeekDay, :StartTimeOfDay, :EndTimeOfDay, :Disabled
from calendar c
where c.id = :CalendarId and c.domain_id = :DomainId
returning id, week_day, start_time_of_day, end_time_of_day, disabled`, map[string]interface{}{
		"WeekDay":        timeRange.Week,
		"StartTimeOfDay": timeRange.StartTimeOfDay,
		"EndTimeOfDay":   timeRange.EndTimeOfDay,
		"Disabled":       timeRange.Disabled,
		"CalendarId":     calendarId,
		"DomainId":       domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.CreateAcceptOfDay", "store.sql_calendar_accept_range.save.app_error", nil,
			fmt.Sprintf("Calendarid=%v, %v", calendarId, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

//TODO check domain_id
func (s SqlCalendarStore) GetAcceptOfDayById(domainId, calendarId, id int64) (*model.CalendarAcceptOfDay, *model.AppError) {
	var out *model.CalendarAcceptOfDay
	err := s.GetReplica().SelectOne(&out, `select id, week_day, start_time_of_day, end_time_of_day, disabled
		from calendar_accept_of_day a
		where a.id = :Id and a.calendar_id = :CalendarId`, map[string]interface{}{
		"Id":         id,
		"CalendarId": calendarId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.CreateAcceptOfDay", "store.sql_calendar_accept_range.get.app_error", nil,
			fmt.Sprintf("Id=%v, Calendarid=%v, %v", id, calendarId, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCalendarStore) UpdateAcceptOfDay(calendarId int64, rangeTime *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError) {
	err := s.GetMaster().SelectOne(&rangeTime, `update calendar_accept_of_day
set start_time_of_day = :StartTimeOfDay,
    end_time_of_day = :EndTimeOfDay,
    disabled = :Disabled,
    week_day = :WeekDay
where id = :Id and calendar_id = :CalendarId
returning id, week_day, start_time_of_day, end_time_of_day, disabled`, map[string]interface{}{
		"StartTimeOfDay": rangeTime.StartTimeOfDay,
		"EndTimeOfDay":   rangeTime.EndTimeOfDay,
		"Disabled":       rangeTime.Disabled,
		"WeekDay":        rangeTime.Week,
		"Id":             rangeTime.Id,
		"CalendarId":     calendarId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.UpdateAcceptOfDay", "store.sql_calendar_accept_range.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", rangeTime.Id, err.Error()), extractCodeFromErr(err))
	}

	return rangeTime, nil
}

//TODO check domain_id ?
func (s SqlCalendarStore) DeleteAcceptOfDay(domainId, calendarId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from calendar_accept_of_day c where c.id=:Id and c.calendar_id = :CalendarId`,
		map[string]interface{}{
			"Id":         id,
			"CalendarId": calendarId,
			"DomainId":   domainId,
		}); err != nil {
		return model.NewAppError("SqlCalendarStore.DeleteAcceptOfDay", "store.sql_calendar_accept_range.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlCalendarStore) CreateExcept(domainId, calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError) {
	var out *model.CalendarExceptDate
	err := s.GetMaster().SelectOne(&out, `insert into calendar_except (calendar_id, name, date, repeat, disabled)
select c.id, :Name, :Date, :Repeat, :Disabled
from calendar c
where c.id = :CalendarId and c.domain_id = :DomainId
returning id, name, date, repeat, disabled`, map[string]interface{}{
		"Name":       except.Name,
		"Date":       except.Date,
		"Repeat":     except.Repeat,
		"Disabled":   except.Disabled,
		"CalendarId": calendarId,
		"DomainId":   domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.CreateExcept", "store.sql_calendar_except.save.app_error", nil,
			fmt.Sprintf("Calendarid=%v, %v", calendarId, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCalendarStore) GetExceptById(domainId, calendarId, id int64) (*model.CalendarExceptDate, *model.AppError) {
	var out *model.CalendarExceptDate
	err := s.GetReplica().SelectOne(&out, `select id, name, date, repeat, disabled
from calendar_except a
where a.id = :Id and a.calendar_id = :CalendarId`, map[string]interface{}{
		"Id":         id,
		"CalendarId": calendarId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetExceptById", "store.sql_calendar_except.get.app_error", nil,
			fmt.Sprintf("Id=%v, CalendarId=%v, %v", id, calendarId, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCalendarStore) GetExceptAllPage(calendarId int64) ([]*model.CalendarExceptDate, *model.AppError) {
	var list []*model.CalendarExceptDate

	if _, err := s.GetReplica().Select(&list, `select id, name, date, repeat, disabled
		from calendar_except a
		where a.calendar_id = :CalendarId
		order by a.id`, map[string]interface{}{"CalendarId": calendarId}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetExceptAllPage", "store.sql_calendar_except.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return list, nil
	}
}

func (s SqlCalendarStore) UpdateExceptDate(calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError) {
	err := s.GetMaster().SelectOne(&except, `update calendar_except
set name = :Name,
    date = :Date,
    repeat = :Repeat,
    disabled = :Disabled
where id = :Id and calendar_id = :CalendarId
returning id, name, date, repeat, disabled`, map[string]interface{}{
		"Name":       except.Name,
		"Date":       except.Date,
		"Repeat":     except.Repeat,
		"Disabled":   except.Disabled,
		"Id":         except.Id,
		"CalendarId": calendarId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.UpdateExceptDate", "store.sql_calendar_except.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", except.Id, err.Error()), extractCodeFromErr(err))
	}

	return except, nil
}

//TODO check domain_id ?
func (s SqlCalendarStore) DeleteExceptDate(domainId, calendarId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from calendar_except c where c.id=:Id and c.calendar_id = :CalendarId`,
		map[string]interface{}{
			"Id":         id,
			"CalendarId": calendarId,
			"DomainId":   domainId,
		}); err != nil {
		return model.NewAppError("SqlCalendarStore.DeleteExceptDate", "store.sql_calendar_except.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
