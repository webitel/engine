package model

import (
	"encoding/json"
	"github.com/nicksnyder/go-i18n/i18n"
	"io"
)

type WebSocketRequest struct {
	// Client-provided fields
	Seq    int64                  `json:"seq"`
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`

	Session Session            `json:"-"`
	T       i18n.TranslateFunc `json:"-"`
	Locale  string             `json:"-"`
}

func (o *WebSocketRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketRequestFromJson(data io.Reader) *WebSocketRequest {
	var o *WebSocketRequest
	json.NewDecoder(data).Decode(&o)
	return o
}
