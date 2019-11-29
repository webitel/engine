package model

const (
	PERMISSION_SCOPE_CALENDAR                   = "calendars"
	PERMISSION_SCOPE_CC_TEAM                    = "cc_team"
	PERMISSION_SCOPE_CC_AGENT                   = "cc_agent"
	PERMISSION_SCOPE_CC_QUEUE                   = "cc_queue"
	PERMISSION_SCOPE_CC_BUCKET                  = "cc_bucket"
	PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE       = "cc_resource"
	PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE_GROUP = "cc_resource_group"
	PERMISSION_SCOPE_CC_LIST                    = "cc_list"
	PERMISSION_SCOPE_ACR_ROUTING                = "acr_routing"
)

type PermissionAccess uint8

const (
	PERMISSION_ACCESS_CREATE PermissionAccess = iota
	PERMISSION_ACCESS_READ
	PERMISSION_ACCESS_UPDATE
	PERMISSION_ACCESS_DELETE
)

func (p PermissionAccess) Value() uint32 {
	return [...]uint32{8, 4, 2, 1}[p]
}

func (p PermissionAccess) Name() string {
	return [...]string{"create", "read", "update", "delete"}[p]
}
