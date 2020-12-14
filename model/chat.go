package model

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
	Id     string `json:"id"`
	UserId int64  `json:"user_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}

type Conversation struct {
	Id        string         `json:"id" db:"id"`
	InviteId  *string        `json:"invite_id" db:"invite_id"`
	ChannelId *string        `json:"channel_id" db:"channel_id"`
	Title     *string        `json:"title" db:"title"`
	CreatedAt int64          `json:"created_at" db:"created_at"`
	UpdatedAt int64          `json:"updated_at" db:"updated_at"`
	JoinedAt  *int64         `json:"joined_at" db:"joined_at"`
	Members   []*ChatMember  `json:"members" db:"members"`
	Messages  []*ChatMessage `json:"messages" db:"messages"`
}

func NewWebSocketChatEvent(event *ChatEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CHAT)
	e.Add("data", event.Data)
	e.Add("action", event.Event)

	return e
}
