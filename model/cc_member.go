package model

import "encoding/json"

type Member struct {
	Id             int64                 `json:"id" db:"id"`
	Queue          Lookup                `json:"queue" db:"queue"`
	CreatedAt      int64                 `json:"created_at" db:"created_at"`
	QueueId        int64                 `json:"queue_id" db:"queue_id"` //FIXME delete attr
	Priority       int                   `json:"priority" db:"priority"`
	ExpireAt       *int64                `json:"expire_at" db:"expire_at"`
	MinOfferingAt  int64                 `json:"min_offering_at" db:"min_offering_at"`
	Name           string                `json:"name" db:"name"`
	StopCause      *string               `json:"stop_cause" db:"stop_cause"`
	Variables      StringMap             `json:"variables" db:"variables"`
	LastActivityAt int64                 `json:"last_hangup_at" db:"last_hangup_at"`
	Attempts       int                   `json:"attempts" db:"attempts"`
	Timezone       Lookup                `json:"timezone" db:"timezone"`
	Bucket         *Lookup               `json:"bucket" db:"bucket"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
	Skills         Int64Array            `json:"skills" db:"skills"`
	StopAt         *int64                `json:"stop_at" db:"stop_at"`
	Reserved       bool                  `json:"reserved" db:"reserved"`
}

type MemberView struct {
}

type SearchMemberRequest struct {
	ListRequest
	Id          *int64
	QueueId     *int64
	Destination *string
	BucketId    *int32
}

type Attempt struct {
	Id          int64               `json:"id" db:"id"`
	Member      Lookup              `json:"member" db:"member"`
	Queue       Lookup              `json:"queue" db:"queue"`
	CreatedAt   int64               `json:"created_at" db:"created_at"`
	Destination MemberCommunication `json:"destination" db:"destination"`
	Weight      int                 `json:"weight" db:"weight"`
	OriginateAt int64               `json:"originate_at" db:"originate_at"`
	AnsweredAt  int64               `json:"answered_at" db:"answered_at"`
	BridgedAt   int64               `json:"bridged_at" db:"bridged_at"`
	HangupAt    int64               `json:"hangup_at" db:"hangup_at"`
	Resource    *Lookup             `json:"resource" db:"resource"`
	LegAId      *string             `json:"leg_a_id" db:"leg_a_id"`
	LegBId      *string             `json:"leg_b_id" db:"leg_b_id"`
	Result      *string             `json:"result" db:"result"`
	Agent       *Lookup             `json:"agent" db:"agent"`
	Bucket      *Lookup             `json:"bucket" db:"bucket"`
	Variables   map[string]string   `json:"variables" db:"variables"`
	Active      bool                `json:"active" db:"active"`
}

type MemberAttempt struct {
	Id          int64             `json:"id" db:"id"`
	CreatedAt   int64             `json:"created_at" db:"created_at"`
	Destination string            `json:"destination" db:"destination"`
	Weight      int               `json:"weight" db:"weight"`
	OriginateAt int64             `json:"originate_at" db:"originate_at"`
	AnsweredAt  int64             `json:"answered_at" db:"answered_at"`
	BridgedAt   int64             `json:"bridged_at" db:"bridged_at"`
	HangupAt    int64             `json:"hangup_at" db:"hangup_at"`
	Resource    Lookup            `json:"resource" db:"resource"`
	LegAId      *string           `json:"leg_a_id" db:"leg_a_id"`
	LegBId      *string           `json:"leg_b_id" db:"leg_b_id"`
	Node        *string           `json:"node" json:"node"`
	Result      *string           `json:"result" db:"result"`
	Agent       *Lookup           `json:"agent" db:"agent"`
	Bucket      *Lookup           `json:"bucket" db:"bucket"`
	Logs        []byte            `json:"logs" db:"logs"`
	Active      bool              `json:"active" db:"active"`
	Variables   map[string]string `json:"variables" db:"variables"`
}

type SearchAttempts struct {
	ListRequest
	CreatedAt FilterBetween `json:"created_at" db:"created_at"`
	Id        *int64        `json:"id" db:"id"`
	MemberId  *int64        `json:"member_id" db:"member_id"`
	//ResourceId  *int32        `json:"resource_id" db:"resource_id" `
	QueueId  *int64 `json:"queue_id" db:"queue_id"`
	BucketId *int64 `json:"bucket_id" db:"bucket_id"`
	//Destination *string       `json:"destination" db:"destination"`
	AgentId *int64  `json:"agent_id" db:"agent_id"`
	Result  *string `json:"result" db:"result"`
}

type MembersAttempt struct {
	Member Lookup
	MemberAttempt
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
