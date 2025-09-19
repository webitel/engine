package auth_manager

import (
	"context"
	"slices"
	"time"

	"github.com/webitel/wlog"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"
)

var (
	sessionGroupRequest singleflight.Group
)

const tokenRequestTimeout = time.Second * 15

type Session struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	DomainId   int64         `json:"domain_id"`
	DomainName string        `json:"domain_name"`
	Expire     int64         `json:"expire"`
	UserId     int64         `json:"user_id"`
	userIp     atomic.String `json:"user_ip"`
	RoleIds    []int         `json:"role_ids"`

	Token            string              `json:"token"`
	Scopes           []SessionPermission `json:"scopes"`
	active           []string            `json:"-"`
	adminPermissions []PermissionAccess
	actions          []string
	validLicense     []string
}

func (self *Session) UseRBAC(acc PermissionAccess, perm SessionPermission) bool {
	if !perm.rbac {
		return false
	}

	for _, v := range self.adminPermissions {
		if v == acc {
			return false
		}
	}

	return perm.rbac
}

func (s *Session) HasAdminPermission(permAccess PermissionAccess) bool {
	return slices.Contains(s.adminPermissions, permAccess)
}

func (self *Session) GetAclRoles() []int {
	return self.RoleIds
}

func (self *Session) HasLicense(name string) bool {
	for _, v := range self.validLicense {
		if v == name {
			return true
		}
	}

	return false
}

func (self *Session) GetUserId() int64 {
	return self.UserId
}

func (self *Session) GetDomainId() int64 {
	return self.DomainId
}

func (self *Session) SetIp(ip string) {
	self.userIp.Store(ip)
}

func (self *Session) GetUserIp() string {
	return self.userIp.Load()
}

func (self *Session) HasCallCenterLicense() bool {
	return self.HasLicense(LicenseCallCenter)
}

func (self *Session) HasChatLicense() bool {
	return self.HasLicense(LicenseChat)
}

func (self *Session) CountLicenses() int {
	return len(self.active)
}

func (self *Session) GetPermission(name string) SessionPermission {
	for _, v := range self.Scopes {
		if v.Name == name {
			return v
		}
	}
	return NotAllowPermission(name)
}

func NotAllowPermission(name string) SessionPermission {
	return SessionPermission{
		Id:     0,
		Name:   name,
		Obac:   true,
		rbac:   true,
		Access: 0,
	}
}

// GetMillis is a convenience method to get milliseconds since epoch.
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (self *Session) IsExpired() bool {
	if self.Expire > 0 {
		return self.Expire*1000 < GetMillis()
	}
	return false //never expire
}

func (self *Session) Trace() map[string]interface{} {
	return map[string]interface{}{"id": self.Id, "domain_id": self.DomainId}
}

func (self *Session) IsValid() error {

	if len(self.Id) < 1 {
		return ErrValidId
	}
	if self.UserId < 1 {
		return ErrValidUserId
	}
	if len(self.Token) < 1 {
		return ErrValidToken
	}

	//if self.DomainId < 1 {
	//	return model.NewBadRequestError("model.session.is_valid.domain_id.app_error", "").SetTranslationParams(self.Trace())
	//}

	if len(self.RoleIds) < 1 {
		return ErrValidRoleIds
	}

	return nil
}

func (self *Session) HasAction(name string) bool {
	for _, v := range self.actions {
		if v == name {
			return true
		}
	}

	return false
}

func (am *authManager) getSession(c context.Context, token string) (Session, error) {

	if v, ok := am.session.Get(token); ok {
		return *v, nil
	}

	result, err, shared := sessionGroupRequest.Do(token, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(c, tokenRequestTimeout)
		defer cancel()

		return am.GetSession(ctx, token)
	})

	if err != nil {
		return Session{}, err
	}

	session := result.(*Session)

	if !shared {
		am.session.Add(token, session)
		am.log.With(wlog.String("user_name", session.Name)).Debug("store")
	}

	return *session, nil
}
