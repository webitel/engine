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
	Options []QuestionOption `json:"options,omitempty"`
	//score
	Min float32 `json:"min,omitempty"`
	Max float32 `json:"max,omitempty"`
}

type QuestionOption struct {
	Name  string  `json:"name"`
	Score float32 `json:"score"`
}

func (q Questions) SumMax(required bool) int {
	i := 0
	for _, v := range q {
		if v.Required == required {
			switch v.Type {
			case QuestionTypeScore:
				i += int(v.Max)
			case QuestionTypeOptions:
				i += maxScoreOption(v.Options)
			}
		}
	}

	return i
}

func (q *Questions) ToJson() []byte {
	if q == nil {
		return nil
	}

	data, _ := json.Marshal(q)
	return data
}

func maxScoreOption(ops []QuestionOption) int {
	if len(ops) == 0 {
		return 0
	}

	max := ops[0].Score
	for _, v := range ops {
		if v.Score > max {
			max = v.Score
		}
	}

	return int(max)
}

func (q *Question) ValidAnswer(a QuestionAnswer) bool {
	switch q.Type {
	case QuestionTypeScore:
		return (q.Min <= a.Score) && (a.Score <= q.Max)
	case QuestionTypeOptions:
		for _, v := range q.Options {
			if v.Score == a.Score {
				return true
			}
		}
	}

	return false
}
