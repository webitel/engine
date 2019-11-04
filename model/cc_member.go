package model

import "encoding/json"

type Member struct {
	Id             int64                 `json:"id" db:"id"`
	QueueId        int64                 `json:"queue_id" db:"queue_id"`
	Priority       int                   `json:"priority" db:"priority"`
	ExpireAt       *int64                `json:"expire_at" db:"expire_at"`
	Name           string                `json:"name" db:"name"`
	StopCause      *string               `json:"stop_cause" db:"stop_cause"`
	Variables      StringMap             `json:"variables" db:"variables"`
	LastActivityAt int64                 `json:"last_activity_at" db:"last_activity_at"`
	Attempts       int                   `json:"attempts" db:"attempts"`
	Timezone       Lookup                `json:"timezone" db:"timezone"`
	Bucket         *Lookup               `json:"bucket" db:"bucket"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
}

type MemberCommunication struct {
	Id             int64  `json:"id"`
	Priority       int    `json:"priority" db:"priority"`
	Destination    string `json:"destination" db:"destination"`
	State          int    `json:"state" db:"state"`
	Description    string `json:"description" db:"description"`
	LastActivityAt int64  `json:"last_activity_at" db:"last_activity_at"`
	Attempts       int    `json:"attempts" db:"attempts"`
	LastCause      string `json:"last_cause" db:"last_cause"`
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
