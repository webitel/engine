package app

import (
	"bytes"
	"context"
	"github.com/webitel/engine/model"
)

func (a *App) CreateSchemaVariable(ctx context.Context, domainId int64, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	var err model.AppError
	if variable.Encrypt {
		variable.Value, err = a.EncryptBytes(variable.Value)
		if err != nil {
			return nil, err
		}
		var buffer bytes.Buffer
		buffer.WriteString(`"`)
		buffer.Write(variable.Value)
		buffer.WriteString(`"`)
		variable.Value = buffer.Bytes()
	}
	return a.Store.SchemeVariable().Create(ctx, domainId, variable)
}

func (a *App) SearchSchemeVariable(ctx context.Context, domainId int64, search *model.SearchSchemeVariable) ([]*model.SchemeVariable, bool, model.AppError) {

	list, err := a.Store.SchemeVariable().Search(ctx, domainId, &search.ListRequest, nil)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetSchemeVariable(ctx context.Context, domainId int64, id int32) (*model.SchemeVariable, model.AppError) {
	return a.Store.SchemeVariable().Get(ctx, domainId, id)
}

func (a *App) UpdateSchemaVariable(ctx context.Context, domainId int64, id int32, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError) {
	old, err := a.GetSchemeVariable(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	old.Name = variable.Name
	if old.Encrypt {
		if len(variable.Value) != 0 {
			old.Value, err = a.EncryptBytes(variable.Value)
		}
	} else {
		old.Value = variable.Value
	}

	if err = old.IsValid(); err != nil {
		return nil, err
	}

	old, err = a.Store.SchemeVariable().Update(ctx, domainId, old)
	if err != nil {
		return nil, err
	}

	return old, nil
}

func (a *App) DeleteSchemaVariable(ctx context.Context, domainId int64, id int32) (*model.SchemeVariable, model.AppError) {
	old, err := a.GetSchemeVariable(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	err = a.Store.SchemeVariable().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	return old, nil
}
