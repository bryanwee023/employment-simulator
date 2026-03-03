package actions

import (
	"encoding/json"
	"fmt"

	"github.com/bryanwee023/employment-simulator/employee/mail"
)

// ExecuteFn executes an action with the given JSON arguments.
// Returns an error if execution fails, or an optional hint string
// for the LLM to retry with corrected input.
type ExecuteFn func(args json.RawMessage) (error, string)

type Action struct {
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Execute     ExecuteFn       `json:"-"`
}

func (a Action) String() string {
	return fmt.Sprintf("%s", a.Description)
}

var OfficeURL string

func NewActions(mailer *mail.Mailer) map[string]Action {
	return map[string]Action{
		"send_email":    SendEmailAction(mailer),
		"do_nothing":    ActionDoNothing,
		"get_org_chart": ActionGetOrgChart,
	}
}
