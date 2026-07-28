package sqlstore

import (
	"context"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
)

type SqlSkillPresetStore struct {
	SqlStore
}

func NewSqlSkillPresetStore(sqlStore SqlStore) *SqlSkillPresetStore {
	return &SqlSkillPresetStore{SqlStore: sqlStore}
}

func (s *SqlSkillPresetStore) Create(ctx context.Context, preset *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	query := `
		with preset_ins as (
			insert into "call_center"."skill_preset" (
				"domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "description"
			)
			values (
				:DomainID, :CreatedBy, now(), :UpdatedBy, now(), :Name, :Description
			)
			returning "id","domain_id", "name", "created_by", "created_at", "updated_by", "updated_at", "description"
		),
		skills_in_preset_ins as (
			insert into "call_center"."skills_in_skill_preset" (
				"domain_id", "skill_preset_id", "skill_id"
			)
			select
				:DomainID,
				p.id,
				s.id
			from unnest(:Skills::int8[]) s(id)
			cross join preset_ins p
			returning "skill_preset_id", "skill_id"
		)
		select
			p.id as id,
			p.domain_id as domain_id,
			call_center.cc_get_lookup(uc.id, uc.name) as created_by,
			p.created_at as created_at,
			call_center.cc_get_lookup(ua.id, ua.name) as updated_by,
			p.updated_at as updated_at,
			p.name as "name",
			p.description as "description",
			s.skills as "skills"
		from preset_ins p
		left join lateral (
			select jsonb_agg(
				call_center.cc_get_lookup(s.id, s.name)
			) as skills
			from skills_in_preset_ins si
			inner join call_center.cc_skill s on s.id = si.skill_id
		) s on true
		left join directory.wbt_user uc on uc.id = p.created_by
		left join directory.wbt_user ua on ua.id = p.updated_by
	`

	args := map[string]any{
		"DomainID":    preset.DomainID,
		"CreatedBy":   preset.CreatedBy.GetSafeId(),
		"UpdatedBy":   preset.UpdatedBy.GetSafeId(),
		"Name":        preset.Name,
		"Description": preset.Description,
		"Skills":      pq.Int64Array(preset.ReduceSkillsIDs()),
	}

	var result *model.SkillPreset
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == DuplicationViolationErrorCode {
				return nil, model.NewBadRequestError("sqlstore.skill_preset_store.create_already_exists", " Skill preset with this name already exists.")
			}
		}

		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.create", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlSkillPresetStore) Update(ctx context.Context, preset *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	query := `
		with preset_upd as (
			update "call_center"."skill_preset"
			set "updated_by"  = :UpdatedBy,
				"updated_at"  = now(),
				"name"        = :Name,
				"description" = :Description
			where "id" = :ID and "domain_id" = :DomainID
			returning "id", "domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "description"
		),
		binded_skills_del as (
			delete from "call_center"."skills_in_skill_preset"
			where "skill_preset_id" = :ID
			  and "skill_id" <> all(:Skills::int8[])
		),
		new_skills_ins as (
			insert into "call_center"."skills_in_skill_preset" (
				"domain_id", "skill_preset_id", "skill_id"
			)
			select
				p.domain_id,
				p.id,
				s.id
			from unnest(:Skills::int8[]) as s(id)
			cross join preset_upd p
			on conflict ("skill_preset_id", "skill_id") do nothing
			returning "skill_id"
		),
		actual_skills as (
			select skill_id
			from new_skills_ins
			union
			select skill_id
			from "call_center"."skills_in_skill_preset"
			where skill_preset_id = :ID
			  and skill_id = any(:Skills::int8[])
		)
		select
			p.id as id,
			p.domain_id as domain_id,
			call_center.cc_get_lookup(uc.id, uc.name) as created_by,
			p.created_at as created_at,
			call_center.cc_get_lookup(ua.id, ua.name) as updated_by,
			p.updated_at as updated_at,
			p.name as "name",
			p.description as "description",
			coalesce(s.skills, '[]'::jsonb) as "skills"
		from preset_upd p
		left join lateral (
			select jsonb_agg(
				call_center.cc_get_lookup(s.id, s.name)
			) as skills
			from actual_skills ask
			inner join "call_center"."cc_skill" s on s.id = ask.skill_id
		) s on true
		left join directory.wbt_user uc on uc.id = p.created_by
		left join directory.wbt_user ua on ua.id = p.updated_by
	`
	args := map[string]any{
		"ID":          preset.ID,
		"DomainID":    preset.DomainID,
		"UpdatedBy":   preset.UpdatedBy.GetSafeId(),
		"Name":        preset.Name,
		"Description": preset.Description,
		"Skills":      pq.Int64Array(preset.ReduceSkillsIDs()),
	}

	var result *model.SkillPreset
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.update", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlSkillPresetStore) Patch(ctx context.Context, patchCmd *model.PatchSkillPresetCmd) (*model.SkillPreset, model.AppError) {
	query := `
		select * from call_center.cc_patch_skill_preset(
			:ID,
			:DomainID,
			:UpdatedBy,
			:Name,
			:Description,
			:PatchDescription,
			:Skills::int8[],
			:PatchSkills
		)
	`

	args := map[string]any{
		"ID":               patchCmd.ID,
		"DomainID":         patchCmd.DomainID,
		"UpdatedBy":        patchCmd.UpdatedBy.GetSafeId(),
		"Name":             nil,
		"Description":      nil,
		"PatchDescription": false,
		"Skills":           pq.Int64Array([]int64{}),
		"PatchSkills":      false,
	}

	for _, field := range patchCmd.Fields {
		switch field {
		case "name":
			args["Name"] = patchCmd.Name
		case "description":
			args["Description"] = patchCmd.Description
			args["PatchDescription"] = true
		case "skills":
			args["Skills"] = pq.Int64Array(patchCmd.ReduceSkillsIDs())
			args["PatchSkills"] = true
		}
	}

	var result model.SkillPreset
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.patch", err.Error(), extractCodeFromErr(err))
	}

	return &result, nil
}

func (s *SqlSkillPresetStore) Delete(ctx context.Context, deleteCmd *model.DeleteSkillPresetCmd) ([]*model.SkillPreset, model.AppError) {
	query := `
		with preset_del as (
			delete from "call_center"."skill_preset"
			where "domain_id" = :DomainID and "id" = any(:IDs)
			returning "id", "domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "description"
		),
		skills_del as (
			delete from "call_center"."skills_in_skill_preset"
			where "skill_preset_id" in (select id from preset_del)
			returning "skill_preset_id", "skill_id"
		)
		select
			p.id as id,
			p.domain_id as domain_id,
			call_center.cc_get_lookup(uc.id, uc.name) as created_by,
			p.created_at as created_at,
			call_center.cc_get_lookup(ua.id, ua.name) as updated_by,
			p.updated_at as updated_at,
			p.name as "name",
			p.description as "description",
			s.skills as "skills"
		from preset_del p
		left join lateral (
			select jsonb_agg(
				call_center.cc_get_lookup(s.id, s.name)
			) as skills
			from skills_del sd
			inner join "call_center"."cc_skill" s on s.id = sd.skill_id
			where sd.skill_preset_id = p.id
		) s on true
		left join directory.wbt_user uc on uc.id = p.created_by
		left join directory.wbt_user ua on ua.id = p.updated_by
	`

	args := map[string]any{
		"DomainID": deleteCmd.DomainID,
		"IDs":      pq.Int64Array(deleteCmd.IDs),
	}

	var result []*model.SkillPreset
	if _, err := s.GetMaster().WithContext(ctx).Select(&result, query, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.delete", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlSkillPresetStore) Search(ctx context.Context, search *model.SearchSkillPresetQuery) ([]*model.SkillPreset, model.AppError) {
	query := `
		"domain_id" = :DomainID
		and (:IDs::int[] is null or "id" = any(:IDs::int[]))
		and (
			cardinality(:SkillIDs::int8[]) = 0 or :SkillIDs::int8[] is null
			or exists (
					select 1
					from "call_center"."skills_in_skill_preset" si
					where si.skill_preset_id = id
						and si.skill_id = any(:SkillIDs::int8[])
			)
		)
		and (:Q::text is null or "name" ilike :Q::text)
		and (:SkipDefault is false or "is_system" is false)
	`
	args := map[string]any{
		"DomainID":    search.DomainId,
		"IDs":         pq.Int64Array(search.IDs),
		"SkillIDs":    pq.Int64Array(search.SkillIDs),
		"Q":           search.GetQ(),
		"SkipDefault": search.SkipDefault,
	}

	search.Sort = search.OrderBy()

	var result []*model.SkillPreset
	if err := s.ListQuery(ctx, &result, search.ListRequest, query, model.SkillPreset{}, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.search", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlSkillPresetStore) Get(ctx context.Context, search *model.GetSkillPresetQuery) (*model.SkillPreset, model.AppError) {
	query := `
			select
				id,
				domain_id,
				created_by,
				created_at,
				updated_by,
				updated_at,
				name,
				description,
				skills
			from "call_center"."cc_skill_preset_view"
			where "domain_id" = :DomainID and "id" = :ID
		`

	args := map[string]any{
		"DomainID": search.DomainID,
		"ID":       search.ID,
	}

	var result *model.SkillPreset
	if err := s.GetReplica().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.skill_preset_store.get", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}
