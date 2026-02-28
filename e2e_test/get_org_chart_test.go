//go:build e2e

package e2etest

import (
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publishes a "get org chart" email and verifies the employee
// service consumes it (message gets acked off the queue).
//
// Requires: RabbitMQ + employee service running (e.g. via docker-compose.test.yml).
func TestGetOrgChart(t *testing.T) {
	conn, ch, err := getRabbitMqClient()
	if err != nil {
		t.Fatalf("Failed to get RabbitMQ client: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	queueName := "emails.emp-1"
	email := Email{
		From:    Employee{ID: "emp-1", Name: "Employee 1", Role: "Staff"},
		Subject: "Get Org Chart",
		Body:    "Please get the org chart of the company.",
	}
	body, _ := json.Marshal(email)

	err = ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Failed to publish email: %v", err)
	}

	t.Log("Published get org chart email, waiting for consumer to ack...")

	// Poll the queue until the message is consumed or timeout.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		q, err := ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
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
