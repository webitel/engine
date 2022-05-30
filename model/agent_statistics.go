package model

type AgentStatistics struct {
	Utilization float32 `json:"utilization" db:"utilization"`
	Occupancy   float32 `json:"occupancy" db:"occupancy"`

	CallAbandoned int32   `json:"call_abandoned" db:"call_abandoned"`
	CallHandled   int32   `json:"call_handled" db:"call_handled"`
	AvgTalkSec    float32 `json:"avg_talk_sec" db:"avg_talk_sec"`
	AvgHoldSec    float32 `json:"avg_hold_sec" db:"avg_hold_sec"`

	ChatAccepts int32   `json:"chat_accepts" db:"chat_accepts"`
	ChatAht     float32 `json:"chat_aht" db:"chat_aht"`
}
