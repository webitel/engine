package model

import (
	"encoding/json"
	"fmt"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"io"
)

type WebSocketRequest struct {
	// Client-provided fields
	Seq    int64                  `json:"seq"`
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`

	Session auth_manager.Session `json:"-"`
	T       i18n.TranslateFunc   `json:"-"`
	Locale  string               `json:"-"`
}

func (o *WebSocketRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *WebSocketRequest) GetFieldString(name string) string {
	if tmp, ok := o.Data[name]; ok {
		return fmt.Sprintf("%v", tmp)
	}
	return ""
}

func WebSocketRequestFromJson(data io.Reader) *WebSocketRequest {
	var o *WebSocketRequest
	json.NewDecoder(data).Decode(&o)
	return o
}
