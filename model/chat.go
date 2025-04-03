package model

const (
	AgentLeave LeaveCause = "agent_leave"
)

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

type ChatEvent struct {
	Event    string `json:"event"`
	UserId   int64  `json:"user_id"`
	DomainId int64  `json:"domain_id"`
	Data     map[string]interface{}
}

type ChatFile struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
	Mime string `json:"mime"`
	Size int64  `json:"size"`
}

type ChatMessage struct {
	Id        int64                  `json:"id"`
	ChannelId string                 `json:"channel_id"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
	Type      string                 `json:"type"`
	Text      *string                `json:"text,omitempty"`
	File      map[string]interface{} `json:"file,omitempty"`
}

type ChatMember struct {
	Id         string   `json:"id"`
	UserId     int64    `json:"user_id"`
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	ExternalId *string  `json:"external_id,omitempty"`
	Via        *Gateway `json:"via,omitempty"`
}

type Gateway struct {
	Id   int64  `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type Conversation struct {
	Id        string         `json:"id" db:"id"`
	InviteId  *string        `json:"invite_id" db:"invite_id"`
	ChannelId *string        `json:"channel_id" db:"channel_id"`
	Title     *string        `json:"title" db:"title"`
	CreatedAt int64          `json:"created_at" db:"created_at"`
	UpdatedAt int64          `json:"updated_at" db:"updated_at"`
	JoinedAt  *int64         `json:"joined_at" db:"joined_at"`
	ClosedAt  *int64         `json:"closed_at" db:"closed_at"`
	Variables StringMap      `json:"variables"`
	Members   []*ChatMember  `json:"members" db:"members"`
	Messages  []*ChatMessage `json:"messages" db:"messages"`
	LeavingAt *int64         `json:"leaving_at" db:"leaving_at"`
	Task      *CCTask        `json:"task"`
}

func NewWebSocketChatEvent(event *ChatEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CHAT)
	e.Add("data", event.Data)
	e.Add("action", event.Event)

	return e
}
