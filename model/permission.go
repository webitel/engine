package model

const (
	PERMISSION_SCOPE_CALL                       = "calls"
	PERMISSION_SCOPE_CALENDAR                   = "calendars"
	PERMISSION_SCOPE_CC_TEAM                    = "cc_team"
	PERMISSION_SCOPE_CC_AGENT                   = "cc_agent"
	PERMISSION_SCOPE_CC_QUEUE                   = "cc_queue"
	PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE       = "cc_resource"
	PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP = "cc_resource_group"
	PERMISSION_SCOPE_CC_LIST                    = "cc_list"
	PERMISSION_SCOPE_CC_LIST_NUMBER             = "cc_list_number"
	PERMISSION_SCOPE_ACR_ROUTING                = "acr_routing"   //rename
	PERMISSION_SCOPE_ACR_CHAT_PLAN              = "acr_chat_plan" //"acr_chat_plan"

	PERMISSION_SCOPE_SCHEMA       = "schema"
	PERMISSION_SCOPE_DICTIONARIES = "dictionaries"

	PERMISSION_SCOPE_EMAIL_PROFILE = "email_profile"

	PERMISSION_SCOPE_USERS   = "users"
	PERMISSION_SCOPE_TRIGGER = "trigger"

	PermissionAuditFrom  = "cc_audit_form" // "cc_form"
	PermissionAuditRate  = "rating"
	PermissionRecordFile = "record_file"
	PermissionSkill      = "cc_skill"
	PermissionWebHook    = "email_profile" // todo
	PermissionContacts   = "contacts"

	PermissionControlAgentScreen = "control_agent_screen"
)
