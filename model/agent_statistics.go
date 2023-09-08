package model

type AgentStatistics struct {
	Utilization float32 `json:"utilization" db:"utilization"`
	Occupancy   float32 `json:"occupancy" db:"occupancy"`

	CallAbandoned int32 `json:"call_abandoned" db:"call_abandoned"`
	CallHandled   int32 `json:"call_handled" db:"call_handled"`
	CallMissed    int32 `json:"call_missed" db:"call_missed"`
	CallInbound   int32 `json:"call_inbound" db:"call_inbound"`

	AvgTalkSec float32 `json:"avg_talk_sec" db:"avg_talk_sec"`
	AvgHoldSec float32 `json:"avg_hold_sec" db:"avg_hold_sec"`

	ChatAccepts      int32   `json:"chat_accepts" db:"chat_accepts"`
	ChatAht          float32 `json:"chat_aht" db:"chat_aht"`
	ScoreRequiredAvg float32 `json:"score_required_avg" db:"score_required_avg"`
	ScoreOptionalAvg float32 `json:"score_optional_avg" db:"score_optional_avg"`
	ScoreRequiredSum float32 `json:"score_required_sum" db:"score_required_sum"`
	ScoreOptionalSum float32 `json:"score_optional_sum" db:"score_optional_sum"`
	ScoreCount       int64   `json:"score_count" db:"score_count"`
	SumTalkSec       int64   `json:"sum_talk_sec" db:"sum_talk_sec"`
	VoiceMail        int32   `json:"voice_mail" db:"voice_mail"`
	Available        int32   `json:"available" db:"available"`
	Online           int32   `json:"online" db:"online"`
	Processing       int32   `json:"processing" db:"processing"`
	TaskAccepts      int32   `json:"task_accepts" db:"task_accepts"`
	QueueTalkSec     int32   `json:"queue_talk_sec" db:"queue_talk_sec"`
}
