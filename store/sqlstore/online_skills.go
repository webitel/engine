package sqlstore

import (
	"context"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
)

type SqlOnlineSkillsStore struct{ SqlStore }

func NewSqlOnlineSkillsStore(sqlStore SqlStore) *SqlOnlineSkillsStore {
	return &SqlOnlineSkillsStore{SqlStore: sqlStore}
}

func (s *SqlOnlineSkillsStore) Create(ctx context.Context, preset *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	query := `
		with preset_ins as (
			insert into "call_center"."cc_online_skills" (
				"domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "description"
			)
			values (
				:DomainID, :CreatedBy, now(), :UpdatedBy, now(), :Name, :Description
			)
			returning "id","domain_id", "name", "created_by", "created_at", "updated_by", "updated_at", "description"
		),
		skills_in_preset_ins as (
			insert into "call_center"."cc_skills_in_online_skills" (
				"domain_id", "online_skill_id", "skill_id"
			)
			select
				:DomainID,
				p.id,
				s.id
			from unnest(:Skills::int8[]) s(id)
			cross join preset_ins p
			returning "online_skill_id", "skill_id"
		)
		select
			p.id as id,
			p.domain_id as domain_id,
			call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
			p.created_at as created_at,
			call_center.cc_get_lookup(ua.id, coalesce(ua.name, ua.username)) as updated_by,
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

	var result *model.OnlineSkills
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == DuplicationViolationErrorCode {
				return nil, model.NewBadRequestError("sqlstore.online_skills_store.create.already_exists", "Online skills with this name already exists.")
			}
		}

		return nil, model.NewCustomCodeError("sqlstore.online_skills_store.create", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlOnlineSkillsStore) Update(ctx context.Context, preset *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	query := `
		with preset_upd as (
			update "call_center"."cc_online_skills"
			set "updated_by"  = :UpdatedBy,
				"updated_at"  = now(),
				"name"        = :Name,
				"description" = :Description
			where "id" = :ID and "domain_id" = :DomainID
			returning "id", "domain_id", "created_by", "created_at", "updated_by", "updated_at", "name", "description"
		),
		binded_skills_del as (
			delete from "call_center"."cc_skills_in_online_skills"
			where "online_skill_id" = :ID
			  and "skill_id" <> all(:Skills::int8[])
		),
		new_skills_ins as (
			insert into "call_center"."cc_skills_in_online_skills" (
				"domain_id", "online_skill_id", "skill_id"
			)
			select
				p.domain_id,
				p.id,
				s.id
			from unnest(:Skills::int8[]) as s(id)
			cross join preset_upd p
			on conflict ("online_skill_id", "skill_id") do nothing
			returning "skill_id"
		),
		actual_skills as (
			select skill_id
			from new_skills_ins
			union
			select skill_id
			from "call_center"."cc_skills_in_online_skills"
			where online_skill_id = :ID
			  and skill_id = any(:Skills::int8[])
		)
		select
			p.id as id,
			p.domain_id as domain_id,
			call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
			p.created_at as created_at,
			call_center.cc_get_lookup(ua.id, coalesce(ua.name, ua.username)) as updated_by,
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

	var result *model.OnlineSkills
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == DuplicationViolationErrorCode {
				return nil, model.NewBadRequestError("sqlstore.online_skills_store.update.already_exists", "Online skills with this name already exists.")
			}
		}

		return nil, model.NewCustomCodeError("sqlstore.online_skills_store.update", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlOnlineSkillsStore) Patch(ctx context.Context, patchCmd *model.PatchOnlineSkillsCmd) (*model.OnlineSkills, model.AppError) {
	query := `
		select * from call_center.cc_patch_online_skills(
			:ID,
			:DomainID,
			:UpdatedBy,
			:Name,
			:Description,
			:PatchDescription,
			:Skills::int8[],
			:PatchSkills
		);
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

	var result model.OnlineSkills
	if err := s.GetMaster().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		if e, ok := err.(*pq.Error); ok {
			if e.Code == DuplicationViolationErrorCode {
				return nil, model.NewBadRequestError("sqlstore.online_skills_store.patch.already_exists", "Online skills with this name already exists.")
			}
		}

		return nil, model.NewCustomCodeError("sqlstore.online_skills_store.patch", err.Error(), extractCodeFromErr(err))
	}

	return &result, nil
}

func (s *SqlOnlineSkillsStore) Delete(ctx context.Context, deleteCmd *model.DeleteSkillPresetCmd) model.AppError {
	query := `delete from "call_center"."cc_online_skills" where "id" = :ID and "domain_id" = :DomainID;`

	args := map[string]any{
		"DomainID": deleteCmd.DomainID,
		"ID":       deleteCmd.ID,
	}

	if _, err := s.GetMaster().WithContext(ctx).Exec(query, args); err != nil {
		return model.NewCustomCodeError("sqlstore.online_skills_store.delete", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s *SqlOnlineSkillsStore) Search(ctx context.Context, search *model.SearchOnlineSkillsQuery) ([]*model.OnlineSkills, model.AppError) {
	query := `
		"domain_id" = :DomainID
		and (:IDs::int[] is null or "id" = any(:IDs::int[]))
		and (
			cardinality(:SkillIDs::int8[]) = 0 or :SkillIDs::int8[] is null
			or exists (
					select 1
					from "call_center"."cc_skills_in_online_skills" si
					where si.online_skill_id = id
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

	var result []*model.OnlineSkills
	if err := s.ListQuery(ctx, &result, search.ListRequest, query, model.OnlineSkills{}, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.online_skills_store.search", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}

func (s *SqlOnlineSkillsStore) Get(ctx context.Context, search *model.GetSkillPresetQuery) (*model.OnlineSkills, model.AppError) {
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
			from "call_center"."cc_online_skills_list"
			where "domain_id" = :DomainID and "id" = :ID
		`

	args := map[string]any{
		"DomainID": search.DomainID,
		"ID":       search.ID,
	}

	var result *model.OnlineSkills
	if err := s.GetReplica().WithContext(ctx).SelectOne(&result, query, args); err != nil {
		return nil, model.NewCustomCodeError("sqlstore.online_skills_store.get", err.Error(), extractCodeFromErr(err))
	}

	return result, nil
}
