package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	excahge,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient" queue's
	handler func(T),
) error {

	ch, queue, err := DeclareAndBind(conn, excahge, queueName, key, queueType)

	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	messages, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	go func() {

		defer ch.Close()

		for message := range messages {

			var payload T
			err := json.Unmarshal(message.Body, &payload)

			if err != nil {
				fmt.Printf("could not unmarshal message: %v", err)
				continue
			}

			handler(payload)
			err = message.Ack(false)

			if err != nil {
				fmt.Printf("could not delete message: %v", err)
				continue
			}

		}

	}()

	return nil
}
