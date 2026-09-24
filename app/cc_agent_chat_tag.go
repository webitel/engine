package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) SearchAgentChatTag(ctx context.Context, domainId int64, userId int64, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, bool, model.AppError) {
	list, err := app.Store.TeamChatTag().GetAllPageByUser(ctx, domainId, userId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
