package chat_manager

const (
	AgentLeave LeaveCause = "agent_leave"
)

type ChatFile struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
	Mime string `json:"mime"`
	Size int64  `json:"size"`
}

type LeaveCause string

func (l LeaveCause) String() string {
	return string(l)
}

const (
	FlowEnd     = "flow_end"
	ClientLeave = "client_leave"
	FlowErr     = "flow_err"
)

type CloseCause string

func (l CloseCause) String() string {
	return string(l)
}
