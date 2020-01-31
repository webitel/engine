package auth_manager

import (
	"context"
	"github.com/webitel/engine/auth_manager/auth"
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
	api    auth.SAClient
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

	connection.api = auth.NewSAClient(connection.client)

	return connection, nil
}

func (ac *authConnection) GetSession(token string) (*Session, error) {

	resp, err := ac.api.Current(context.TODO(), &auth.VerifyTokenRequest{token})

	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, ErrStatusUnauthenticated
		}
		return nil, ErrInternal
	}

	if resp.Session == nil {
		return nil, ErrStatusForbidden
	}

	session := &Session{
		Id:         resp.Session.Uuid,
		UserId:     resp.Session.UserId,
		DomainId:   resp.Session.Dc,
		DomainName: resp.Session.Domain,
		Expire:     resp.Session.ExpiresAt,
		Token:      token,
		Scopes:     transformScopes(resp.Scope),
		RoleIds:    transformRoles(resp.Roles),
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

func transformScopes(src []*auth.AccessScope) []SessionPermission {
	dst := make([]SessionPermission, 0, len(src))
	for _, v := range src {
		dst = append(dst, SessionPermission{
			Id:   int(v.Id),
			Name: v.Class,
			//Abac:   v.Abac,
			Obac:   v.Obac,
			Rbac:   v.Rbac,
			Access: v.Access,
		})
	}
	return dst
}

func transformRoles(src map[string]int64) []int {
	dst := make([]int, 0, len(src))
	for _, v := range src {
		dst = append(dst, int(v))
	}
	return dst
}
