package model

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

func (a *ResourceInTeam) IsValid() *AppError {
	//FIXME
	return nil
}
