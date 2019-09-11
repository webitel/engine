package rabbit

import "github.com/webitel/engine/model"

func (a *AMQP) DeclareExchanges() error {
	err := a.channel.ExchangeDeclare(
		model.MQ_CALL_EXCHANGE,
		model.MQ_TOPIC,
		true,
		false,
		false,
		false,
		nil,
	)

	return err
}
