package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
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

func (s SqlCalendarStore) Create(ctx context.Context, calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	var out *model.Calendar
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with c as (
		  insert into flow.calendar (name,  domain_id, start_at, end_at, description, timezone_id, created_at, created_by, updated_at, updated_by, accepts, excepts)
		  values (:Name, :DomainId, :StartAt, :EndAt, :Description, :TimezoneId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, flow.calendar_json_to_accepts(:Accepts::jsonb), flow.calendar_json_to_excepts(:Excepts::jsonb))
		  returning *
		)
		select
			c.id,
			c.name, 
			c.start_at,
			c.end_at, 
			c.description, 
			c.domain_id, 
			call_center.cc_get_lookup(ct.id, ct.name) as timezone,
		    c.created_at,
		    call_center.cc_get_lookup(uc.id, uc.name) as created_by,
		    c.updated_at,
		    call_center.cc_get_lookup(u.id, u.name) as updated_by,
		    flow.calendar_accepts_to_jsonb(c.accepts)::jsonb as accepts,
		    call_center.cc_arr_type_to_jsonb(c.excepts)::jsonb as excepts
		from c
		  inner join flow.calendar_timezones ct on ct.id = c.timezone_id
	      left join directory.wbt_user uc on uc.id = c.created_by
	      left join directory.wbt_user u on u.id = c.updated_by`,
		map[string]interface{}{
			"Name":        calendar.Name,
			"DomainId":    calendar.DomainId,
			"StartAt":     calendar.StartAt,
			"EndAt":       calendar.EndAt,
			"Description": calendar.Description,
			"TimezoneId":  calendar.Timezone.Id,
			"CreatedAt":   calendar.CreatedAt,
			"CreatedBy":   calendar.CreatedBy.GetSafeId(),
			"UpdatedAt":   calendar.UpdatedAt,
			"UpdatedBy":   calendar.UpdatedBy.GetSafeId(),
			"Accepts":     calendar.AcceptsToJson(),
			"Excepts":     calendar.ExceptsToJson(),
		}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.Save", "store.sql_calendar.save.app_error", nil,
			fmt.Sprintf("id=%v, %v", calendar.Id, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlCalendarStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchCalendar) ([]*model.Calendar, *model.AppError) {
	var calendars []*model.Calendar

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
	}

	err := s.ListQueryFromSchema(ctx, &calendars, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.Calendar{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAllPage", "store.sql_calendar.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return calendars, nil
	}
}

func (s SqlCalendarStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from flow.calendar_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlCalendarStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchCalendar) ([]*model.Calendar, *model.AppError) {
	var calendars []*model.Calendar

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQueryFromSchema(ctx, &calendars, "flow", search.ListRequest,
		`domain_id = :DomainId
				and exists(select 1
				  from flow.calendar_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.Calendar{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetAllPage", "store.sql_calendar.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return calendars, nil
	}
}

func (s SqlCalendarStore) Get(ctx context.Context, domainId int64, id int64) (*model.Calendar, *model.AppError) {
	var calendar *model.Calendar
	if err := s.GetReplica().WithContext(ctx).SelectOne(&calendar, `
			select c.id,
			   c.name,
			   c.start_at,
			   c.end_at,
			   c.description,
			   c.domain_id,
			   call_center.cc_get_lookup(ct.id, ct.name) as timezone,
			   c.created_at,
			   call_center.cc_get_lookup(uc.id, uc.name) as created_by,
			   c.updated_at,
			   call_center.cc_get_lookup(u.id, u.name) as updated_by,
			   flow.calendar_accepts_to_jsonb(c.accepts) as accepts,
			   call_center.cc_arr_type_to_jsonb(c.excepts) as excepts
		from flow.calendar c
			   left join flow.calendar_timezones ct on c.timezone_id = ct.id
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

func (s SqlCalendarStore) Update(ctx context.Context, calendar *model.Calendar) (*model.Calendar, *model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&calendar, `with c as (
    update flow.calendar
	set name = :Name,
    timezone_id = :TimezoneId,
    description = :Description,
    end_at = :EndAt,
    start_at = :StartAt,
    updated_at = :UpdatedAt,
	updated_by = :UpdatedBy,
	accepts = flow.calendar_json_to_accepts(:Accepts),
	excepts = flow.calendar_json_to_excepts(:Excepts::jsonb)
where id = :Id and domain_id = :DomainId
    returning *
)
select c.id,
       c.name,
       c.end_at,
       c.end_at,
       c.description,
       c.domain_id,
       call_center.cc_get_lookup(ct.id, ct.name) as timezone,
       c.created_at,
       call_center.cc_get_lookup(uc.id, uc.name) as created_by,
       c.updated_at,
       call_center.cc_get_lookup(u.id, u.name) as updated_by,
	   flow.calendar_accepts_to_jsonb(c.accepts) as accepts,
	   call_center.cc_arr_type_to_jsonb(c.excepts) as excepts
from c
       left join flow.calendar_timezones ct on c.timezone_id = ct.id
       left join directory.wbt_user uc on uc.id = c.created_by
       left join directory.wbt_user u on u.id = c.updated_by`, map[string]interface{}{
		"Name":        calendar.Name,
		"TimezoneId":  calendar.Timezone.Id,
		"Description": calendar.Description,
		"StartAt":     calendar.StartAt,
		"EndAt":       calendar.EndAt,
		"Id":          calendar.Id,
		"DomainId":    calendar.DomainId,
		"UpdatedAt":   calendar.UpdatedAt,
		"UpdatedBy":   calendar.UpdatedBy.GetSafeId(),
		"Accepts":     calendar.AcceptsToJson(),
		"Excepts":     calendar.ExceptsToJson(),
	})
	if err != nil {
		return nil, model.NewAppError("SqlCalendarStore.Update", "store.sql_calendar.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", calendar.Id, err.Error()), extractCodeFromErr(err))
	}
	return calendar, nil
}

func (s SqlCalendarStore) Delete(ctx context.Context, domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from flow.calendar c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlCalendarStore.Delete", "store.sql_calendar.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlCalendarStore) GetTimezoneAllPage(ctx context.Context, search *model.SearchTimezone) ([]*model.Timezone, *model.AppError) {
	var timezones []*model.Timezone

	if _, err := s.GetReplica().WithContext(ctx).Select(&timezones, `select id, name, utc_offset::text as "offset" 
		from flow.calendar_timezones  t
		where  (:Q::varchar isnull or t.name ilike :Q::varchar)
		order by name limit :Limit offset :Offset`, map[string]interface{}{
		"Limit":  search.GetLimit(),
		"Offset": search.GetOffset(),
		"Q":      search.GetQ(),
	}); err != nil {
		return nil, model.NewAppError("SqlCalendarStore.GetTimezoneAllPage", "store.sql_calendar_timezone.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return timezones, nil
	}
}
