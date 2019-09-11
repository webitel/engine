package model

const (
	MQ_CALL_EXCHANGE = "TAP.CALLS"

	MQ_CALL_TEMPLATE_ROUTING_KEY = "call.%d.%d"

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
