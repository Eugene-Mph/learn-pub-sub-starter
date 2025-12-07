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

	puhlishCh, err := conn.Channel()

	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	userName, err := gamelogic.ClientWelcome()

	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	gameState := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)

	if err != nil {
		log.Fatalf("could not subscibe to pause:%v", err)
	}

	moveQueueName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName)

	err = pubsub.SubscribeJSON(
		conn,
		string(routing.ExchangePerilTopic),
		moveQueueName,
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, puhlishCh),
	)

	if err != nil {
		log.Fatalf("could not subscribe to army_moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		string(routing.ExchangePerilTopic),
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.SimpleQueueDurable,
		hanlerWar(gameState),
	)

	if err != nil {
		log.Fatalf("could not subscribe to war declaration: %v", err)
	}

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

			armyMove, err := gameState.CommandMove(inputWords)

			if err != nil {
				fmt.Println(err)
				continue
			}

			err = pubsub.PublishJson(
				puhlishCh,
				routing.ExchangePerilTopic,
				moveQueueName,
				armyMove,
			)

			if err != nil {
				log.Printf("error: %v", err)
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

}
