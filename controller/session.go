package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (ctrl *Controller) GetSessionFromCtx(ctx context.Context) (*auth_manager.Session, model.AppError) {
	return ctrl.app.GetSessionFromCtx(ctx)
}
