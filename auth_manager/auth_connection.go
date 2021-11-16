package auth_manager

import (
	"context"
	"google.golang.org/grpc/metadata"

	"github.com/webitel/engine/auth_manager/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	"time"
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
		Scopes:     transformScopes(resp.Scope),
		RoleIds:    transformRoles(resp.UserId, resp.Roles), ///FIXME
		actions:    make([]string, 0, 1),
	}

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
