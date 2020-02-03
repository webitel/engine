package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (ctrl *Controller) GetSessionFromCtx(ctx context.Context) (*auth_manager.Session, *model.AppError) {
	return ctrl.app.GetSessionFromCtx(ctx)
}
