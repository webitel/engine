package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateChatTag(ctx context.Context, domainId int64, teamId int64, tag *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {
	return app.Store.TeamChatTag().Create(ctx, domainId, teamId, tag)
}

func (app *App) SearchChatTag(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, bool, model.AppError) {
	list, err := app.Store.TeamChatTag().GetAllPage(ctx, domainId, teamId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetChatTag(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamChatTag, model.AppError) {
	return app.Store.TeamChatTag().Get(ctx, domainId, teamId, id)
}

func (app *App) UpdateChatTag(ctx context.Context, domainId int64, teamId int64, tag *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {
	oldTag, err := app.GetChatTag(ctx, domainId, teamId, tag.Id)
	if err != nil {
		return nil, err
	}

	oldTag.Tag = tag.Tag
	oldTag.UpdatedAt = tag.UpdatedAt
	oldTag.UpdatedBy = tag.UpdatedBy

	oldTag, err = app.Store.TeamChatTag().Update(ctx, domainId, teamId, oldTag)
	if err != nil {
		return nil, err
	}

	return oldTag, nil
}

func (app *App) RemoveChatTag(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamChatTag, model.AppError) {
	tag, err := app.GetChatTag(ctx, domainId, teamId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.TeamChatTag().Delete(ctx, domainId, teamId, id)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
