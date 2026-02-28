//go:build e2e

package e2etest

import (
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publishes a "do nothing" email and verifies the employee
// service consumes it (message gets acked off the queue).
//
// Requires: RabbitMQ + employee service running (e.g. via docker-compose.test.yml).
func TestDoNothing(t *testing.T) {
	conn, ch, err := getRabbitMqClient()
	if err != nil {
		t.Fatalf("Failed to get RabbitMQ client: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	queueName := "emails.kevin1"
	email := Email{
		From:    Employee{ID: "kevin1", Name: "Kevin Stone", Role: "Alignment Manager"},
		Subject: "No action needed",
		Body:    "This is a useless email. Do nothing.",
	}
	body, _ := json.Marshal(email)

	err = ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Failed to publish email: %v", err)
	}

	t.Log("Published do-nothing email, waiting for consumer to ack...")

	// Poll the queue until the message is consumed or timeout.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		q, err := ch.QueueInspect(queueName)
		if err != nil {
			t.Fatalf("Failed to inspect queue: %v", err)
		}
		if q.Messages == 0 {
			t.Log("Message consumed successfully")
			return
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatal("Timed out waiting for message to be consumed")
}
