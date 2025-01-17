package model

import "time"

type QuickReply struct {
	AclRecord
	Id      int     `json:"id" db:"id"`
	Name    string  `json:"name" db:"name"`
	Text    string  `json:"text" db:"text"`
	Queue   *Lookup `json:"queue" db:"queue"`
	Team    *Lookup `json:"team" db:"team"`
	Article *Lookup `json:"article" db:"article"`
}

type SearchQuickReply struct {
	ListRequest
	Ids  []uint32
	Name *string
}

type QuickReplyPatch struct {
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy Lookup     `json:"updated_by"`
	Name      *string    `json:"name"`
	Text      *string    `json:"text"`
	Queue     *Lookup    `json:"queue" db:"queue"`
	Team      *Lookup    `json:"team" db:"team"`
	Article   *Lookup    `json:"article" db:"article"`
}

func (p QuickReply) AllowFields() []string {
	return []string{"id", "created_by", "created_at", "updated_by", "updated_at", "name", "text", "team", "queue", "article"}
}

func (QuickReply) DefaultOrder() string {
	return "-name"
}

func (QuickReply) DefaultFields() []string {
	return []string{"id", "name", "description", "limit_min", "allow_agent", "allow_supervisor", "allow_admin"}
}

func (QuickReply) EntityName() string {
	return "cc_pause_cause_list"
}

func (p *QuickReply) Patch(patch *QuickReplyPatch) {
	p.UpdatedAt = patch.UpdatedAt
	p.UpdatedBy = &patch.UpdatedBy

	if patch.Name != nil {
		p.Name = *patch.Name
	}

	if patch.Text != nil {
		p.Text = *patch.Text
	}

	if patch.Queue != nil {
		p.Queue = patch.Queue
	}

	if patch.Team != nil {
		p.Team = patch.Team
	}

	if patch.Article != nil {
		p.Article = patch.Article
	}
}

// Todo
func (r *QuickReply) IsValid() AppError {
	return nil
}
