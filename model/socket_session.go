package model

import "time"

type SocketSession struct {
	Id              string    `json:"id" db:"id"`
	UserId          int64     `json:"user_id" db:"user_id"`
	DomainId        int64     `json:"domain_id" db:"domain_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	UserAgent       string    `json:"user_agent" db:"user_agent"`
	Ip              string    `json:"ip" db:"ip"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	Ver             string    `json:"ver" db:"ver"`
	AppId           string    `json:"app_id" db:"app_id"`
}

type SocketSessionView struct {
	Id              string     `json:"id" db:"id"`
	CreatedAt       *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
	UserAgent       string     `json:"user_agent" db:"user_agent"`
	Ip              string     `json:"ip" db:"ip"`
	ApplicationName string     `json:"application_name" db:"application_name"`
	Ver             string     `json:"ver" db:"ver"`
	User            *Lookup    `json:"user" db:"user"`
	Duration        int64      `json:"duration" db:"duration"`
	Pong            int64      `json:"pong" db:"pong"`
}

type SearchSocketSessionView struct {
	ListRequest
	UserIds []int64
}

func (SocketSessionView) DefaultOrder() string {
	return "created_at"
}

func (SocketSessionView) AllowFields() []string {
	return []string{"id", "created_at", "updated_at", "user_agent", "ip", "application_name", "ver", "user", "duration", "pong"}
}

func (s SocketSessionView) DefaultFields() []string {
	return s.AllowFields()
}

func (SocketSessionView) EntityName() string {
	return "socket_session_view"
}
