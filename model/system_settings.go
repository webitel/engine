package model

import "encoding/json"

type SystemSetting struct {
	Id    int32           `json:"id" db:"id"`
	Name  string          `json:"name" db:"name"`
	Value json.RawMessage `json:"value" db:"value"`
}

type SystemSettingPath struct {
	Value json.RawMessage `json:"value" db:"value"`
}

type SearchSystemSetting struct {
	ListRequest
}

func (SystemSetting) DefaultOrder() string {
	return "name"
}

func (SystemSetting) AllowFields() []string {
	return []string{"id", "name", "value"}
}

func (s SystemSetting) DefaultFields() []string {
	return s.AllowFields()
}

func (SystemSetting) EntityName() string {
	return "system_settings"
}

func (s *SystemSetting) IsValid() AppError {
	//FIXME
	return nil
}

func (s *SystemSetting) Patch(p *SystemSettingPath) {
	if p.Value != nil {
		s.Value = p.Value
	}
}
