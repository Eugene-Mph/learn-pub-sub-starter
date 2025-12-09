package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	decoder func([]byte) (T, error),
) error {

	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)

	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	err = ch.Qos(10, 0, false)

	if err != nil {
		return fmt.Errorf("could not set Qos: %v", err)
	}

	msgs, err := ch.Consume(
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

		for msg := range msgs {

			payload, err := decoder(msg.Body)

			if err != nil {
				fmt.Printf("could not decode message: %v", err)
				continue
			}

			switch handler(payload) {

			case Ack:
				msg.Ack(false)

			case NackDiscard:
				msg.Nack(false, false)

			case NackRequeue:
				msg.Nack(false, true)
			}

		}
	}()

	return nil
}

func jsonDecoder[T any](data []byte) (T, error) {
	var payload T
	err := json.Unmarshal(data, &payload)
	return payload, err
}

func gobDecoder[T any](data []byte) (T, error) {
	buffer := bytes.NewBuffer(data)
	var payload T

	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&payload)
	return payload, err
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType) error {

	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		gobDecoder[T],
	)
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		jsonDecoder[T],
	)
}
