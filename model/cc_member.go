package model

import "encoding/json"

type Member struct {
	Id             int64                 `json:"id" db:"id"`
	QueueId        int64                 `json:"queue_id" db:"queue_id"`
	Priority       int                   `json:"priority" db:"priority"`
	ExpireAt       *int64                `json:"expire_at" db:"expire_at"`
	MinOfferingAt  int64                 `json:"min_offering_at" db:"min_offering_at"`
	Name           string                `json:"name" db:"name"`
	StopCause      *string               `json:"stop_cause" db:"stop_cause"`
	Variables      StringMap             `json:"variables" db:"variables"`
	LastActivityAt int64                 `json:"last_activity_at" db:"last_activity_at"`
	Attempts       int                   `json:"attempts" db:"attempts"`
	Timezone       Lookup                `json:"timezone" db:"timezone"`
	Bucket         *Lookup               `json:"bucket" db:"bucket"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
	Skills         Int64Array            `json:"skills" db:"skills"`
}

type MemberAttempt struct {
	Id          int64   `json:"id" db:"id"`
	CreatedAt   int64   `json:"created_at" db:"created_at"`
	Destination string  `json:"destination" db:"destination"`
	Weight      int     `json:"weight" db:"weight"`
	OriginateAt int64   `json:"originate_at" db:"originate_at"`
	AnsweredAt  int64   `json:"answered_at" db:"answered_at"`
	BridgedAt   int64   `json:"bridged_at" db:"bridged_at"`
	HangupAt    int64   `json:"hangup_at" db:"hangup_at"`
	Resource    Lookup  `json:"resource" db:"resource"`
	LegAId      *string `json:"leg_a_id" db:"leg_a_id"`
	LegBId      *string `json:"leg_b_id" db:"leg_b_id"`
	Node        *string `json:"node" json:"node"`
	Result      *string `json:"result" db:"result"`
	Agent       *Lookup `json:"agent" db:"agent"`
	Bucket      *Lookup `json:"bucket" db:"bucket"`
	Logs        []byte  `json:"logs" db:"logs"`
	Success     *bool   `json:"success" db:"success"`
	Active      bool    `json:"active" db:"active"`
}

func (a *MemberAttempt) IsValid() *AppError {
	//FIXME
	return nil
}

func (member *Member) GetSkills() Int64Array {
	if member.Skills == nil {
		return Int64Array{}
	}
	return member.Skills
}

type MemberCommunication struct {
	Id             int64   `json:"id"`
	Destination    string  `json:"destination" db:"destination"`
	Type           Lookup  `json:"type"`
	Priority       int     `json:"priority" db:"priority"`
	State          int     `json:"state" db:"state"`
	Description    string  `json:"description" db:"description"`
	LastActivityAt int64   `json:"last_activity_at" db:"last_activity_at"`
	Attempts       int     `json:"attempts" db:"attempts"`
	LastCause      string  `json:"last_cause" db:"last_cause"`
	Resource       *Lookup `json:"resource" db:"resource"`
	Display        string  `json:"display" db:"display"`
}

func (m *Member) ToJsonCommunications() string {
	data, _ := json.Marshal(m.Communications)
	return string(data)
}

func (m *Member) GetBucketId() *int64 {
	if m.Bucket != nil {
		return NewInt64(int64(m.Bucket.Id))
	}
	return nil
}

func (m *Member) GetExpireAt() int64 {
	if m.ExpireAt != nil {
		return *m.ExpireAt
	}
	return 0
}

func (m *Member) IsValid() *AppError {
	//FIXME
	return nil
}
