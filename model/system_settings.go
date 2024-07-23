package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	SysNameOmnichannel            = "enable_omnichannel"
	SysNameMemberInsertChunkSize  = "member_chunk_size"
	SysNameSchemeVersionLimit     = "scheme_version_limit"
	SysNameAmdCancelNotHuman      = "amd_cancel_not_human"
	SysNameTwoFactorAuthorization = "enable_2fa"
	SysNamePasswordRegExp         = "password_reg_exp"
	SysNamePasswordValidationText = "password_validation_text"
)

type SysValue json.RawMessage

type SystemSetting struct {
	Id    int32           `json:"id" db:"id"`
	Name  string          `json:"name" db:"name"`
	Value json.RawMessage `json:"value" db:"value"`
}

type AvailableSystemSetting struct {
	Name string `json:"name" db:"name"`
}

type SystemSettingPath struct {
	Value json.RawMessage `json:"value" db:"value"`
}

type SearchSystemSetting struct {
	ListRequest
}

type AvailableSearchSystemSetting struct {
	ListRequest
}

func (v *SysValue) Str() *string {
	if v == nil {
		return nil
	}

	var val string
	err := json.Unmarshal(*v, &val)
	if err != nil {
		return nil
	}

	return &val
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
	case SysNameOmnichannel, SysNameAmdCancelNotHuman:
		return nil
	case SysNameMemberInsertChunkSize, SysNameSchemeVersionLimit:
		value := SysValue(s.Value)
		i := value.Int()

		if i == nil || *i < 1 {
			return NewBadRequestError("model.SystemSetting.valid.int.value", "The value should be more than 1")
		}
	case SysNameTwoFactorAuthorization:
		value := SysValue(s.Value)
		i := value.Bool()

		if i == nil {
			return NewBadRequestError("model.SystemSetting.invalid.bool.value", "invalid bool value")
		}
	case SysNamePasswordRegExp:
		value := SysValue(s.Value)
		str := value.Str()
		if str == nil || *str == "" {
			return NewBadRequestError("model.SystemSetting.invalid.str.value", "The value invalid string value")
		}
	case SysNamePasswordValidationText:
		value := SysValue(s.Value)
		str := value.Str()
		if str == nil || *str == "" {
			return NewBadRequestError("model.SystemSetting.invalid.str.value", "The value invalid string value")
		}
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

func (v *SysValue) Bool() *bool {
	if v == nil {
		return nil
	}

	i, err := strconv.ParseBool(string(*v))
	if err != nil {
		return nil
	}

	return &i
}
