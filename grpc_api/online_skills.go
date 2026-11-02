package grpc_api

import (
	"context"

	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"google.golang.org/protobuf/types/known/emptypb"
)

type onlineSkills struct {
	*API

	engine.UnimplementedOnlineSkillsServiceServer
}

func NewOnlineSkillsApi(api *API) *onlineSkills { return &onlineSkills{API: api} }

func (api *onlineSkills) CreateOnlineSkills(ctx context.Context, in *engine.CreateOnlineSkillsRequest) (*engine.CreateOnlineSkillsResponse, error) {
	onlineSkills := model.OnlineSkills{
		Name:   in.GetName(),
		Skills: make([]*model.Lookup, 0, len(in.GetSkills())),
	}

	onlineSkills.TryUseDescription(in.GetDescription())

	for _, skill := range in.GetSkills() {
		onlineSkills.Skills = append(onlineSkills.Skills, &model.Lookup{Id: int(skill.GetId())})
	}

	response, err := api.ctrl.CreateOnlineSkills(ctx, &onlineSkills)
	if err != nil {
		return nil, err
	}

	return &engine.CreateOnlineSkillsResponse{Item: mapOnlineSkillsToProto(response)}, nil
}

func (api *onlineSkills) DeleteOnlineSkills(ctx context.Context, in *engine.DeleteOnlineSkillsRequest) (*emptypb.Empty, error) {
	deleteCmd := &model.DeleteSkillPresetCmd{
		ID: in.GetId(),
	}

	err := api.ctrl.DeleteOnlineSkills(ctx, deleteCmd)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (api *onlineSkills) GetOnlineSkills(ctx context.Context, in *engine.GetOnlineSkillsRequest) (*engine.GetOnlineSkillsResponse, error) {
	query := &model.GetSkillPresetQuery{
		ID: in.GetId(),
	}

	response, err := api.ctrl.GetOnlineSkills(ctx, query)
	if err != nil {
		return nil, err
	}

	return &engine.GetOnlineSkillsResponse{Item: mapOnlineSkillsToProto(response)}, nil
}

func (api *onlineSkills) PatchOnlineSkills(ctx context.Context, in *engine.PatchOnlineSkillsRequest) (*engine.PatchOnlineSkillsResponse, error) {
	patch := &model.PatchOnlineSkillsCmd{
		Fields: in.GetFields(),
		ID:     in.GetId(),
	}

	for _, field := range in.GetFields() {
		switch field {
		case "name":
			patch.Name = in.Name
		case "description":
			patch.Description = in.Description
		case "skills":
			patch.Skills = make([]*model.Lookup, 0, len(in.GetSkills()))
			for _, skill := range in.GetSkills() {
				patch.Skills = append(patch.Skills, &model.Lookup{Id: int(skill.GetId())})
			}
		}
	}

	result, err := api.ctrl.PatchOnlineSkills(ctx, patch)
	if err != nil {
		return nil, err
	}

	return &engine.PatchOnlineSkillsResponse{Item: mapOnlineSkillsToProto(result)}, nil
}

func (api *onlineSkills) SearchOnlineSkills(ctx context.Context, in *engine.SearchOnlineSkillsRequest) (*engine.SearchOnlineSkillsResponse, error) {
	query := &model.SearchOnlineSkillsQuery{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.GetFields(),
			Sort:    in.GetSort(),
		},
		IDs:         in.GetIds(),
		SkillIDs:    in.GetSkillIds(),
		SkipDefault: in.GetSkipDefault(),
	}

	response, err := api.ctrl.SearchOnlineSkills(ctx, query)
	if err != nil {
		return nil, err
	}
	query.RemoveLastElemIfNeed(&response)

	return &engine.SearchOnlineSkillsResponse{
		Items: mapOnlineSkillsSliceToProto(response),
		Next:  !query.EndOfList(),
	}, nil
}

func (api *onlineSkills) UpdateOnlineSkills(ctx context.Context, in *engine.UpdateOnlineSkillsRequest) (*engine.UpdateOnlineSkillsResponse, error) {
	updateCmd := model.OnlineSkills{
		ID:          in.GetId(),
		Name:        in.GetName(),
		Description: model.NewString(in.GetDescription()),
		Skills:      make([]*model.Lookup, 0, len(in.GetSkills())),
	}

	for _, skill := range in.GetSkills() {
		updateCmd.Skills = append(updateCmd.Skills, &model.Lookup{Id: int(skill.GetId())})
	}

	response, err := api.ctrl.UpdateOnlineSkills(ctx, &updateCmd)
	if err != nil {
		return nil, err
	}

	return &engine.UpdateOnlineSkillsResponse{Item: mapOnlineSkillsToProto(response)}, nil
}

func mapOnlineSkillsSliceToProto(in []*model.OnlineSkills) []*engine.OnlineSkills {
	response := make([]*engine.OnlineSkills, 0, len(in))

	for _, skill := range in {
		response = append(response, mapOnlineSkillsToProto(skill))
	}

	return response
}

func mapOnlineSkillsToProto(in *model.OnlineSkills) *engine.OnlineSkills {
	if in == nil {
		return nil //nolint:nilnil
	}

	return &engine.OnlineSkills{
		Id:          in.ID,
		CreatedBy:   GetProtoLookup(in.CreatedBy),
		CreatedAt:   in.CreatedAtUnix(),
		UpdatedBy:   GetProtoLookup(in.UpdatedBy),
		UpdatedAt:   in.UpdatedAtUnix(),
		Name:        in.Name,
		Description: in.GetDescription(),
		Skills:      GetProtoLookups(in.Skills),
	}
}
