package model

type CommunicationType struct {
	Id          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Code        string `json:"code" db:"code"`
	Channel     string `json:"channel" db:"channel"`
	Description string `json:"description" db:"description"`
	Default     bool   `json:"default" db:"default"`
}

type CommunicationTypePatch struct {
	Name        *string `json:"name" db:"name"`
	Code        *string `json:"code" db:"code"`
	Channel     *string `json:"channel" db:"channel"`
	Description *string `json:"description" db:"description"`
	Default     *bool   `json:"default" db:"default"`
}

type SearchCommunicationType struct {
	ListRequest
	Ids      []uint32
	Channels []string
	Default  *bool
}

func (CommunicationType) DefaultOrder() string {
	return "id"
}

func (a CommunicationType) AllowFields() []string {
	return []string{"id", "name", "code", "description", "channel", "default"}
}

func (a CommunicationType) DefaultFields() []string {
	return []string{"id", "name", "code", "description", "channel", "default"}
}

func (a CommunicationType) EntityName() string {
	return "cc_communication_list"
}

func (s *CommunicationType) IsValid() AppError {
	return nil
}

func (q *CommunicationType) Patch(patch *CommunicationTypePatch) {
	if patch.Default != nil {
		q.Default = *patch.Default
	}
	if patch.Name != nil {
		q.Name = *patch.Name
	}
	if patch.Description != nil {
		q.Description = *patch.Description
	}
	if patch.Code != nil {
		q.Code = *patch.Code
	}
	if patch.Channel != nil {
		q.Channel = *patch.Channel
	}
}
