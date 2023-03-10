package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type auditForm struct {
	*API
	engine.UnsafeAuditFormServiceServer
}

func (api *auditForm) CreateAuditForm(ctx context.Context, in *engine.CreateAuditFormRequest) (*engine.AuditForm, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	form := &model.AuditForm{
		Id:          0,
		AclRecord:   model.AclRecord{},
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Enabled:     in.GetEnabled(),
		Questions:   toAuditQuestions(in.Questions),
		Teams:       GetLookups(in.GetTeams()),
	}

	form, err = api.ctrl.CreateAuditForm(ctx, session, form)
	if err != nil {
		return nil, err
	}

	return transformAuditFrom(form), nil
}

func (api *auditForm) SearchAuditForm(ctx context.Context, in *engine.SearchAuditFormRequest) (*engine.ListAuditForm, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AuditForm
	var endList bool
	req := &model.SearchAuditForm{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.ctrl.SearchAuditForm(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AuditForm, 0, len(list))
	for _, v := range list {
		items = append(items, transformAuditFrom(v))
	}
	return &engine.ListAuditForm{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *auditForm) ReadAuditForm(ctx context.Context, in *engine.ReadAuditFormRequest) (*engine.AuditForm, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var form *model.AuditForm
	form, err = api.ctrl.ReadAuditForm(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return transformAuditFrom(form), nil
}

func (api *auditForm) UpdateAuditForm(ctx context.Context, in *engine.UpdateAuditFormRequest) (*engine.AuditForm, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	form := &model.AuditForm{
		Id:          in.Id,
		AclRecord:   model.AclRecord{},
		Name:        in.Name,
		Description: in.Description,
		Enabled:     in.Enabled,
		Questions:   toAuditQuestions(in.Questions),
		Teams:       GetLookups(in.GetTeams()),
	}

	form, err = api.ctrl.PutAuditForm(ctx, session, form)

	if err != nil {
		return nil, err
	}

	return transformAuditFrom(form), nil
}

func (api *auditForm) PatchAuditForm(ctx context.Context, in *engine.PatchAuditFormRequest) (*engine.AuditForm, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var form *model.AuditForm
	patch := &model.AuditFormPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		case "enabled":
			patch.Enabled = &in.Enabled
		case "teams":
			patch.Teams = GetLookups(in.Teams)
		case "questions":
			patch.Questions = toAuditQuestions(in.Questions)
		}
	}

	form, err = api.ctrl.PatchAuditForm(ctx, session, in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return transformAuditFrom(form), nil
}

func (api *auditForm) DeleteAuditForm(ctx context.Context, in *engine.DeleteAuditFormRequest) (*engine.AuditForm, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var form *model.AuditForm
	form, err = api.ctrl.DeleteAuditForm(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformAuditFrom(form), nil
}

//func (a auditForm) mustEmbedUnimplementedAuditFormServiceServer() {
//	//TODO implement me
//	panic("implement me")
//}

func NewAuditFormApi(api *API) *auditForm {
	return &auditForm{API: api}
}

func transformAuditFrom(src *model.AuditForm) *engine.AuditForm {
	return &engine.AuditForm{
		Id:          src.Id,
		CreatedAt:   model.TimeToInt64(src.CreatedAt),
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedAt:   model.TimeToInt64(src.UpdatedAt),
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		Name:        src.Name,
		Description: src.Description,
		Enabled:     src.Enabled,
		Questions:   transformAuditQuestions(src.Questions),
		Teams:       GetProtoLookups(src.Teams),
	}
}

func toAuditQuestions(src []*engine.Questions) model.Questions {
	q := make(model.Questions, 0, len(src))
	for _, v := range src {
		switch i := v.To.(type) {
		case *engine.Questions_Options:
			ops := make([]model.QuestionOption, 0, len(i.Options.Options))
			for _, o := range i.Options.Options {
				ops = append(ops, model.QuestionOption{
					Name:  o.GetName(),
					Score: o.GetScore(),
				})
			}
			q = append(q, model.Question{
				Type:     model.QuestionTypeOptions,
				Required: i.Options.GetRequired(),
				Question: i.Options.GetQuestion(),
				Options:  ops,
			})

		case *engine.Questions_Score:
			q = append(q, model.Question{
				Type:     model.QuestionTypeScore,
				Required: i.Score.GetRequired(),
				Question: i.Score.GetQuestion(),
				Min:      i.Score.GetMin(),
				Max:      i.Score.GetMax(),
			})
		}
	}

	return q
}

func transformAuditQuestions(src model.Questions) []*engine.Questions {
	q := make([]*engine.Questions, 0, len(src))
	for _, v := range src {
		switch v.Type {
		case model.QuestionTypeOptions:
			ops := make([]*engine.QuestionOptions_Option, 0, len(v.Options))
			for _, j := range v.Options {
				ops = append(ops, &engine.QuestionOptions_Option{
					Name:  j.Name,
					Score: j.Score,
				})
			}
			q = append(q, &engine.Questions{
				To: &engine.Questions_Options{
					Options: &engine.QuestionOptions{
						Required: v.Required,
						Question: v.Question,
						Options:  ops,
					},
				},
			})
		case model.QuestionTypeScore:
			q = append(q, &engine.Questions{
				To: &engine.Questions_Score{
					Score: &engine.QuestionScore{
						Required: v.Required,
						Question: v.Question,
						Min:      v.Min,
						Max:      v.Max,
					},
				},
			})
		}
	}

	return q
}
