package grpc_api

import (
	"context"
	"strings"

	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type skillPreset struct {
	*API

	engine.UnimplementedSkillPresetServiceServer
}

func NewSkillPresetApi(api *API) *skillPreset { return &skillPreset{API: api} }

func (api *skillPreset) CreateSkillPreset(ctx context.Context, in *engine.CreateSkillPresetRequest) (*engine.CreateSkillPresetResponse, error) {
	skillPreset := model.SkillPreset{
		Name:   in.GetName(),
		Skills: make([]*model.Lookup, 0, len(in.GetSkills())),
	}

	if strings.TrimSpace(in.GetDescription()) != "" {
		skillPreset.Description = model.NewString(in.GetDescription())
	}

	for _, skill := range in.GetSkills() {
		skillPreset.Skills = append(skillPreset.Skills, &model.Lookup{Id: int(skill.GetId())})
	}

	response, err := api.ctrl.CreateSkillPreset(ctx, &skillPreset)
	if err != nil {
		return nil, err
	}

	return &engine.CreateSkillPresetResponse{Item: mapSkillPresetToProto(response)}, nil
}

func (api *skillPreset) DeleteSkillPreset(ctx context.Context, in *engine.DeleteSkillPresetRequest) (*engine.DeleteSkillPresetResponse, error) {
	deleteCmd := &model.DeleteSkillPresetCmd{
		IDs: in.GetIds(),
	}

	result, err := api.ctrl.DeleteSkillPreset(ctx, deleteCmd)
	if err != nil {
		return nil, err
	}

	return &engine.DeleteSkillPresetResponse{Items: mapSkillPresetsToProto(result)}, nil
}

func (api *skillPreset) GetSkillPreset(ctx context.Context, in *engine.GetSkillPresetRequest) (*engine.GetSkillPresetResponse, error) {
	query := &model.GetSkillPresetQuery{
		ID: in.GetId(),
	}

	response, err := api.ctrl.GetSkillPreset(ctx, query)
	if err != nil {
		return nil, err
	}

	return &engine.GetSkillPresetResponse{Item: mapSkillPresetToProto(response)}, nil
}

func (api *skillPreset) PatchSkillPreset(ctx context.Context, in *engine.PatchSkillPresetRequest) (*engine.PatchSkillPresetResponse, error) {
	patch := &model.PatchSkillPresetCmd{
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

	result, err := api.ctrl.PatchSkillPreset(ctx, patch)
	if err != nil {
		return nil, err
	}

	return &engine.PatchSkillPresetResponse{Item: mapSkillPresetToProto(result)}, nil
}

func (api *skillPreset) SearchSkillPreset(ctx context.Context, in *engine.SearchSkillPresetRequest) (*engine.SearchSkillPresetResponse, error) {
	query := &model.SearchSkillPresetQuery{
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

	response, err := api.ctrl.SearchSkillPreset(ctx, query)
	if err != nil {
		return nil, err
	}
	query.RemoveLastElemIfNeed(&response)

	return &engine.SearchSkillPresetResponse{
		Items: mapSkillPresetsToProto(response),
		Next:  !query.EndOfList(),
	}, nil
}

func (api *skillPreset) UpdateSkillPreset(ctx context.Context, in *engine.UpdateSkillPresetRequest) (*engine.UpdateSkillPresetResponse, error) {
	updateCmd := model.SkillPreset{
		ID:          in.GetId(),
		Name:        in.GetName(),
		Description: model.NewString(in.GetDescription()),
		Skills:      make([]*model.Lookup, 0, len(in.GetSkills())),
	}

	for _, skill := range in.GetSkills() {
		updateCmd.Skills = append(updateCmd.Skills, &model.Lookup{Id: int(skill.GetId())})
	}

	response, err := api.ctrl.UpdateSkillPreset(ctx, &updateCmd)
	if err != nil {
		return nil, err
	}

	return &engine.UpdateSkillPresetResponse{Item: mapSkillPresetToProto(response)}, nil
}

func mapSkillPresetsToProto(in []*model.SkillPreset) []*engine.SkillPreset {
	response := make([]*engine.SkillPreset, 0, len(in))

	for _, skill := range in {
		response = append(response, mapSkillPresetToProto(skill))
	}

	return response
}

func mapSkillPresetToProto(in *model.SkillPreset) *engine.SkillPreset {
	if in == nil {
		return nil //nolint:nilnil
	}

	return &engine.SkillPreset{
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
