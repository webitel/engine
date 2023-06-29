package model

import (
	"encoding/json"
	"io"
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
	Name     string `json:"name" db:"name"`
	Repeat   bool   `json:"repeat" db:"repeat"`
	Date     int64  `json:"date" db:"date"`
	Disabled bool   `json:"disabled" db:"disabled"`
}

// Description of the Calendar
// swagger:model Calendar
type Calendar struct {
	DomainRecord
	Name        string                `json:"name" db:"name"`
	StartAt     *int64                `json:"start_at" db:"start_at"`
	EndAt       *int64                `json:"end_at" db:"end_at"`
	Timezone    Lookup                `json:"timezone"`
	Description string                `json:"description,omitempty"`
	Accepts     []CalendarAcceptOfDay `json:"accepts" db:"accepts"`
	Excepts     []*CalendarExceptDate `json:"excepts" db:"excepts"`
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
		"accepts", "excepts"}
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
}

func (a *CalendarAcceptOfDay) IsValid() AppError {
	//TODO FIXME
	return nil
}

func (a *CalendarExceptDate) IsValid() AppError {
	//TODO FIXME
	return nil
}

func (c *Calendar) AcceptsToJson() string {
	b, _ := json.Marshal(c.Accepts)
	return string(b)
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

	return nil
}

func CalendarFromJson(data io.Reader) *Calendar {
	var calendar Calendar
	if err := json.NewDecoder(data).Decode(&calendar); err != nil {
		return nil
	} else {
		return &calendar
	}
}

func CalendarsToJson(calendars []*Calendar) string {
	b, _ := json.Marshal(calendars)
	return string(b)
}

func (c *Calendar) ToJson() string {
	b, _ := json.Marshal(c)
	return string(b)
}
