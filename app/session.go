package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"net/http"
)

func (app *App) GetSession(token string) (*auth_manager.Session, *model.AppError) {
	session, err := app.sessionManager.GetSession(token)

	if err != nil {
		switch err {
		case auth_manager.ErrInternal:
			return nil, model.NewAppError("App.GetSession", "app.session.app_error", nil, err.Error(), http.StatusInternalServerError)

		case auth_manager.ErrStatusForbidden:
			return nil, model.NewAppError("App.GetSession", "app.session.forbidden", nil, err.Error(), http.StatusInternalServerError)

		case auth_manager.ErrValidId:
			return nil, model.NewAppError("App.GetSession", "app.session.is_valid.id.app_error", nil, err.Error(), http.StatusInternalServerError)

		case auth_manager.ErrValidUserId:
			return nil, model.NewAppError("App.GetSession", "app.session.is_valid.user_id.app_error", nil, err.Error(), http.StatusInternalServerError)

		case auth_manager.ErrValidToken:
			return nil, model.NewAppError("App.GetSession", "app.session.is_valid.token.app_error", nil, err.Error(), http.StatusInternalServerError)

		case auth_manager.ErrValidRoleIds:
			return nil, model.NewAppError("App.GetSession", "app.session.is_valid.role_ids.app_error", nil, err.Error(), http.StatusInternalServerError)

		}
	}

	return session, nil
}
