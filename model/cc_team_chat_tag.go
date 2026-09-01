package model

import (
	"strings"
	"time"
	"unicode/utf8"
)

type TeamChatTag struct {
	Id        uint32     `json:"id" db:"id"`
	Tag       string     `json:"tag" db:"tag"`
	UpdatedBy *Lookup    `json:"updated_by" db:"updated_by"`
	CreatedBy *Lookup    `json:"created_by" db:"created_by"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
}

type SearchTeamChatTag struct {
	ListRequest
	Ids []uint32
}

func (tct TeamChatTag) AllowFields() []string {
	return tct.DefaultFields()
}

func (tct TeamChatTag) DefaultOrder() string {
	return "+tag"
}

func (TeamChatTag) DefaultFields() []string {
	return []string{"id", "tag"}
}

func (TeamChatTag) EntityName() string {
	return "cc_team_chat_tag_list"
}

func (tct *TeamChatTag) IsValid() AppError {
	tct.Tag = strings.TrimSpace(tct.Tag)
	if len(tct.Tag) == 0 {
		return NewBadRequestError("model.cc_team_chat_tag.is_valid.tag.app_error", "Tag is required")
	}
	if utf8.RuneCountInString(tct.Tag) > 64 {
		return NewBadRequestError("model.cc_team_chat_tag.is_valid.tag.max_length", "Tag must be 64 characters or less")
	}
	return nil
}
