package model

import (
	"encoding/json"
	"strings"
)

type AppointmentWidget struct {
	Profile        AppointmentProfile `json:"profile" db:"profile"`
	List           []AppointmentDate  `json:"list" db:"list"`
	ComputedList   []byte
	allowedOrigins []originPattern
}

type AppointmentProfile struct {
	Uri                 string   `json:"uri"`
	Id                  int      `json:"id" db:"id"`
	DomainId            int64    `json:"domain_id" db:"domain_id"`
	QueueId             int      `json:"queue_id" db:"queue_id"`
	CommunicationTypeId int      `json:"communication_type" db:"communication_type"`
	Duration            string   `json:"duration" db:"duration"`
	Days                int      `json:"days" db:"days"`
	AvailableAgents     int      `json:"available_agents" db:"available_agents"`
	AllowOrigins        []string `json:"allow_origins" db:"allow_origins"`
	Timezone            string   `json:"timezone"`
	TimezoneId          int      `json:"timezone_id"`
}

type AppointmentDate struct {
	Date  string            `json:"date" db:"date"`
	Times []AppointmentTime `json:"times" db:"times"`
}

type AppointmentTime struct {
	Time     string `json:"time"`
	Reserved bool   `json:"reserved"`
}

type Appointment struct {
	Key          string    `json:"-" db:"-"`
	Id           int64     `json:"-" db:"id"`
	Ip           string    `json:"-" db:"import_id"`
	Timezone     string    `json:"-" db:"timezone"`
	ScheduleDate string    `json:"schedule_date" db:"schedule_date"`
	ScheduleTime string    `json:"schedule_time" db:"schedule_time"`
	Name         string    `json:"name" db:"name"`
	Destination  string    `json:"destination" db:"destination"`
	Variables    StringMap `json:"variables" db:"variables"`
	Computed     []byte    `json:"-" db:"-"`
}

type AppointmentResponse struct {
	Timezone    string            `json:"timezone"`
	Type        string            `json:"type"`
	List        []AppointmentDate `json:"list,omitempty"`
	Appointment *Appointment      `json:"appointment,omitempty"`
}

func (a *AppointmentResponse) ToJSON() []byte {
	data, _ := json.Marshal(a)
	return data
}

func (a *Appointment) ToJSON() []byte {
	data, _ := json.Marshal(a)
	return data
}

func AppointmentsDateToJson(src []AppointmentDate) []byte {
	data, _ := json.Marshal(src)
	return data
}

func AppointmentFromJson(data []byte) *Appointment {
	var app Appointment
	json.Unmarshal(data, &app)
	return &app
}

func (w *AppointmentWidget) ValidAppointment(a *Appointment) bool {
	for _, v := range w.List {
		if v.Date == a.ScheduleDate {
			for _, t := range v.Times {
				if t.Time == a.ScheduleTime {
					return !t.Reserved
				}
			}
			break
		}
	}

	return false
}

type originPattern interface {
	match(origin string) bool
}

type originAny bool

func (pttn originAny) match(origin string) bool {
	return origin != "" && (bool)(pttn)
}

type originWildcard [2]string

func (pttn originWildcard) match(origin string) bool {
	prefix, suffix := pttn[0], pttn[1]
	return len(origin) >= len(prefix)+len(suffix) &&
		strings.HasPrefix(origin, prefix) &&
		strings.HasSuffix(origin, suffix)
}

type originString string

func (pttn originString) match(origin string) bool {
	return (string)(pttn) == (origin)
}

func (w *AppointmentWidget) InitOrigin() {
	w.allowedOrigins = make([]originPattern, 0, len(w.Profile.AllowOrigins))
	for _, origin := range w.Profile.AllowOrigins {
		// Normalize
		origin = strings.ToLower(origin)
		if origin == "*" {
			// If "*" is present in the list, turn the whole list into a match all
			w.allowedOrigins = append(w.allowedOrigins[:0], originAny(true))
			break
		} else if i := strings.IndexByte(origin, '*'); i >= 0 {
			// Split the origin in two: start and end string without the *
			w.allowedOrigins = append(w.allowedOrigins, originWildcard{origin[0:i], origin[i+1:]})
		} else if origin != "" {
			w.allowedOrigins = append(w.allowedOrigins, originString(origin))
		}
	}
}

func (w *AppointmentWidget) AllowOrigin(origin string) bool {
	if len(w.allowedOrigins) != 0 {
		origin = strings.ToLower(origin)
		for _, allowedOrigin := range w.allowedOrigins {
			if allowedOrigin.match(origin) {
				return true
			}
		}
		return false
	}

	return true
}
