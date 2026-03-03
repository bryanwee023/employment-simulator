package actions

import (
	"encoding/json"
	"errors"

	"github.com/bryanwee023/employment-simulator/employee/mail"
)

func SendEmailAction(mailer *mail.Mailer) Action {
	return Action{
		Description: "Send a short email to a co-worker",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"to":      {"type": "string", "description": "employee ID of the recipient"},
				"subject": {"type": "string"},
				"body":    {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (error, string) {
			var params struct {
				To      string `json:"to"`
				Subject string `json:"subject"`
				Body    string `json:"body"`
			}

			if err := json.Unmarshal(args, &params); err != nil {
				return nil, "unexpected json structure."
			}

			if err := mailer.Send(params.To, params.Subject, params.Body); err != nil {
				if errors.Is(err, mail.ErrQueueNotFound) {
					return nil, err.Error()
				}
				return err, ""
			}

			return nil, ""
		},
	}
}
