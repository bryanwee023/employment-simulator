//go:build e2e

package e2etest

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

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

type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
}

func (e LogEntry) String() string {
	return fmt.Sprintf("[%s] %s: %s", e.Timestamp.Format(time.DateTime), e.Event, e.Data)
}

func getEmployeeURL(i int) string {
	return fmt.Sprintf("http://localhost:%d", 8081+i)
}

func getOfficeURL() string {
	url := os.Getenv("OFFICE_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	return url
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
