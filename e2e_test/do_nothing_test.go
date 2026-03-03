//go:build e2e

package e2etest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	queueName := "emails.emp-1"
	email := Email{
		From:    Employee{ID: "ceo", Name: "The CEO", Role: "CEO"},
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
	consumed := false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		q, err := ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
		if err != nil {
			t.Fatalf("Failed to inspect queue: %v", err)
		}
		if q.Messages == 0 {
			consumed = true
			t.Log("Message consumed successfully")
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !consumed {
		t.Fatal("Timed out waiting for message to be consumed")
	}

	// Wait a bit for the agent to finish processing and logging.
	time.Sleep(5 * time.Second)

	// Query the employee's logs to verify the do_nothing action was chosen.
	resp, err := http.Get(fmt.Sprintf("%s/logs", getEmployeeURL(0)))
	if err != nil {
		t.Fatalf("Failed to fetch logs: %v", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read logs response: %v", err)
	}

	logs := string(body)

	var logEntries []LogEntry
	if err := json.Unmarshal([]byte(logs), &logEntries); err != nil {
		t.Fatalf("Failed to unmarshal logs: %v", err)
	}

	found := false
	for _, entry := range logEntries {
		if entry.Event == "do_nothing" {
			found = true
		}
	}

	t.Log("Found do_nothing action in logs")
	if !found {
		t.Fatal("Expected logs to contain do_nothing action")
	}
}
