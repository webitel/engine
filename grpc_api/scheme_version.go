package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type schemeVersion struct {
	*API
	engine.UnsafeSchemeVersionServiceServer
}

func NewSchemeVersionApi(api *API) *schemeVersion {
	return &schemeVersion{API: api}
}

func (api schemeVersion) Search(ctx context.Context, in *engine.SearchSchemeVersionRequest) (*engine.SearchSchemeVersionResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.SchemeVersion
	var endList bool

	req := &model.SearchSchemeVersion{
		ListRequest: model.ExtractSearchOptions(in),
		SchemeId:    in.GetSchemeId(),
	}

	list, endList, err = api.ctrl.SearchSchemeVersions(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.SchemeVersion, 0, len(list))
	for _, v := range list {
		items = append(items, transformSchemeVersion(v))
	}
	return &engine.SearchSchemeVersionResponse{
		Next:  !endList,
		Items: items,
	}, nil
}

func transformSchemeVersion(s *model.SchemeVersion) *engine.SchemeVersion {
	res := &engine.SchemeVersion{
		Id:        s.Id,
		SchemeId:  s.SchemeId,
		CreatedAt: s.CreatedAt,
		CreatedBy: &engine.Lookup{Id: int64(s.CreatedBy.Id), Name: s.CreatedBy.Name},
		Scheme:    UnmarshalJsonpb(s.Scheme),
		Payload:   UnmarshalJsonpb(s.Payload),
		Version:   uint64(s.Version),
	}
	if s.Note != nil {
		res.Note = *s.Note
	}

	return res
}
