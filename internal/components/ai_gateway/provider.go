package ai_gateway

import (
	"context"
	"encoding/json"
	"fmt"
)

type Provider interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ChatMessage struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

type ChatRequest struct {
	Model        string
	SystemPrompt string
	Prompt       string
	Messages     []ChatMessage
	MaxTokens    int
	Temperature  float64
	Tools        []ToolDefinition
}

type ChatResponse struct {
	Content      string
	Model        string
	FinishReason string
	ToolCalls    []ToolCall
	Usage        Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}

func NewProvider(name, apiKey, baseURL string) (Provider, error) {
	switch name {
	case "openai":
		return newOpenAIProvider(apiKey, baseURL), nil
	case "anthropic":
		return newAnthropicProvider(apiKey, baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", name)
	}
}
