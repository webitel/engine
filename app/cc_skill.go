package app

import "github.com/webitel/engine/model"

func (app *App) CreateSkill(skill *model.Skill) (*model.Skill, *model.AppError) {
	return app.Store.Skill().Create(skill)
}

func (app *App) GetSkill(id, domainId int64) (*model.Skill, *model.AppError) {
	return app.Store.Skill().Get(domainId, id)
}

func (app *App) GetSkillsPage(domainId int64, page, perPage int) ([]*model.Skill, *model.AppError) {
	return app.Store.Skill().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) RemoveSkill(domainId, id int64) (*model.Skill, *model.AppError) {
	skill, err := app.Store.Skill().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.Skill().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return skill, nil
}

func (app *App) UpdateSkill(skill *model.Skill) (*model.Skill, *model.AppError) {
	oldSkill, err := app.Store.Skill().Get(skill.DomainId, skill.Id)

	if err != nil {
		return nil, err
	}

	oldSkill.Name = skill.Name
	oldSkill.Description = skill.Description

	_, err = app.Store.Skill().Update(oldSkill)
	if err != nil {
		return nil, err
	}

	return oldSkill, nil
}
