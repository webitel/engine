package model

import "net/http"

type ResourceInTeam struct {
	Id          int64     `json:"id" db:"id"`
	TeamId      int64     `json:"team_id" db:"team_id"`
	Agent       *Lookup   `json:"agent" db:"agent"`
	Skill       *Lookup   `json:"skill" db:"skill"`
	Buckets     []*Lookup `json:"buckets" db:"buckets"`
	Lvl         int       `json:"lvl" db:"lvl"`
	MinCapacity int       `json:"min_capacity" db:"min_capacity"`
	MaxCapacity int       `json:"max_capacity" db:"max_capacity"`
}

type ResourceInTeamPatch struct {
	Agent       *Lookup   `json:"agent" db:"agent"`
	Skill       *Lookup   `json:"skill" db:"skill"`
	Buckets     []*Lookup `json:"buckets" db:"buckets"`
	Lvl         *int      `json:"lvl" db:"lvl"`
	MinCapacity *int      `json:"min_capacity" db:"min_capacity"`
	MaxCapacity *int      `json:"max_capacity" db:"max_capacity"`
}

func (r *ResourceInTeam) Patch(p *ResourceInTeamPatch) {
	if p.Agent != nil {
		r.Agent = p.Agent
	}

	if p.Skill != nil {
		r.Skill = p.Skill
	}

	if p.Buckets != nil {
		r.Buckets = p.Buckets
	}

	if p.Lvl != nil {
		r.Lvl = *p.Lvl
	}

	if p.MinCapacity != nil {
		r.MinCapacity = *p.MinCapacity
	}

	if p.MaxCapacity != nil {
		r.MaxCapacity = *p.MaxCapacity
	}
}

type SearchResourceInTeam struct {
	ListRequest
	OnlyAgents bool
}

func (r *ResourceInTeam) AgentId() *int64 {
	if r.Agent == nil {
		return nil
	}
	return NewInt64(int64(r.Agent.Id))
}

func (r *ResourceInTeam) SkillId() *int64 {
	if r.Skill == nil {
		return nil
	}
	return NewInt64(int64(r.Skill.Id))
}

func (r *ResourceInTeam) BucketIds() []int {
	if r.Buckets == nil {
		return nil
	}

	var res = make([]int, 0, len(r.Buckets))
	for _, v := range r.Buckets {
		res = append(res, v.Id)
	}
	return res
}

func (r *ResourceInTeam) IsValid() *AppError {
	if r.Agent != nil && r.Skill != nil {
		return NewAppError("ResourceInTeam.IsValid", "model.resource_in_team.is_valid.agent_skill.app_error", nil, "", http.StatusBadRequest)
	}
	return nil
}
