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
	if err := s.GetMaster().SelectOne(&out, `with i as (
		  insert into calendar (name,  domain_id, start, finish,description, timezone_id)
		  values (:Name, :DomainId, :Start, :Finish, :Description, :TimezoneId)
		  returning *
		)
		select i.id, i.name, i.start, i.finish, i.description, json_build_object('id', ct.id, 'name', ct.name)::jsonb as timezone  
		from i
		  inner join calendar_timezones ct on ct.id = i.timezone_id`,
		map[string]interface{}{"Name": calendar.Name, "DomainId": calendar.DomainId, "Start": calendar.Start,
			"Finish": calendar.Finish, "Description": calendar.Description, "TimezoneId": calendar.Timezone.Id}); err != nil {
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
       json_build_object('id', ct.id, 'name', ct.name)::jsonb as timezone
from calendar c
       left join calendar_timezones ct on c.timezone_id = ct.id
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
       json_build_object('id', ct.id, 'name', ct.name)::jsonb as timezone
from calendar c
       left join calendar_timezones ct on c.timezone_id = ct.id
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
			   json_build_object('id', ct.id, 'name', ct.name)::jsonb as timezone
		from calendar c
			   left join calendar_timezones ct on c.timezone_id = ct.id
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

func (s SqlCalendarStore) GetByGroups(domainId int64, id int64, groups []int) (*model.Calendar, *model.AppError) {
	var calendar *model.Calendar
	if err := s.GetReplica().SelectOne(&calendar, `
			select c.id,
			   c.name,
			   c.start,
			   c.finish,
			   c.description,
			   json_build_object('id', ct.id, 'name', ct.name)::jsonb as timezone
		from calendar c
			   left join calendar_timezones ct on c.timezone_id = ct.id
		where c.domain_id = :DomainId and c.id = :Id and (
			exists(select 1
			  from calendar_acl a
			  where a.dc = c.domain_id and a.object = c.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
		  )	
		`, map[string]interface{}{"Id": id, "DomainId": domainId, "Groups": pq.Array(groups), "Access": model.PERMISSION_ACCESS_READ.Value()}); err != nil {
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
	err := s.GetMaster().SelectOne(&calendar, `update calendar
	set name = :Name,
    timezone_id = :TimezoneId,
    description = :Description,
    finish = :Finish,
    start = :Start
where id = :Id returning *`, map[string]interface{}{
		"Name":        calendar.Name,
		"TimezoneId":  calendar.Timezone.Id,
		"Description": calendar.Description,
		"Finish":      calendar.Finish,
		"Start":       calendar.Start,
		"Id":          calendar.Id,
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

func (s SqlCalendarStore) GetAcceptOfDay(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError) {
	var list []*model.CalendarAcceptOfDay

	if _, err := s.GetReplica().Select(&list, `select id, week_day, start_time_of_day, end_time_of_day
		from calendar_accept_of_day a
		where a.calendar_id = :CalendarId
		order by a.week_day`, map[string]interface{}{"CalendarId": calendarId}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAcceptOfDay", "store.sql_calendar_accept.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return list, nil
	}
}
