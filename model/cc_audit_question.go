package model

import "encoding/json"

type QuestionType string

const (
	QuestionTypeScore   QuestionType = "score"
	QuestionTypeOptions QuestionType = "options"
)

type Questions []Question

type Question struct {
	Type     QuestionType `json:"type"`
	Required bool         `json:"required"`
	Question string       `json:"question"`
	//options
	Options []QuestionOption `json:"options"`
	//score
	Min int32 `json:"min"`
	Max int32 `json:"max"`
}

type QuestionOption struct {
	Name  string `json:"name"`
	Score int32  `json:"score"`
}

func (q *Questions) ToJson() []byte {
	if q == nil {
		return nil
	}

	data, _ := json.Marshal(q)
	return data
}
