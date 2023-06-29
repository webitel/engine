package model

type ChatPlan struct {
	Id          int32  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	Enabled     bool   `json:"enabled" db:"enabled"`
	Name        string `json:"name" db:"name"`
	Schema      Lookup `json:"schema" db:"schema"`
	Description string `json:"description" db:"description"`
}

type PatchChatPlan struct {
	Enabled     *bool   `json:"enabled" db:"enabled"`
	Name        *string `json:"name" db:"name"`
	Schema      *Lookup `json:"schema" db:"schema"`
	Description *string `json:"description" db:"description"`
}

type SearchChatPlan struct {
	ListRequest
	Ids     []int32
	Name    *string
	Enabled *bool
}

func (c *ChatPlan) Patch(patch *PatchChatPlan) {
	if patch.Schema != nil {
		c.Schema = *patch.Schema
	}

	if patch.Enabled != nil {
		c.Enabled = *patch.Enabled
	}

	if patch.Name != nil {
		c.Name = *patch.Name
	}

	if patch.Description != nil {
		c.Description = *patch.Description
	}
}

func (ChatPlan) DefaultOrder() string {
	return "id"
}

func (ChatPlan) AllowFields() []string {
	return []string{"id", "domain_id", "enabled", "name", "schema", "description"}
}

func (ChatPlan) DefaultFields() []string {
	return []string{"id", "enabled", "name", "schema", "description"}
}

func (ChatPlan) EntityName() string {
	return "acr_chat_plan_list"
}

func (c *ChatPlan) IsValid() AppError {
	return nil
}
