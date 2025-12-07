package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackReque
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	excahge,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient" queue's
	handler func(T) AckType,
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

			ackType := handler(payload)

			switch ackType {

			case Ack:
				message.Ack(false)
				fmt.Println("Ack")

			case NackDiscard:
				message.Nack(false, false)
				fmt.Println("NackDiscard")

			case NackReque:
				message.Nack(false, true)
				fmt.Println("NackReque")

			}

		}

	}()

	return nil
}
