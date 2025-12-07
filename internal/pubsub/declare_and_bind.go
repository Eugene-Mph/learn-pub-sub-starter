package pubsub

import (
	"fmt"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueTransient SimpleQueueType = iota // 0
	SimpleQueueDurable                          // 1
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {

	Ch, err := conn.Channel()

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}

	queue, err := Ch.QueueDeclare(
		queueName,
		queueType != SimpleQueueTransient, // durable
		queueType == SimpleQueueTransient, // autoDelete
		queueType == SimpleQueueTransient, // exclusive
		false,                             // no-wait
		amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDLX,
		},                               
	)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	err = Ch.QueueBind(
		queue.Name, // queue name
		key,        // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // args
	)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}

	return Ch, queue, nil

}
