package actions

import (
	"encoding/json"
	"log"
)

var ActionSendEmail = Action{
	Description: "Send a short email to a co-worker",
	Parameters: json.RawMessage(`{
		"type": "object",
		"properties": {
			"to":      {"type": "string"},
			"subject": {"type": "string"},
			"body":    {"type": "string"}
		}
	}`),
	Execute: sendEmail,
}

func sendEmail(args json.RawMessage) (error, string) {
	var params struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, "invalid json structure."
	}
	log.Printf("Sending email to %s: %s", params.To, params.Subject)
	// TODO: Implement actual sending via RabbitMQ
	return nil, ""
}
