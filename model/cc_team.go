package model

type AgentTeam struct {
	DomainRecord
	Name              string `json:"name" db:"name"`
	Description       string `json:"description" db:"description"`
	Strategy          string `json:"strategy" db:"strategy"`
	MaxNoAnswer       int16  `json:"max_no_answer" db:"max_no_answer"`
	WrapUpTime        int16  `json:"wrap_up_time" db:"wrap_up_time"`
	RejectDelayTime   int16  `json:"reject_delay_time" db:"reject_delay_time"`
	BusyDelayTime     int16  `json:"busy_delay_time" db:"busy_delay_time"`
	NoAnswerDelayTime int16  `json:"no_answer_delay_time" db:"no_answer_delay_time"`
	CallTimeout       int16  `json:"call_timeout" db:"call_timeout"`
}

type SearchAgentTeam struct {
	ListRequest
}

func (team *AgentTeam) IsValid() *AppError {
	return nil
}
