package model

import (
	"encoding/json"
	"unicode/utf8"
)

// Description of the CalendarAcceptOfDay
// swagger:model CalendarAcceptOfDay
type CalendarAcceptOfDay struct {
	Day            int8  `json:"day" db:"day"`
	StartTimeOfDay int16 `json:"start_time_of_day" db:"start_time_of_day"`
	EndTimeOfDay   int16 `json:"end_time_of_day" db:"end_time_of_day"`
	Disabled       bool  `json:"disabled" db:"disabled"`
}

// Description of the CalendarExceptDate
// swagger:model CalendarExceptDate
type CalendarExceptDate struct {
	Name      string `json:"name" db:"name"`
	Repeat    bool   `json:"repeat" db:"repeat"`
	Date      int64  `json:"date" db:"date"`
	Disabled  bool   `json:"disabled" db:"disabled"`
	WorkStart int32  `json:"work_start" db:"work_start"`
	WorkStop  int32  `json:"work_stop" db:"work_stop"`
	Working   bool   `json:"working" db:"working"`
}

// Description of the Calendar
// swagger:model Calendar
type Calendar struct {
	DomainRecord
	Name        string                 `json:"name" db:"name"`
	StartAt     *int64                 `json:"start_at" db:"start_at"`
	EndAt       *int64                 `json:"end_at" db:"end_at"`
	Timezone    Lookup                 `json:"timezone"`
	Description string                 `json:"description,omitempty"`
	Accepts     []CalendarAcceptOfDay  `json:"accepts" db:"accepts"`
	Excepts     []*CalendarExceptDate  `json:"excepts" db:"excepts"`
	Specials    []*CalendarAcceptOfDay `json:"specials" db:"specials"`
}

type SearchCalendar struct {
	ListRequest
	Ids []uint32
}

func (Calendar) DefaultOrder() string {
	return "id"
}

func (a Calendar) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description", "start_at", "end_at", "timezone", "created_at", "created_by", "updated_at", "updated_by",
		"accepts", "excepts", "specials"}
}

func (a Calendar) DefaultFields() []string {
	return []string{"id", "name", "description"}
}

func (a Calendar) EntityName() string {
	return "calendar_view"
}

// Description of the Timezone
// swagger:model Timezone
type Timezone struct {
	Id     int64  `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Offset string `json:"offset" db:"offset"`
}

type SearchTimezone struct {
	ListRequest
	Ids []uint32
}

func (a *CalendarAcceptOfDay) IsValid() AppError {
	// TODO FIXME
	return nil
}

func (a *CalendarExceptDate) IsValid() AppError {
	// TODO FIXME
	return nil
}

func (c *Calendar) AcceptsToJson() string {
	b, _ := json.Marshal(c.Accepts)
	return string(b)
}

func (c *Calendar) SpecialsToJson() *string {
	if c.Specials == nil {
		return nil
	}
	b, _ := json.Marshal(c.Specials)
	return NewString(string(b))
}

func (c *Calendar) ExceptsToJson() *string {
	if c.Excepts == nil {
		return nil
	}
	b, _ := json.Marshal(c.Excepts)
	return NewString(string(b))
}

func (c *Calendar) IsValid() AppError {
	if utf8.RuneCountInString(c.Name) <= 3 {
		return NewBadRequestError("model.calendar.is_valid.name.app_error", "name="+c.Name)
	}

	if c.DomainId == 0 {
		return NewBadRequestError("model.calendar.is_valid.domain_id.app_error", "name="+c.Name)
	}

	if len(c.Accepts) == 0 {
		return NewBadRequestError("model.calendar.is_valid.accepts.app_error", "name="+c.Name)
	}

	for _, a := range c.Accepts {
		if !(a.StartTimeOfDay >= 0 && a.StartTimeOfDay <= 1440) {
			return NewBadRequestError("model.calendar.is_valid.accepts.start_time_of_day", "start_time_of_day must be in the range 0-1440.")
		}
		if !(a.EndTimeOfDay >= 0 && a.EndTimeOfDay <= 1440) {
			return NewBadRequestError("model.calendar.is_valid.accepts.end_time_of_day", "end_time_of_day must be in the range 0-1440.")
		}
	}

	uq := make(map[string]struct{})
	for _, v := range c.Excepts {
		if v.Disabled {
			continue
		}
		key := Int64ToTime(v.Date).Format("2006-02-01")
		if _, ok := uq[key]; ok {
			return NewBadRequestError("model.calendar.is_valid.excepts.date", "You can't add another holiday on the same date "+key)
		}
		uq[key] = struct{}{}
	}

	return nil
}

func (c *Calendar) ToJson() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (Timezone) DefaultOrder() string {
	return "name"
}

func (a Timezone) AllowFields() []string {
	return []string{"id", "name", "offset"}
}

func (a Timezone) DefaultFields() []string {
	return []string{"id", "name", "offset"}
}

func (a Timezone) EntityName() string {
	return "calendar_timezones_view"
}
