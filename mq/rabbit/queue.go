package rabbit

import (
	"fmt"
	"github.com/webitel/engine/model"
)

func (a *AMQP) DeclareQueues() error {
	var err error

	a.queueForCall, err = a.channel.QueueDeclare(
		fmt.Sprintf("engine.%s", model.NewId()),
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	return err
}

func (a *AMQP) BindCallEvents(domainId, userId int64) error {
	return a.channel.QueueBind(a.queueForCall.Name, fmt.Sprintf(model.MQ_CALL_TEMPLATE_ROUTING_KEY, domainId, userId), model.CallExchange, false, nil)
}

func (a *AMQP) UnBindCallEvents(domainId, userId int64) error {
	return a.channel.QueueUnbind(a.queueForCall.Name, fmt.Sprintf(model.MQ_CALL_TEMPLATE_ROUTING_KEY, domainId, userId), model.CallExchange, nil)
}
