package model

import "time"

type QuickReply struct {
	AclRecord
	Id      int       `json:"id" db:"id"`
	Name    string    `json:"name" db:"name"`
	Text    string    `json:"text" db:"text"`
	Queues  []*Lookup `json:"queue" db:"queues"`
	Teams   []*Lookup `json:"team" db:"teams"`
	Article *Lookup   `json:"article" db:"article"`
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
	Queues    []*Lookup  `json:"queue" db:"queues"`
	Teams     []*Lookup  `json:"team" db:"teams"`
	Article   *Lookup    `json:"article" db:"article"`
}

func (p QuickReply) AllowFields() []string {
	return []string{"id", "domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "text", "teams", "queues"}
}

func (QuickReply) DefaultOrder() string {
	return "-name"
}

func (QuickReply) DefaultFields() []string {
	return []string{"id", "created_at", "created_by", "updated_at", "updated_by", "name", "text", "teams", "queues"}
}

func (QuickReply) EntityName() string {
	return "cc_quick_reply_list"
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

	if patch.Queues != nil {
		p.Queues = patch.Queues
	}

	if patch.Teams != nil {
		p.Teams = patch.Teams
	}

	if patch.Article != nil {
		p.Article = patch.Article
	}
}

// Todo
func (r *QuickReply) IsValid() AppError {
	return nil
}
