package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type emailProfile struct {
	*API
}

func NewEmailProfileApi(app *API) *emailProfile {
	return &emailProfile{app}
}

func (api *emailProfile) CreateEmailProfile(ctx context.Context, in *engine.CreateEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.EmailProfile{
		DomainRecord: model.DomainRecord{
			DomainId: in.GetDomainId(),
		},
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:  in.GetEnabled(),
		Host:     in.GetHost(),
		Login:    in.GetLogin(),
		Password: in.GetPassword(),
		Mailbox:  in.GetMailbox(),
		SmtpPort: int(in.GetSmtpPort()),
		ImapPort: int(in.GetImapPort()),
	}
	var profile *model.EmailProfile
	profile, err = api.ctrl.CreateEmailProfile(session, req)
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func (api *emailProfile) SearchEmailProfile(ctx context.Context, in *engine.SearchEmailProfileRequest) (*engine.ListEmailProfile, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.EmailProfile
	var endList bool
	req := &model.SearchEmailProfile{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.Sort,
			Fields:   in.Fields,
			Q:        in.GetQ(),
		},
	}

	list, endList, err = api.ctrl.SearchEmailProfile(session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.EmailProfile, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineEmailProfile(v))
	}

	return &engine.ListEmailProfile{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *emailProfile) ReadEmailProfile(ctx context.Context, in *engine.ReadEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var profile *model.EmailProfile
	profile, err = api.ctrl.GetEmailProfile(session, in.GetDomainId(), int(in.GetId()))
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func (api *emailProfile) UpdateEmailProfile(ctx context.Context, in *engine.UpdateEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	profile := &model.EmailProfile{
		DomainRecord: model.DomainRecord{
			Id:       in.Id,
			DomainId: in.GetDomainId(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:  in.GetEnabled(),
		Host:     in.GetHost(),
		Login:    in.GetLogin(),
		Password: in.GetPassword(),
		Mailbox:  in.GetMailbox(),
		SmtpPort: int(in.GetSmtpPort()),
		ImapPort: int(in.GetImapPort()),
	}

	profile, err = api.ctrl.UpdateEmailProfile(session, profile)
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func (api *emailProfile) DeleteEmailProfile(ctx context.Context, in *engine.DeleteEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var profile *model.EmailProfile
	profile, err = api.ctrl.RemoveEmailProfile(session, in.GetDomainId(), int(in.GetId()))
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func toEngineEmailProfile(src *model.EmailProfile) *engine.EmailProfile {
	return &engine.EmailProfile{
		Id:        src.Id,
		DomainId:  src.DomainId,
		CreatedAt: src.CreatedAt,
		CreatedBy: &engine.Lookup{
			Id:   int64(src.CreatedBy.Id),
			Name: src.CreatedBy.Name,
		},
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: &engine.Lookup{
			Id:   int64(src.UpdatedBy.Id),
			Name: src.UpdatedBy.Name,
		},
		Name:        src.Name,
		Description: src.Description,
		Schema:      GetProtoLookup(&src.Schema),
		Enabled:     src.Enabled,
		Host:        src.Host,
		Login:       src.Login,
		Mailbox:     src.Mailbox,
		SmtpPort:    int32(src.SmtpPort),
		ImapPort:    int32(src.ImapPort),
	}
}
