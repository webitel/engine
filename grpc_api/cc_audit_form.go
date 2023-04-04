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
		Ids:     in.Id,
		TeamIds: in.TeamId,
	}

	if in.Enabled {
		req.Enabled = &in.Enabled
	}

	if in.Archive {
		req.Archive = &in.Archive
	}

	if in.Editable {
		req.Editable = &in.Editable
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

func (api *auditForm) CreateAuditFormRate(ctx context.Context, in *engine.CreateAuditFormRateRequest) (*engine.AuditRate, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var auditRate *model.AuditRate

	ans := make([]*model.QuestionAnswer, 0, len(in.Answers))
	for _, v := range in.Answers {
		if v == nil {
			ans = append(ans, nil)
		} else {
			ans = append(ans, &model.QuestionAnswer{
				Score: v.GetScore(),
			})
		}
	}

	rate := model.Rate{
		CallId: nil,
		Form: &model.Lookup{
			Id: int(in.GetForm().GetId()),
		},
		Answers: ans,
		Comment: in.Comment,
	}
	if in.CallId != "" {
		rate.CallId = &in.CallId
	}

	auditRate, err = api.ctrl.RateAuditForm(ctx, session, rate)
	if err != nil {
		return nil, err
	}

	return transformAuditRate(auditRate), nil
}

func (api *auditForm) SearchAuditRate(ctx context.Context, in *engine.SearchAuditRateRequest) (*engine.ListAuditRate, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AuditRate
	var endList bool
	req := &model.SearchAuditRate{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:          in.Id,
		CallIds:      in.CallId,
		CreatedAt:    nil,
		FormIds:      nil,
		RatedUserIds: in.RatedUser,
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}

	list, endList, err = api.ctrl.SearchAuditRate(ctx, session, in.GetFormId(), req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AuditRate, 0, len(list))
	for _, v := range list {
		items = append(items, transformAuditRate(v))
	}
	return &engine.ListAuditRate{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *auditForm) ReadAuditRate(ctx context.Context, in *engine.ReadAuditRateRequest) (*engine.AuditRate, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var rate *model.AuditRate
	rate, err = api.ctrl.ReadAuditRate(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return transformAuditRate(rate), nil
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

func transformAuditRate(src *model.AuditRate) *engine.AuditRate {
	return &engine.AuditRate{
		Id:            src.Id,
		CreatedAt:     model.TimeToInt64(src.CreatedAt),
		CreatedBy:     GetProtoLookup(src.CreatedBy),
		UpdatedAt:     model.TimeToInt64(src.UpdatedAt),
		UpdatedBy:     GetProtoLookup(src.UpdatedBy),
		Form:          GetProtoLookup(src.Form),
		Questions:     transformAuditQuestions(src.Questions),
		Answers:       transformAuditAnswers(src.Answers),
		ScoreRequired: src.ScoreRequired,
		ScoreOptional: src.ScoreOptional,
		Comment:       src.Comment,
		RatedUser:     GetProtoLookup(src.RatedUser),
	}
}

func toAuditQuestions(src []*engine.Question) model.Questions {
	q := make(model.Questions, 0, len(src))
	for _, v := range src {
		item := model.Question{
			Required: v.Required,
			Question: v.Question,
		}

		switch v.Type {
		case engine.AuditQuestionType_question_score:
			item.Type = model.QuestionTypeScore
			item.Max = v.Max
			item.Min = v.Min
		case engine.AuditQuestionType_question_option:
			item.Type = model.QuestionTypeOptions
			item.Options = make([]model.QuestionOption, 0, len(v.Options))
			for _, o := range v.Options {
				item.Options = append(item.Options, model.QuestionOption{
					Name:  o.GetName(),
					Score: o.GetScore(),
				})
			}

		}

		q = append(q, item)
	}

	return q
}

func transformAuditQuestions(src model.Questions) []*engine.Question {
	q := make([]*engine.Question, 0, len(src))
	for _, v := range src {
		item := &engine.Question{
			Required: v.Required,
			Question: v.Question,
		}
		switch v.Type {
		case model.QuestionTypeScore:
			item.Type = engine.AuditQuestionType_question_score
			item.Max = v.Max
			item.Min = v.Min
		case model.QuestionTypeOptions:
			item.Type = engine.AuditQuestionType_question_option
			item.Options = make([]*engine.Question_Option, 0, len(v.Options))
			for _, j := range v.Options {
				item.Options = append(item.Options, &engine.Question_Option{
					Name:  j.Name,
					Score: j.Score,
				})
			}
		}

		q = append(q, item)

	}

	return q
}

func transformAuditAnswers(src model.QuestionAnswers) []*engine.QuestionAnswer {
	q := make([]*engine.QuestionAnswer, 0, len(src))
	for _, v := range src {
		if v == nil {
			q = append(q, nil)
		} else {
			q = append(q, &engine.QuestionAnswer{
				Score: v.Score,
			})
		}
	}

	return q
}
