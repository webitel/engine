package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	SysNameOmnichannel              = "enable_omnichannel"
	SysNameMemberInsertChunkSize    = "member_chunk_size"
	SysNameSchemeVersionLimit       = "scheme_version_limit"
	SysNameAmdCancelNotHuman        = "amd_cancel_not_human"
	SysNameTwoFactorAuthorization   = "enable_2fa"
	SysNameExportSettings           = "export_settings"
	SysNameSearchNumberLength       = "search_number_length"
	SysNameChatAiConnection         = "chat_ai_connection"
	SysNamePasswordRegExp           = "password_reg_exp"
	SysNamePasswordValidationText   = "password_validation_text"
	SysNameAutolinkCallToContact    = "autolink_call_to_contact"
	SysNamePeriodToPlaybackRecord   = "period_to_playback_records"
	SysNameIsFulltextSearchEnabled  = "is_fulltext_search_enabled"
	SysNameHideContact              = "wbt_hide_contact"
	SysNameShowFullContact          = "show_full_contact"
	SysNameCallEndSoundNotification = "call_end_sound_notification"
	SysNameCallEndPushNotification  = "call_end_push_notification"
	SysNameChatEndSoundNotification = "chat_end_sound_notification"
	SysNameChatEndPushNotification  = "chat_end_push_notification"
	SysNameTaskEndSoundNotification = "task_end_sound_notification"
	SysNameTaskEndPushNotification  = "task_end_push_notification"
	SysNamePushNotificationTimeout  = "push_notification_timeout"
	SysNameLabelsToLimitContacts    = "labels_to_limit_contacts"
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
	Name []string
}

type AvailableSearchSystemSetting struct {
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
	case SysNameOmnichannel, SysNameAmdCancelNotHuman:
		return nil
	case SysNameMemberInsertChunkSize, SysNameSchemeVersionLimit, SysNameSearchNumberLength, SysNamePeriodToPlaybackRecord, SysNamePushNotificationTimeout:
		value := SysValue(s.Value)
		i := value.Int()

		if i == nil || *i < 1 {
			return NewBadRequestError("model.SystemSetting.invalid.int.value", "The value should be more than 1")
		}
	case SysNameChatAiConnection,
		SysNamePasswordRegExp,
		SysNamePasswordValidationText:
		value := SysValue(s.Value)
		str := value.Str()
		if str == nil || *str == "" {
			return NewBadRequestError("model.SystemSetting.invalid.str.value", "The value invalid string value")
		}
	case SysNameTwoFactorAuthorization,
		SysNameAutolinkCallToContact,
		SysNameIsFulltextSearchEnabled,
		SysNameHideContact,
		SysNameShowFullContact,
		SysNameCallEndSoundNotification,
		SysNameCallEndPushNotification,
		SysNameChatEndSoundNotification,
		SysNameChatEndPushNotification,
		SysNameTaskEndSoundNotification,
		SysNameTaskEndPushNotification:
		value := SysValue(s.Value)
		i := value.Bool()

		if i == nil {
			return NewBadRequestError("model.SystemSetting.invalid.bool.value", "invalid bool value")
		}
	case SysNameExportSettings:
		export := struct {
			Format    string `json:"format,omitempty"`
			Separator string `json:"separator,omitempty"`
		}{}
		err := json.Unmarshal(s.Value, &export)
		if err != nil {
			return NewBadRequestError("model.SystemSetting.export_settings.invalid.value", "value is not properly formed")
		}
	case SysNameLabelsToLimitContacts:
		var lookups []struct {
			Id   int32  `json:"id"`
			Name string `json:"name"`
		}
		err := json.Unmarshal(s.Value, &lookups)
		if err != nil {
			return NewBadRequestError("model.SystemSetting.labels_to_limit_contacts.invalid.value", `value is not properly formed required: [{"id": "string", "name": "string"}]`)
		}
	default:
		return NewBadRequestError("model.SystemSetting.invalid_value", fmt.Sprintf("%s is not allowed", s.Name))
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
