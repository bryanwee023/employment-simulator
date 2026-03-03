package main

// Employee represents employee metadata
type Employee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Email represents an incoming email message
type Email struct {
	From    Employee `json:"from"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

// LLM is the interface for language model backends
type LLM interface {
	Chat(userPrompt string) (string, error)
	StartSession(systemPrompt string)
	ClearSession()
}
