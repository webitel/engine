package model

const (
	MQ_CALL_EXCHANGE = "call"

	MQ_CALL_TEMPLATE_ROUTING_KEY = "events.%d.%d"

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
