package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

type TokenLocation int

const (
	TokenLocationNotFound = iota
	TokenLocationHeader
	TokenLocationQueryString
)

func ParseAuthTokenFromRequest(r *http.Request) (string, TokenLocation) {
	authHeader := r.Header.Get(model.HEADER_AUTH)
	if len(authHeader) > 6 && strings.ToUpper(authHeader[0:6]) == model.HEADER_BEARER {
		// Default session token
		return authHeader[7:], TokenLocationHeader
	} else if len(authHeader) > 14 && authHeader[0:14] == model.HEADER_TOKEN {
		// OAuth token
		return authHeader[15:], TokenLocationHeader
	}

	// Attempt to parse token out of the query string
	if token := r.URL.Query().Get("access_token"); token != "" {
		return token, TokenLocationQueryString
	}

	return "", TokenLocationNotFound
}

func (a *App) MakePermissionError(session *auth_manager.Session, permission auth_manager.SessionPermission, access auth_manager.PermissionAccess) model.AppError {

	return model.NewForbiddenError("api.context.permissions.app_error", fmt.Sprintf("userId=%d, permission=%s access=%s", session.UserId, permission.Name, access.Name()))
}

func (a *App) MakeResourcePermissionError(session *auth_manager.Session, id int64, permission auth_manager.SessionPermission, access auth_manager.PermissionAccess) model.AppError {

	return model.NewForbiddenError("api.context.permissions.app_error", fmt.Sprintf("userId=%d, id=%d permission=%s access=%s", session.UserId, id, permission.Name, access.Name()))
}

func (a *App) MakeActionPermissionError(session *auth_manager.Session, action string, access auth_manager.PermissionAccess) model.AppError {

	return model.NewForbiddenError("api.context.action.permissions", fmt.Sprintf("userId=%d, action=%s access=%s", session.UserId, action, access.Name()))
}
