package model

import "encoding/json"

const (
	SOCKET_MAX_MESSAGE_SIZE_KB  = 8 * 1024 // 8KB
	PING_TIMEOUT_BUFFER_SECONDS = 5
)

// event.open_socket.DOMAIN.USER
// event.close_socket.DOMAIN.USER
// text/json
type RegisterToWebsocketEvent struct {
	UserId    int64  `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
	AppId     string `json:"app_id"`

	Addr     string `json:"addr"`
	SocketId string `json:"socket_id"`
}

func (o *RegisterToWebsocketEvent) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}
