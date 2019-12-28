package model

type ResourceInTeam struct {
	Id          int64   `json:"id" db:"id"`
	TeamId      int64   `json:"team_id" db:"team_id"`
	Agent       *Lookup `json:"agent" db:"agent"`
	Skill       *Lookup `json:"skill" db:"skill"`
	Bucket      *Lookup `json:"bucket" db:"bucket"`
	Lvl         int     `json:"lvl" db:"lvl"`
	MinCapacity int     `json:"min_capacity" db:"min_capacity"`
	MaxCapacity int     `json:"max_capacity" db:"max_capacity"`
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

func (r *ResourceInTeam) BucketId() *int {
	if r.Bucket == nil {
		return nil
	}
	return NewInt(r.Bucket.Id)
}

func (a *ResourceInTeam) IsValid() *AppError {
	//FIXME
	return nil
}
