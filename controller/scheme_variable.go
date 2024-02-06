package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

// TODO PermissionSchemeVariable
func (c *Controller) CreateSchemeVariable(ctx context.Context, session *auth_manager.Session, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	//if !session.HasAction(auth_manager.PermissionSchemeVariable) {
	//	return nil, false, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	//}

	return c.app.CreateSchemaVariable(ctx, session.Domain(0), variable)
}

// TODO PermissionSchemeVariable
func (c *Controller) SearchSchemeVariable(ctx context.Context, session *auth_manager.Session, search *model.SearchSchemeVariable) ([]*model.SchemeVariable, bool, model.AppError) {
	//if !session.HasAction(auth_manager.PermissionSchemeVariable) {
	//	return nil, false, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	//}

	return c.app.SearchSchemeVariable(ctx, session.Domain(0), search)
}

// TODO PermissionSchemeVariable
func (c *Controller) GetSchemeVariable(ctx context.Context, session *auth_manager.Session, id int32) (*model.SchemeVariable, model.AppError) {
	//if !session.HasAction(auth_manager.PermissionSchemeVariable) {
	//	return nil, false, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	//}

	return c.app.GetSchemeVariable(ctx, session.Domain(0), id)
}

// TODO PermissionSchemeVariable
func (c *Controller) UpdateSchemaVariable(ctx context.Context, session *auth_manager.Session, id int32, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	//if !session.HasAction(auth_manager.PermissionSchemeVariable) {
	//	return nil, false, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	//}

	return c.app.UpdateSchemaVariable(ctx, session.Domain(0), id, variable)
}

// TODO PermissionSchemeVariable
func (c *Controller) DeleteSchemaVariable(ctx context.Context, session *auth_manager.Session, id int32) (*model.SchemeVariable, model.AppError) {
	//if !session.HasAction(auth_manager.PermissionSchemeVariable) {
	//	return nil, false, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	//}

	return c.app.DeleteSchemaVariable(ctx, session.Domain(0), id)
}
