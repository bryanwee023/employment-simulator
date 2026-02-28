package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bryanwee023/employment-simulator/employee/actions"
	"github.com/bryanwee023/employment-simulator/employee/llm"
	amqp "github.com/rabbitmq/amqp091-go"
)

var self Employee

func main() {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	officeURL := os.Getenv("OFFICE_URL")
	if officeURL == "" {
		officeURL = "http://localhost:8080"
	}

	actions.OfficeURL = officeURL

	var err error
	self, err = fetchEmployee(officeURL)
	if err != nil {
		log.Fatalf("Failed to fetch employee from office: %v", err)
	}

	log.Printf("Registered as %s (%s)", self.Name, self.Role)

	model := llm.NewOllamaLLM(ollamaURL)
	agent := NewAgent(self, model, actions.Actions)

	if err := startEmailLoop(amqpURL, agent); err != nil {
		log.Fatalf("Failed to start email loop: %v", err)
	}

	http.HandleFunc("/me", meHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Employee service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// ============================
// Handlers
// ============================

func meHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(self)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok"}`)
}

// ============================
// Loops
// ============================

func startEmailLoop(url string, agent *Agent) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	queueName := "emails." + self.ID
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Printf("Listening for emails on queue %s", queueName)

	go func() {
		defer conn.Close()
		defer ch.Close()

		for msg := range msgs {
			var email Email
			if err := json.Unmarshal(msg.Body, &email); err != nil {
				log.Printf("Failed to parse email: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Received [%s: %s]", email.From.Name, email.Subject)
			if err := agent.Handle(email); err != nil {
				log.Printf("Failed to handle email: %v", err)
				msg.Nack(false, false)
				continue
			}

			msg.Ack(false)
		}

		log.Println("Email consumer channel closed")
	}()

	return nil
}
