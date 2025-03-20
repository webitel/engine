package rabbit

import (
	"fmt"
	"github.com/webitel/engine/model"
)

func (a *AMQP) BindCallEvents(domainId, userId int64) error {
	return a.channel.QueueBind(a.queueForCall.Name, fmt.Sprintf(model.CallRoutingTemplate, domainId, userId), model.CallExchange, false, nil)
}

func (a *AMQP) UnBindCallEvents(domainId, userId int64) error {
	return a.channel.QueueUnbind(a.queueForCall.Name, fmt.Sprintf(model.CallRoutingTemplate, domainId, userId), model.CallExchange, nil)
}
