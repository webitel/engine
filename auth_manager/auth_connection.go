package auth_manager

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"

	"time"

	"github.com/webitel/engine/auth_manager/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

type AuthClient interface {
	Name() string
	Close() error
	Ready() bool
	GetSession(token string) (*Session, error)
}

const (
	AUTH_CONNECTION_TIMEOUT = 2 * time.Second
)

const (
	LicenseCallManager = "CALL_MANAGER"
	LicenseCallCenter  = "CALL_CENTER"
)

type authConnection struct {
	name   string
	host   string
	client *grpc.ClientConn
	api    api.AuthClient
}

func NewAuthServiceConnection(name, url string) (AuthClient, error) {
	var err error
	connection := &authConnection{
		name: name,
		host: url,
	}

	connection.client, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(AUTH_CONNECTION_TIMEOUT))

	if err != nil {
		return nil, err
	}

	connection.api = api.NewAuthClient(connection.client)

	return connection, nil
}

func (ac *authConnection) GetSession(token string) (*Session, error) {
	//FIXME
	header := metadata.New(map[string]string{"x-webitel-access": token})
	ctx := metadata.NewOutgoingContext(context.TODO(), header)

	resp, err := ac.api.UserInfo(ctx, &api.UserinfoRequest{})

	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, ErrStatusUnauthenticated
		}

		return nil, ErrInternal
	}

	if resp == nil {
		return nil, ErrStatusUnauthenticated
	}

	session := &Session{
		Id:         token,
		UserId:     resp.UserId,
		DomainId:   resp.Dc,
		DomainName: resp.Domain,
		Expire:     resp.ExpiresAt,
		Token:      token,
		RoleIds:    transformRoles(resp.UserId, resp.Roles), ///FIXME
		Scopes:     transformScopes(resp.Scope),
		actions:    make([]string, 0, 1),
		Name:       resp.Name,
	}

	session.validLicense, session.active = licenseActiveScope(resp)

	if len(resp.Permissions) > 0 {
		session.adminPermissions = make([]PermissionAccess, len(resp.Permissions), len(resp.Permissions))
		for _, v := range resp.Permissions {
			switch v.Id {
			case "add":
				session.adminPermissions = append(session.adminPermissions, PERMISSION_ACCESS_CREATE)
			case "read":
				session.adminPermissions = append(session.adminPermissions, PERMISSION_ACCESS_READ)
			case "write":
				session.adminPermissions = append(session.adminPermissions, PERMISSION_ACCESS_UPDATE)
			case "delete":
				session.adminPermissions = append(session.adminPermissions, PERMISSION_ACCESS_DELETE)
			case "view_cdr_phone_numbers":
				session.actions = append(session.actions, PERMISSION_VIEW_NUMBERS)
			case "playback_record_file":
				session.actions = append(session.actions, PERMISSION_RECORD_FILE)
			}
		}
	}

	return session, nil
}

func (ac *authConnection) Ready() bool {
	switch ac.client.GetState() {
	case connectivity.Idle, connectivity.Ready:
		return true
	}
	return false
}

func (ac *authConnection) Name() string {
	return ac.name
}

func (ac *authConnection) Close() error {
	err := ac.client.Close()
	if err != nil {
		return ErrInternal
	}
	return nil
}

// returns the provided original scope
// from all license products assigned to user
//
// NOTE: include <readonly> access
//
//	{ obac:true, access:"r" }
func transformScopes(src []*api.Objclass) []SessionPermission {
	dst := make([]SessionPermission, 0, len(src))
	var access int
	for _, v := range src {
		access, _ = parseAccess(v.Access) //
		dst = append(dst, SessionPermission{
			Id:   int(v.Id),
			Name: v.Class,
			//Abac:   v.Abac,
			Obac:   v.Obac,
			rbac:   v.Rbac,
			Access: uint32(access),
		})
	}
	return dst
}

// returns the scope from all license products
// active now within their validity boundaries
func licenseActiveScope(src *api.Userinfo) ([]string, []string) {
	var (
		l           = len(src.License)
		validLicene = make([]string, 0, l)
		now         = time.Now().UnixMilli()
		scope       = make([]string, 0, len(src.GetScope()))
		// canonical name transformations
		objClass = func(name string) string {
			name = strings.TrimSpace(name)
			name = strings.ToLower(name)
			return name
		}
		// indicates whether such `name` exists in scope
		hasScope = func(name string) bool {
			if len(scope) == 0 {
				return name == ""
			}
			// name = objClass(name) // CaseIgnoreMatch(!)
			if len(name) == 0 {
				return true // len(scope) != 0
			}
			e, n := 0, len(scope)
			for ; e < n && scope[e] != name; e++ {
				// break; match found !
			}
			return e < n
		}
		// add unique `setof` to the scope
		addScope = func(setof []string) {
			var name string
			for _, class := range setof {
				name = objClass(class) // CaseIgnoreMatch(!)
				if len(name) == 0 {
					continue
				}
				if !hasScope(name) {
					scope = append(scope, name)
				}
			}
		}
	)
	// gather active only products scopes
	for _, prod := range src.License {
		if len(prod.Scope) == 0 {
			continue // forceless
		}
		if 0 < prod.ExpiresAt && prod.ExpiresAt <= now {
			// Expired ! Grant READONLY access
		} else if 0 < prod.IssuedAt && now < prod.IssuedAt {
			// Inactive ! No access grant yet !
		} else {
			// Active ! +OK
			addScope(prod.Scope)
			validLicene = append(validLicene, prod.Prod)
		}
	}

	if len(scope) == 0 {
		// ALL License Product(s) are inactive !
		return nil, nil
	}

	var (
		objclass        string
		e, n            = 0, len(src.Scope)
		caseIgnoreMatch = strings.EqualFold
	)
	for i := 0; i < len(scope); i++ {
		objclass = scope[i]
		for e = 0; e < n && !caseIgnoreMatch(src.Scope[e].Class, objclass); e++ {
			// Lookup for caseIgnoreMatch(!) with userinfo.Scope OBAC grants
		}
		if e == n {
			// NOT FOUND ?! OBAC Policy: Access Denied ?!
			scope = append(scope[0:i], scope[i+1:]...)
			i--
			continue
		}
	}
	return validLicene, scope
}

func transformRoles(userId int64, src []*api.ObjectId) []int {
	dst := make([]int, 0, len(src)+1)
	dst = append(dst, int(userId))
	for _, v := range src {
		dst = append(dst, int(v.Id))
	}
	return dst
}

func parseAccess(s string) (grants int, err error) {
	// grants = 0 // NoAccess
	var grant int
	for _, c := range s {
		switch c {
		case 'x':
			grant = 8 // XMode
		case 'r':
			grant = 4 // ReadMode
		case 'w':
			grant = 2 // WriteMode
		case 'd':
			grant = 1 // DeleteMode
		default:
			return 0, ErrValidScope
		}
		if (grants & grant) == grant { // grants.HasMode(grant)
			grants |= (grant << 4) // grants.GrantMode(grant)
			continue
		}
		grants |= grant // grants.SetMode(grant)
	}
	return grants, nil
}
