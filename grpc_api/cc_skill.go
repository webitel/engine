package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type skill struct {
	app *app.App
}

func NewSkillApi(app *app.App) *skill {
	return &skill{app: app}
}

func (api *skill) CreateSkill(ctx context.Context, in *engine.CreateSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skill := &model.Skill{
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	}

	err = skill.IsValid()
	if err != nil {
		return nil, err
	}

	skill, err = api.app.CreateSkill(skill)
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) SearchSkill(ctx context.Context, in *engine.SearchSkillRequest) (*engine.ListSkill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Skill
	list, err = api.app.GetSkillsPage(session.Domain(int64(in.DomainId)), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Skill, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkill(v))
	}
	return &engine.ListSkill{
		Items: items,
	}, nil
}

func (api *skill) ReadSkill(ctx context.Context, in *engine.ReadSkillRequest) (*engine.Skill, error) {
	var skill *model.Skill
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skill, err = api.app.GetSkill(in.Id, session.Domain(in.GetDomainId()))
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) UpdateSkill(ctx context.Context, in *engine.UpdateSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var skill *model.Skill

	skill, err = api.app.UpdateSkill(&model.Skill{
		Id:          in.Id,
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func (api *skill) DeleteSkill(ctx context.Context, in *engine.DeleteSkillRequest) (*engine.Skill, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var skill *model.Skill
	skill, err = api.app.RemoveSkill(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformSkill(skill), nil
}

func transformSkill(src *model.Skill) *engine.Skill {
	return &engine.Skill{
		Id:          src.Id,
		DomainId:    src.DomainId,
		Name:        src.Name,
		Description: src.Description,
	}
}
