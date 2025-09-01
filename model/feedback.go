package model

import "encoding/json"

type FeedbackKey struct {
	DomainId int64     `json:"dc"`
	Source   string    `json:"src"`
	SourceId string    `json:"sid"`
	Payload  StringMap `json:"p,omitempty"`
}

type Feedback struct {
	Id          int64             `json:"id" db:"id"`
	SourceId    string            `json:"source_id" db:"source_id"`
	Source      string            `json:"source" db:"source"`
	Payload     map[string]string `json:"payload" db:"payload"`
	CreatedAt   int64             `json:"created_at" db:"created_at"`
	Rating      float32           `json:"rating" db:"rating"`
	Description string            `json:"description" db:"description"`
}

func (f *FeedbackKey) ToJson() []byte {
	d, _ := json.Marshal(f)
	return d
}

func FeedbackKeyFromJson(src []byte) (FeedbackKey, AppError) {
	var f FeedbackKey
	err := json.Unmarshal(src, &f)
	if err != nil {
		return f, NewBadRequestError("model.feedback.parse", err.Error())
	}
	return f, nil
}
