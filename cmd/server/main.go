package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)

	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	publishCh, err := conn.Channel()

	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	gamelogic.PrintServerHelp()

	for {
		inputWords := gamelogic.GetInput()

		switch inputWords[0] {
		case "pause":
			stateControl(
				publishCh,
				"paused",
				true,
			)

		case "resume":
			stateControl(
				publishCh,
				"resumed",
				false,
			)
		case "quit":
			fmt.Println("goodbye")
			return
		default:
			fmt.Println("unknown command")
		}
	}

}

func stateControl(ch *amqp.Channel, message string, pause bool) {
	err := pubsub.PublishJson(
		ch,
		string(routing.ExchangePerilDirect),
		string(routing.PauseKey),
		routing.PlayingState{
			IsPaused: pause,
		},
	)

	if err != nil {
		log.Printf("could not publish time: %v", err)
	}

	fmt.Printf("Publishing %s game state\n", message)
}
