package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bryanwee023/employment-simulator/employee/actions"
)

type Agent struct {
	employee Employee
	llm      LLM
	logger   *Logger
	actions  map[string]actions.Action
}

func NewAgent(employee Employee, llm LLM, actions map[string]actions.Action, logger *Logger) *Agent {
	return &Agent{employee: employee, llm: llm, actions: actions, logger: logger}
}

func (a *Agent) Handle(email Email) error {
	logger = a.logger

	logger.Log("Received email", email)

	systemPrompt, userPrompt := a.buildPrompts(email)

	response, err := a.llm.Chat(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("llm call failed: %w", err)
	}

	logger.Log("LLM response", response)

	err, hint := a.execute(response)
	if err != nil {
		return fmt.Errorf("failed to execute action: %w", err)
	}

	if hint != "" {
		logger.Log("Hint from execution", hint)
		// TODO: Relay hint to llm call and try again.
	}

	return nil
}

func (a *Agent) buildPrompts(email Email) (string, string) {
	actionsJSON, _ := json.Marshal(a.actions)

	systemPrompt := fmt.Sprintf(
		"You are a %s (%s).\n"+
			"Given the received email below, decide which action to take.\n"+
			"Available actions: %s\n"+
			"Respond with JSON: {\"action\": \"<action_name>\", \"arguments\": {...}}",
		a.employee.Name,
		a.employee.Role,
		string(actionsJSON),
	)

	userPrompt := fmt.Sprintf(
		"From: %s (%s)\nSubject: %s\n\n%s",
		email.From.Name, email.From.Role, email.Subject, email.Body,
	)

	return systemPrompt, userPrompt
}

func (a *Agent) execute(response string) (error, string) {
	var choice struct {
		ActionName string          `json:"action"`
		Arguments  json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(response), &choice); err != nil {
		return nil, "invalid json structure"
	}

	actionName := strings.ToLower(choice.ActionName)
	action, ok := a.actions[actionName]
	if !ok {
		return nil, "unknown action"
	}

	a.logger.Log(actionName, choice.Arguments)
	return action.Execute(choice.Arguments)
}
