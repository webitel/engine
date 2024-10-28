package model

type AgentTeam struct {
	DomainRecord
	Name              string `json:"name" db:"name"`
	Description       string `json:"description" db:"description"`
	Strategy          string `json:"strategy" db:"strategy"`
	MaxNoAnswer       int16  `json:"max_no_answer" db:"max_no_answer"`
	WrapUpTime        int16  `json:"wrap_up_time" db:"wrap_up_time"`
	NoAnswerDelayTime int16  `json:"no_answer_delay_time" db:"no_answer_delay_time"`
	CallTimeout       int16  `json:"call_timeout" db:"call_timeout"`
	InviteChatTimeout int16  `json:"invite_chat_timeout" db:"invite_chat_timeout"`
	TaskAcceptTimeout int16  `json:"task_accept_timeout" db:"task_accept_timeout"`

	Admin               []*Lookup `json:"admin" db:"admin"`
	ForecastCalculation *Lookup   `json:"forecast_calculation" db:"forecast_calculation_id"`
}

func (team AgentTeam) DefaultOrder() string {
	return "id"
}

func (team AgentTeam) AllowFields() []string {
	return append(team.DefaultFields(), "forecast_calculation")
}

func (team AgentTeam) DefaultFields() []string {
	return []string{"id", "name", "description", "strategy", "max_no_answer", "wrap_up_time",
		"no_answer_delay_time", "call_timeout", "updated_at", "admin", "invite_chat_timeout", "task_accept_timeout"}
}

func (team AgentTeam) EntityName() string {
	return "cc_team_list"
}

type SearchAgentTeam struct {
	ListRequest
	Ids      []uint32
	Strategy []string
	AdminIds []uint32
}

func (team *AgentTeam) IsValid() AppError {
	return nil
}

func (team *AgentTeam) ForecastCalculationId() *int64 {
	if team.ForecastCalculation != nil {
		return NewInt64(int64(team.ForecastCalculation.Id))
	}

	return nil
}
