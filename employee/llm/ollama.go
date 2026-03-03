package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const ollamaModel = "llama3.2:1b"

type OllamaLLM struct {
	url     string
	history []ollamaMessage
}

func NewOllamaLLM(url string) *OllamaLLM {
	return &OllamaLLM{
		url:     url,
		history: []ollamaMessage{},
	}
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   string          `json:"format"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

func (o *OllamaLLM) StartSession(systemPrompt string) {
	o.history = []ollamaMessage{
		{Role: "system", Content: systemPrompt},
	}
}

func (o *OllamaLLM) Chat(userPrompt string) (string, error) {
	o.history = append(o.history, ollamaMessage{Role: "user", Content: userPrompt})

	// Copy for the request so we can release the lock before the HTTP call
	messages := make([]ollamaMessage, len(o.history))
	copy(messages, o.history)

	reqBody := ollamaRequest{
		Model:    ollamaModel,
		Messages: messages,
		Stream:   false,
		Format:   "json",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(o.url+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Append assistant response to session history
	o.history = append(o.history, ollamaMessage{
		Role:    "assistant",
		Content: ollamaResp.Message.Content,
	})

	return ollamaResp.Message.Content, nil
}

func (o *OllamaLLM) ClearSession() {
	o.history = []ollamaMessage{}
}
