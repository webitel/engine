package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type schemaVersion struct {
	*API
	gogrpc.UnsafeSchemaVersionServiceServer
}

func NewSchemeVersionApi(api *API) *schemaVersion {
	return &schemaVersion{API: api}
}

func (api schemaVersion) SearchSchemaVersion(ctx context.Context, in *engine.SearchSchemaVersionRequest) (*engine.SearchSchemaVersionResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.SchemeVersion
	var endList bool

	req := &model.SearchSchemeVersion{
		ListRequest: model.ExtractSearchOptions(in),
		SchemeId:    in.GetSchemaId(),
	}

	list, endList, err = api.ctrl.SearchSchemeVersions(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.SchemaVersion, 0, len(list))
	for _, v := range list {
		items = append(items, transformSchemeVersion(v))
	}
	return &engine.SearchSchemaVersionResponse{
		Next:  !endList,
		Items: items,
	}, nil
}

func transformSchemeVersion(s *model.SchemeVersion) *engine.SchemaVersion {
	res := &engine.SchemaVersion{
		Id:        s.Id,
		SchemaId:  s.SchemeId,
		CreatedAt: s.CreatedAt,
		CreatedBy: &engine.Lookup{Id: int64(s.CreatedBy.Id), Name: s.CreatedBy.Name},
		Schema:    UnmarshalJsonpb(s.Scheme),
		Payload:   UnmarshalJsonpb(s.Payload),
		Version:   uint64(s.Version),
	}
	if s.Note != nil {
		res.Note = *s.Note
	}

	return res
}
