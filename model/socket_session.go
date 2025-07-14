package model

import "time"

type SocketSession struct {
	Id        string    `json:"id" db:"id"`
	UserId    int64     `json:"user_id" db:"user_id"`
	DomainId  int64     `json:"domain_id" db:"domain_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Ip        string    `json:"ip" db:"ip"`
	Client    string    `json:"client" db:"client"`
	AppId     string    `json:"app_id" db:"app_id"`
}
