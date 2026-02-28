//go:build e2e

package e2etest

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Employee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Email struct {
	From    Employee `json:"from"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

func getRabbitMqClient() (*amqp.Connection, *amqp.Channel, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}
