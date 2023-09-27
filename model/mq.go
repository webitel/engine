package model

import "encoding/json"

const (
	AppExchange        = "engine"
	CallExchange       = "call"
	CallCenterExchange = "callcenter"
	ChatExchange       = "chat"

	MQ_USER_STATUS_EXCHANGE = "webitel"

	CallRoutingTemplate                 = "events.*.*.%d.%d"
	MQ_USER_STATUS_TEMPLATE_ROUTING_KEY = "presence.user.%d.%d"

	MQ_TOPIC = "topic"
)

const (
	CallCenterAgentStateTemplate = "events.status.%d.%d"
)

type BindQueueEvent struct {
	Id       string
	UserId   int64
	Routing  string
	Exchange string
}

type GetAllBindings func() []*BindQueueEvent

type ExecFlow struct {
	DomainId  int64     `json:"domain_id"`
	SchemaId  int32     `json:"schema_id"`
	Variables Variables `json:"variables"`
}

func StructToVariable(input interface{}) Variables {
	var res Variables

	data, err := json.Marshal(input)
	if err != nil {
		// todo
		return nil
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		// todo
	}

	return res
}
