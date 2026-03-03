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

// Publishes an email to emp-1 instructing them to send an email
// to emp-2, then verifies:
//  1. emp-1 consumed the original message
//  2. emp-1's logs show the send_email action
//  3. emp-2's queue received the forwarded email
//
// Requires: RabbitMQ + office + employee1 + employee2 running.
func TestSendEmail(t *testing.T) {
	conn, ch, err := getRabbitMqClient()
	if err != nil {
		t.Fatalf("Failed to get RabbitMQ client: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	// Send an email to emp-1 asking them to email emp-2.
	emp1Queue := "emails.emp-1"
	email := Email{
		From:    Employee{ID: "ceo", Name: "The CEO", Role: "CEO"},
		Subject: "Forward this to Employee 2",
		Body:    "Please send an email to emp-2 with subject 'Hello' and body 'Do nothing'.",
	}
	body, _ := json.Marshal(email)

	err = ch.Publish("", emp1Queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		t.Fatalf("Failed to publish email: %v", err)
	}

	t.Log("Published email to emp-1, waiting for consumer to ack...")

	// Wait for emp-1 to consume the message.
	consumed := false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		q, err := ch.QueueDeclarePassive(emp1Queue, true, false, false, false, nil)
		if err != nil {
			t.Fatalf("Failed to inspect queue: %v", err)
		}
		if q.Messages == 0 {
			consumed = true
			t.Log("emp-1 consumed the message")
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !consumed {
		t.Fatal("Timed out waiting for emp-1 to consume the message")
	}

	// Wait for the agent to finish processing and logging.
	time.Sleep(10 * time.Second)

	// Verify emp-1's logs contain the send_email action.
	resp, err := http.Get(fmt.Sprintf("%s/logs", getEmployeeURL(0)))
	if err != nil {
		t.Fatalf("Failed to fetch emp-1's logs: %v", err)
	}
	defer resp.Body.Close()

	logBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read logs response: %v", err)
	}

	var logEntries []LogEntry
	if err := json.Unmarshal(logBody, &logEntries); err != nil {
		t.Fatalf("Failed to unmarshal logs: %v", err)
	}

	found := false
	for _, entry := range logEntries {
		if entry.Event == "send_email" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected emp-1's logs to contain send_email action")
	}
	t.Log("Found send_email action in emp-1's logs")

	// Wait for emp-2 to finish processing and logging.
	time.Sleep(10 * time.Second)

	// Verify emp-2's logs contain the do_nothing action.
	resp, err = http.Get(fmt.Sprintf("%s/logs", getEmployeeURL(1)))
	if err != nil {
		t.Fatalf("Failed to fetch emp-2's logs: %v", err)
	}
	defer resp.Body.Close()

	logBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read emp-2 logs response: %v", err)
	}

	var emp2LogEntries []LogEntry
	if err := json.Unmarshal(logBody, &emp2LogEntries); err != nil {
		t.Fatalf("Failed to unmarshal emp-2 logs: %v", err)
	}

	found = false
	for _, entry := range emp2LogEntries {
		if entry.Event == "do_nothing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected emp-2's logs to contain do_nothing action")
	}
	t.Log("Found do_nothing action in emp-2's logs")
}
