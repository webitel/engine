package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type skill struct {
	*API
	engine.UnsafeSkillServiceServer
}

func NewSkillApi(api *API) *skill {
	return &skill{API: api}
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

	skill, err = api.ctrl.CreateSkill(ctx, session, skill)
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
	var endList bool
	req := &model.SearchSkill{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.ctrl.SearchSkill(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Skill, 0, len(list))
	for _, v := range list {
		items = append(items, transformSkill(v))
	}
	return &engine.ListSkill{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *skill) ReadSkill(ctx context.Context, in *engine.ReadSkillRequest) (*engine.Skill, error) {
	var skill *model.Skill
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	skill, err = api.ctrl.ReadSkill(ctx, session, in.Id)
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

	skill, err = api.ctrl.UpdateSkill(ctx, session, &model.Skill{
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
	skill, err = api.ctrl.DeleteSkill(ctx, session, in.Id)
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
