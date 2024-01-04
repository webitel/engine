package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (a *App) SearchSchemeVersions(ctx context.Context, in *model.SearchSchemeVersion) ([]*model.SchemeVersion, bool, model.AppError) {

	filter := &model.Filter{
		Column:         model.SchemeVersionFields.Id,
		Value:          in.SchemeId,
		ComparisonType: model.Equal,
	}

	list, err := a.Store.SchemeVersion().Search(ctx, &in.ListRequest, filter)
	if err != nil {
		return nil, false, err
	}
	in.RemoveLastElemIfNeed(&list)
	return list, in.EndOfList(), nil
}
