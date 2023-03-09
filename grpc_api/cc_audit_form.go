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
		Questions:   "{}", //todo
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
		Questions:   "{}",
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
			//case "questions":
			//	patch.Questions = &in.Questions
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
		//Questions:   "",
		Teams: GetProtoLookups(src.Teams),
	}
}
