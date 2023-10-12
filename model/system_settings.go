package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	SysNameOmnichannel           = "enable_omnichannel"
	SysNameMemberInsertChunkSize = "member_chunk_size"
)

type SysValue json.RawMessage

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
	switch s.Name {
	case SysNameOmnichannel:
	case SysNameMemberInsertChunkSize:
		break
	default:
		return NewBadRequestError("model.SystemSetting.valid.name", fmt.Sprintf("%s not allow", s.Name))
	}
	return nil
}

func (s *SystemSetting) Patch(p *SystemSettingPath) {
	if p.Value != nil {
		s.Value = p.Value
	}
}

func (v *SysValue) Int() *int {
	if v == nil {
		return nil
	}

	i, err := strconv.Atoi(string(*v))
	if err != nil {
		return nil
	}

	return &i
}
