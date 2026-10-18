package model

import "time"

type UserNotification struct {
	Id        int64      `json:"id" db:"id"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	Type      string     `json:"type" db:"type"`
	Message   string     `json:"message" db:"message"`
	ReadAt    *time.Time `json:"read_at" db:"read_at"`
}

type SearchUserNotification struct {
	ListRequest

	CreatedAt *FilterBetween
	Read      *bool
}

func (UserNotification) DefaultOrder() string {
	return "-created_at"
}

func (UserNotification) AllowFields() []string {
	return []string{"id", "created_at", "type", "message", "read_at"}
}

func (s UserNotification) DefaultFields() []string {
	return s.AllowFields()
}

func (UserNotification) EntityName() string {
	return "cc_user_notification"
}
