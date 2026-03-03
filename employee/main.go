package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bryanwee023/employment-simulator/employee/actions"
	"github.com/bryanwee023/employment-simulator/employee/llm"
	"github.com/bryanwee023/employment-simulator/employee/mail"
)

var self Employee
var logger *Logger

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

	mailer, err := mail.NewMailer(amqpURL, self.ID, self.Name, self.Role)
	if err != nil {
		log.Fatalf("Failed to initialize mailer: %v", err)
	}

	logger = NewLogger()

	model := llm.NewOllamaLLM(ollamaURL)
	agent := NewAgent(self, model, actions.NewActions(mailer), logger)

	if err := startEmailLoop(mailer, agent); err != nil {
		log.Fatalf("Failed to start email loop: %v", err)
	}

	http.HandleFunc("/me", meHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/logs", logsHandler)

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

func logsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logger.Entries())
}

// ============================
// Loops
// ============================

func startEmailLoop(mailer *mail.Mailer, agent *Agent) error {
	msgs, err := mailer.Receive(self.ID)
	if err != nil {
		return fmt.Errorf("failed to start receiving emails: %w", err)
	}

	log.Printf("Listening for emails on queue emails.%s", self.ID)

	go func() {
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
