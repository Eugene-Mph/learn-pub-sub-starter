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
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)

	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	defer conn.Close()

	userName, err := gamelogic.ClientWelcome()

	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	fmt.Printf("Queue %s declared and bound!", queue.Name)

	gameState := gamelogic.NewGameState(userName)

	for {
		inputWords := gamelogic.GetInput()

		if len(inputWords) == 0 {
			continue
		}

		switch inputWords[0] {

		case "spawn":

			err := gameState.CommandSpawn(inputWords)

			if err != nil {
				fmt.Println(err)
				continue
			}

		case "move":

			_, err = gameState.CommandMove(inputWords)

			if err != nil {
				fmt.Println(err)
				continue
			}

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("unknown command")

		}
	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("'\nRabbitMQ connection closed.")
}
