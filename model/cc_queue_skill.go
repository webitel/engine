package model

type QueueSkill struct {
	Id      uint32 `json:"id" db:"id"`
	QueueId uint32 `json:"queue_id" db:"queue_id"`

	Skill       Lookup    `json:"skill" db:"skill"`
	Buckets     []*Lookup `json:"buckets" db:"buckets"`
	Lvl         int       `json:"lvl" db:"lvl"`
	MinCapacity int       `json:"min_capacity" db:"min_capacity"`
	MaxCapacity int       `json:"max_capacity" db:"max_capacity"`
	Enabled     bool      `json:"enabled" json:"enabled"`
}

type QueueSkillPatch struct {
	Skill       *Lookup   `json:"skill" db:"skill"`
	Buckets     []*Lookup `json:"buckets" db:"buckets"`
	Lvl         *int      `json:"lvl" db:"lvl"`
	MinCapacity *int      `json:"min_capacity" db:"min_capacity"`
	MaxCapacity *int      `json:"max_capacity" db:"max_capacity"`
	Enabled     *bool     `json:"enabled" db:"enabled"`
}

type SearchQueueSkill struct {
	ListRequest
	QueueId     uint32   `json:"queue_id"`
	Ids         []uint32 `json:"ids"`
	SkillIds    []uint32 `json:"skill_ids"`
	BucketIds   []uint32 `json:"bucket_ids"`
	Lvl         []int32  `json:"lvl"`
	MinCapacity []int32  `json:"min_capacity"`
	MaxCapacity []int32  `json:"max_capacity"`
	Enabled     *bool    `json:"enabled"`
}

func (q QueueSkill) AllowFields() []string {
	return q.DefaultFields()
}

func (q QueueSkill) DefaultOrder() string {
	return "-id"
}

func (q QueueSkill) DefaultFields() []string {
	return []string{"id", "skill", "buckets", "lvl", "min_capacity", "max_capacity", "enabled"}
}

func (q QueueSkill) EntityName() string {
	return "cc_queue_skill_list"
}

func (q *QueueSkill) BucketIds() []uint32 {
	if q.Buckets == nil {
		return nil
	}

	var res = make([]uint32, 0, len(q.Buckets))
	for _, v := range q.Buckets {
		res = append(res, uint32(v.Id))
	}

	return res
}

func (q *QueueSkill) Patch(patch *QueueSkillPatch) {

	if patch.Skill != nil {
		q.Skill = *patch.Skill
	}
	if patch.Buckets != nil {
		q.Buckets = patch.Buckets
	}
	if patch.Lvl != nil {
		q.Lvl = *patch.Lvl
	}
	if patch.MinCapacity != nil {
		q.MinCapacity = *patch.MinCapacity
	}
	if patch.MaxCapacity != nil {
		q.MaxCapacity = *patch.MaxCapacity
	}
	if patch.Enabled != nil {
		q.Enabled = *patch.Enabled
	}
}

// Todo
func (q *QueueSkill) IsValid() AppError {
	if q.MinCapacity < 0 || q.MinCapacity > 100 {
		return NewBadRequestError("queue_skill.valid.min_capacity", "Min capacity must be between 0 and 100")
	}
	if q.MaxCapacity < 0 || q.MaxCapacity > 100 {
		return NewBadRequestError("queue_skill.valid.max_capacity", "Max capacity must be between 0 and 100")
	}
	return nil
}
