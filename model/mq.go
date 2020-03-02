package model

const (
	MQ_APP_EXCHANGE         = "engine"
	MQ_CALL_EXCHANGE        = "call"
	MQ_USER_STATUS_EXCHANGE = "webitel"

	MQ_CALL_TEMPLATE_ROUTING_KEY        = "events.*.%d.%d"
	MQ_USER_STATUS_TEMPLATE_ROUTING_KEY = "user_state.*.%d.*"

	MQ_DIRECT = "direct"
	MQ_TOPIC  = "topic"
)

type BindQueueEvent struct {
	Id       string
	UserId   int64
	Routing  string
	Exchange string
}

type GetAllBindings func() []*BindQueueEvent
