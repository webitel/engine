package model

const (
	MQ_APP_EXCHANGE    = "engine"
	CallExchange       = "call"
	CallCenterExchange = "callcenter"

	MQ_USER_STATUS_EXCHANGE = "webitel"

	CallRoutingTemplate                 = "events.*.*.%d.%d"
	MQ_USER_STATUS_TEMPLATE_ROUTING_KEY = "presence.user.%d.*"

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
