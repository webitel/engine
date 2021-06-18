package auth_manager

type SessionPermission struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	//Abac   bool   `json:"abac"`
	Obac   bool   `json:"obac"`
	rbac   bool   `json:"rbac"`
	Access uint32 `json:"access"`
}

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

func (s *Session) Domain(def int64) int64 {
	return s.DomainId
}

func (s SessionPermission) CanCreate() bool {
	if s.Obac || s.rbac {
		return s.Access&PERMISSION_ACCESS_CREATE.Value() == PERMISSION_ACCESS_CREATE.Value()
	}
	return !s.rbac && !s.Obac
}

func (s SessionPermission) CanRead() bool {
	if s.Obac || s.rbac {
		return s.Access&PERMISSION_ACCESS_READ.Value() == PERMISSION_ACCESS_READ.Value()
	}
	return !s.rbac && !s.Obac
}

func (s SessionPermission) CanUpdate() bool {
	if s.Obac || s.rbac {
		return s.Access&PERMISSION_ACCESS_UPDATE.Value() == PERMISSION_ACCESS_UPDATE.Value()
	}
	return !s.rbac && !s.Obac
}

func (s SessionPermission) CanDelete() bool {
	if s.Obac || s.rbac {
		return s.Access&PERMISSION_ACCESS_DELETE.Value() == PERMISSION_ACCESS_DELETE.Value()
	}
	return !s.rbac && !s.Obac
}
