package auth_manager

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	api "github.com/webitel/engine/pkg/wbt/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	LicenseCallManager = "CALL_MANAGER"
	LicenseCallCenter  = "CALL_CENTER"
	LicenseChat        = "CHAT"
	LicenseEmail       = "EMAIL"
	LicenseWFM         = "WFM"
)

func (am *authManager) ProductLimit(ctx context.Context, token string, productName string) (int, error) {
	header := metadata.New(map[string]string{"x-webitel-access": token})
	outCtx := metadata.NewOutgoingContext(ctx, header)
	tenant, err := am.customer.Api.GetCustomer(outCtx, &api.GetCustomerRequest{})

	if err != nil {
		return 0, err
	}

	if tenant.Customer == nil {
		return 0, errors.New("")
	}

	var limitMax int32

	for _, grant := range tenant.Customer.GetLicense() {
		if grant.Product != productName {
			continue // Lookup productName only !
		}
		if errs := grant.GetStatus().GetErrors(); len(errs) != 0 {
			// Also, ignore single 'product exhausted' (remain < 1) error
			// as we do not consider product user assignments here ...
			if !(len(errs) == 1 && errs[0] == "product exhausted") {
				continue // Currently invalid
			}
		}
		if limitMax < grant.Remain {
			limitMax = grant.Remain
		}
	}

	if limitMax == 0 {
		// FIXME: No CHAT product(s) issued !
		return 0, errors.New("")
	}

	return int(limitMax), nil
}

func (am *authManager) GetSession(c context.Context, token string) (*Session, error) {
	header := metadata.New(map[string]string{"x-webitel-access": token})
	ctx := metadata.NewOutgoingContext(c, header)

	resp, err := am.auth.Api.UserInfo(ctx, &api.UserinfoRequest{})

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
		RoleIds:    transformRoles(resp.UserId, resp.Roles), // /FIXME
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
				session.actions = append(session.actions, PermissionViewNumbers)
			case "playback_record_file":
				session.actions = append(session.actions, PermissionRecordFile)
			case "time_limited_record_file":
				session.actions = append(session.actions, PermissionTimeLimitedRecordFile)
			case "system_setting":
				session.actions = append(session.actions, PermissionSystemSetting)
			case "scheme_variables":
				session.actions = append(session.actions, PermissionSchemeVariables)
			case "reset_active_attempts":
				session.actions = append(session.actions, PermissionResetActiveAttempts)
			default:
				session.actions = append(session.actions, v.Id)
			}
		}
	}

	return session, nil
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
			// Abac:   v.Abac,
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
