package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	SysNameOmnichannel                 = "enable_omnichannel"
	SysNameMemberInsertChunkSize       = "member_chunk_size"
	SysNameSchemeVersionLimit          = "scheme_version_limit"
	SysNameAmdCancelNotHuman           = "amd_cancel_not_human"
	SysNameTwoFactorAuthorization      = "enable_2fa"
	SysNameExportSettings              = "export_settings"
	SysNameSearchNumberLength          = "search_number_length"
	SysNameChatAiConnection            = "chat_ai_connection"
	SysNamePasswordRegExp              = "password_reg_exp"
	SysNamePasswordValidationText      = "password_validation_text"
	SysNameAutolinkCallToContact       = "autolink_call_to_contact"
	SysNamePeriodToPlaybackRecord      = "period_to_playback_records"
	SysNameIsFulltextSearchEnabled     = "is_fulltext_search_enabled"
	SysNameHideContact                 = "wbt_hide_contact"
	SysNameShowFullContact             = "show_full_contact"
	SysNameCallEndSoundNotification    = "call_end_sound_notification"
	SysNameCallEndPushNotification     = "call_end_push_notification"
	SysNameChatEndSoundNotification    = "chat_end_sound_notification"
	SysNameChatEndPushNotification     = "chat_end_push_notification"
	SysNameTaskEndSoundNotification    = "task_end_sound_notification"
	SysNameTaskEndPushNotification     = "task_end_push_notification"
	SysNamePushNotificationTimeout     = "push_notification_timeout"
	SysNameLabelsToLimitContacts       = "labels_to_limit_contacts"
	SysNameAutolinkMailToContact       = "autolink_mail_to_contact"
	SysNameNewMessageSoundNotification = "new_message_sound_notification"
	SysNameNewChatSoundNotification    = "new_chat_sound_notification"
	SysNameScreenshotInterval          = "screenshot_interval"
	SysNamePasswordExpiryDays          = "password_expiry_days"
	SysNamePasswordMinLength           = "password_min_length"
	SysNamePasswordCategories          = "password_categories"
	SysNamePasswordContainsLogin       = "password_contains_login"
	SysNamePasswordWarningDays         = "password_warning_days"
	SysNameDefaultPassword             = "default_password"
	SysNameExpandContactTabs           = "expand_contact_tabs"
	SysNameDefaultWorkspaceTab         = "default_workspace_tab"
	SysNameBlockAllMemberNumbers       = "block_all_member_numbers_from_list"
	SysNameLoginOptions                = "login_options"
	SysNameDefaultMembersFilter        = "default_members_filter"
)

type SysValue json.RawMessage

type SystemSetting struct {
	Id    int32           `json:"id" db:"id"`
	Name  string          `json:"name" db:"name"`
	Value json.RawMessage `json:"value" db:"value"`
}

var availableDefaultMembersFilterOptions = map[string]struct{}{
	"this day":   {},
	"this week":  {},
	"this month": {},
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

func validateDefaultMembersFilter(filter string) AppError {
	if _, exists := availableDefaultMembersFilterOptions[filter]; !exists {
		return NewBadRequestError(
			"model.system_settings.validate_default_members_filter",
			`default_members_filter can be only one of next values: "this day", "this week", "this month"`,
		)
	}

	return nil
}

func (s *SystemSetting) IsValid() AppError {
	switch s.Name {
	case SysNameOmnichannel, SysNameAmdCancelNotHuman:
		return nil
	case SysNameMemberInsertChunkSize, SysNameSchemeVersionLimit, SysNameSearchNumberLength,
		SysNamePeriodToPlaybackRecord, SysNamePushNotificationTimeout, SysNameScreenshotInterval,
		SysNamePasswordExpiryDays, SysNamePasswordMinLength, SysNamePasswordWarningDays:
		value := SysValue(s.Value)
		i := value.Int()

		if i == nil || *i < 1 {
			return NewBadRequestError("model.SystemSetting.invalid.int.value", "The value should be more than 1")
		}
	case SysNameChatAiConnection,
		SysNamePasswordRegExp,
		SysNamePasswordValidationText,
		SysNameDefaultPassword,
		SysNameDefaultWorkspaceTab,
		SysNameLoginOptions,
		SysNameDefaultMembersFilter:
		value := SysValue(s.Value)
		str := value.Str()

		if str == nil || strings.TrimSpace(*str) == "" {
			return NewBadRequestError("model.SystemSetting.invalid.str.value", "The value invalid string value")
		}

		if s.Name == SysNameDefaultMembersFilter {
			if werr := validateDefaultMembersFilter(*str); werr != nil {
				return werr
			}
		}
	case SysNameTwoFactorAuthorization,
		SysNameAutolinkCallToContact,
		SysNameAutolinkMailToContact,
		SysNameIsFulltextSearchEnabled,
		SysNameHideContact,
		SysNameShowFullContact,
		SysNameCallEndSoundNotification,
		SysNameCallEndPushNotification,
		SysNameChatEndSoundNotification,
		SysNameChatEndPushNotification,
		SysNameTaskEndSoundNotification,
		SysNameTaskEndPushNotification,
		SysNameNewMessageSoundNotification,
		SysNameNewChatSoundNotification,
		SysNamePasswordContainsLogin,
		SysNameExpandContactTabs,
		SysNameBlockAllMemberNumbers:
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
		var labels []struct {
			Label string `json:"label"`
		}
		err := json.Unmarshal(s.Value, &labels)
		if err != nil {
			return NewBadRequestError("model.SystemSetting.labels_to_limit_contacts.invalid.value", `value is not properly formed required: [{"label":"string"}]`)
		}
	case SysNamePasswordCategories:
		var categories []struct {
			Category string `json:"category"`
		}
		err := json.Unmarshal(s.Value, &categories)
		if err != nil {
			return NewBadRequestError("model.SystemSetting.password_categories.invalid.value", `value is not properly formed required: [{"category":"string"}]`)
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
