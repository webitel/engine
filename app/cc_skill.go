package app

import (
	"context"
	"github.com/webitel/engine/pkg/wbt/auth_manager"

	"github.com/webitel/engine/model"
)

func (app *App) SkillCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return app.Store.Skill().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) CreateSkill(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError) {
	return app.Store.Skill().Create(ctx, skill)
}

func (app *App) GetSkill(ctx context.Context, id, domainId int64) (*model.Skill, model.AppError) {
	return app.Store.Skill().Get(ctx, domainId, id)
}

func (app *App) GetSkillsPage(ctx context.Context, domainId int64, search *model.SearchSkill) ([]*model.Skill, bool, model.AppError) {
	list, err := app.Store.Skill().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetSkillsPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchSkill) ([]*model.Skill, bool, model.AppError) {
	list, err := app.Store.Skill().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) RemoveSkill(ctx context.Context, domainId, id int64) (*model.Skill, model.AppError) {
	skill, err := app.Store.Skill().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.Skill().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return skill, nil
}

func (app *App) UpdateSkill(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError) {
	oldSkill, err := app.Store.Skill().Get(ctx, skill.DomainId, skill.Id)

	if err != nil {
		return nil, err
	}

	oldSkill.Name = skill.Name
	oldSkill.Description = skill.Description
	oldSkill.UpdatedBy = skill.UpdatedBy
	oldSkill.UpdatedAt = skill.UpdatedAt

	_, err = app.Store.Skill().Update(ctx, oldSkill)
	if err != nil {
		return nil, err
	}

	return oldSkill, nil
}

func (app *App) SearchSkillAgent(ctx context.Context, domainId int64, skillId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, bool, model.AppError) {
	search.SkillIds = []int64{skillId}
	list, err := app.Store.AgentSkill().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
