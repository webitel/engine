package app

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (app *App) GetSession(token string) (*auth_manager.Session, model.AppError) {
	return app.GetSessionWitchContext(context.Background(), token)
}

func (app *App) GetSessionWitchContext(ctx context.Context, token string) (*auth_manager.Session, model.AppError) {
	session, err := app.sessionManager.GetSession(ctx, token)

	if err != nil {
		switch err {
		case auth_manager.ErrInternal:
			return nil, model.NewInternalError("app.session.app_error", err.Error())

		case auth_manager.ErrStatusForbidden:
			return nil, model.NewForbiddenError("app.session.forbidden", err.Error())

		case auth_manager.ErrStatusUnauthenticated:
			return nil, model.NewUnauthorizedError("app.session.unauthenticated", err.Error())

		case auth_manager.ErrValidId:
			return nil, model.NewBadRequestError("app.session.is_valid.id.app_error", err.Error())

		case auth_manager.ErrValidUserId:
			return nil, model.NewBadRequestError("app.session.is_valid.user_id.app_error", err.Error())

		case auth_manager.ErrValidToken:
			return nil, model.NewBadRequestError("app.session.is_valid.token.app_error", err.Error())

		case auth_manager.ErrValidRoleIds:
			return nil, model.NewBadRequestError("app.session.is_valid.role_ids.app_error", err.Error())

		}
	}

	if session.UserId == 0 {
		return nil, model.NewInternalError("app.session.not_found", err.Error())
	}

	return session, nil
}
