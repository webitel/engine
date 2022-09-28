package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type emailProfile struct {
	*API
	engine.UnsafeEmailProfileServiceServer
}

func NewEmailProfileApi(api *API) *emailProfile {
	return &emailProfile{API: api}
}

func (api *emailProfile) CreateEmailProfile(ctx context.Context, in *engine.CreateEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.EmailProfile{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:       in.GetEnabled(),
		Login:         in.GetLogin(),
		Password:      in.GetPassword(),
		Mailbox:       in.GetMailbox(),
		SmtpHost:      in.GetSmtpHost(),
		SmtpPort:      int(in.GetSmtpPort()),
		ImapHost:      in.GetImapHost(),
		ImapPort:      int(in.GetImapPort()),
		FetchInterval: in.GetFetchInterval(),
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
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Sort:    in.Sort,
			Fields:  in.Fields,
			Q:       in.GetQ(),
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
	profile, err = api.ctrl.GetEmailProfile(session, int(in.GetId()))
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func (api *emailProfile) TestEmailProfile(ctx context.Context, in *engine.TestEmailProfileRequest) (*engine.TestEmailProfileResponse, error) {
	return &engine.TestEmailProfileResponse{
		Error: "TODO",
	}, nil
}

func (api *emailProfile) PatchEmailProfile(ctx context.Context, in *engine.PatchEmailProfileRequest) (*engine.EmailProfile, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var profile *model.EmailProfile
	patch := &model.EmailProfilePatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		case "schema.id":
			patch.Schema = GetLookup(in.Schema)
		case "enabled":
			patch.Enabled = &in.Enabled
		case "imap_host":
			patch.ImapHost = &in.ImapHost
		case "smtp_host":
			patch.SmtpHost = &in.SmtpHost
		case "fetch_interval":
			patch.FetchInterval = &in.FetchInterval
		case "login":
			patch.Login = &in.Login
		case "password":
			patch.Password = &in.Password
		case "mailbox":
			patch.Mailbox = &in.Mailbox
		case "smtp_port":
			patch.SmtpPort = model.NewInt(int(in.SmtpPort))
		case "imap_port":
			patch.ImapPort = model.NewInt(int(in.ImapPort))
		}
	}

	profile, err = api.ctrl.PatchEmailProfile(session, int(in.GetId()), patch)

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
			Id: in.Id,
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:       in.GetEnabled(),
		Login:         in.GetLogin(),
		Password:      in.GetPassword(),
		Mailbox:       in.GetMailbox(),
		SmtpHost:      in.GetSmtpHost(),
		SmtpPort:      int(in.GetSmtpPort()),
		ImapHost:      in.GetImapHost(),
		ImapPort:      int(in.GetImapPort()),
		FetchInterval: in.GetFetchInterval(),
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
	profile, err = api.ctrl.RemoveEmailProfile(session, int(in.GetId()))
	if err != nil {
		return nil, err
	}

	return toEngineEmailProfile(profile), nil
}

func toEngineEmailProfile(src *model.EmailProfile) *engine.EmailProfile {
	profile := &engine.EmailProfile{
		Id:            src.Id,
		CreatedAt:     src.CreatedAt,
		CreatedBy:     GetProtoLookup(src.CreatedBy),
		UpdatedAt:     src.UpdatedAt,
		UpdatedBy:     GetProtoLookup(src.UpdatedBy),
		Name:          src.Name,
		Description:   src.Description,
		Schema:        GetProtoLookup(&src.Schema),
		Enabled:       src.Enabled,
		ImapHost:      src.ImapHost,
		Login:         src.Login,
		Mailbox:       src.Mailbox,
		SmtpPort:      int32(src.SmtpPort),
		ImapPort:      int32(src.ImapPort),
		Password:      src.Password,
		SmtpHost:      src.SmtpHost,
		FetchInterval: src.FetchInterval,
		State:         src.State,
		ActivityAt:    src.ActivityAt,
	}

	if src.FetchError != nil {
		profile.FetchError = *src.FetchError
	}

	return profile
}
