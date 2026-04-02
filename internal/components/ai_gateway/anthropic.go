package ai_gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

type anthropicProvider struct {
	client anthropic.Client
}

func newAnthropicProvider(apiKey, baseURL string) *anthropicProvider {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &anthropicProvider{
		client: anthropic.NewClient(opts...),
	}
}

func (a *anthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]anthropic.MessageParam, 0, len(req.Messages)+1)

	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			switch msg.Role {
			case "user":
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
			case "assistant":
				if len(msg.ToolCalls) > 0 {
					blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.ToolCalls)+1)
					if msg.Content != "" {
						blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
					}
					for _, tc := range msg.ToolCalls {
						var input any
						if err := json.Unmarshal(tc.Arguments, &input); err != nil {
							input = map[string]any{}
						}
						blocks = append(blocks, anthropic.ContentBlockParamUnion{
							OfToolUse: &anthropic.ToolUseBlockParam{
								ID:    tc.ID,
								Name:  tc.Name,
								Input: input,
							},
						})
					}
					messages = append(messages, anthropic.NewAssistantMessage(blocks...))
				} else {
					messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
				}
			case "tool":
				isError := false
				messages = append(messages, anthropic.NewUserMessage(
					anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, isError),
				))
			}
		}
	} else {
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)))
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}
	if params.MaxTokens <= 0 {
		params.MaxTokens = 1024
	}
	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.SystemPrompt}}
	}
	if req.Temperature >= 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}

	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, 0, len(req.Tools))
		for _, t := range req.Tools {
			var props any
			var required []string
			var schemaMap map[string]any
			if err := json.Unmarshal(t.InputSchema, &schemaMap); err == nil {
				props = schemaMap["properties"]
				if req, ok := schemaMap["required"].([]any); ok {
					for _, r := range req {
						if s, ok := r.(string); ok {
							required = append(required, s)
						}
					}
				}
			}
			tools = append(tools, anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        t.Name,
					Description: param.NewOpt(t.Description),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: props,
						Required:   required,
					},
				},
			})
		}
		params.Tools = tools
	}

	message, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Anthropic chat request failed: %w", err)
	}

	resp := &ChatResponse{
		Model:        string(message.Model),
		FinishReason: string(message.StopReason),
		Usage: Usage{
			InputTokens:  int(message.Usage.InputTokens),
			OutputTokens: int(message.Usage.OutputTokens),
		},
	}

	for _, block := range message.Content {
		switch block.Type {
		case "text":
			resp.Content = block.Text
		case "tool_use":
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: block.Input,
			})
		}
	}

	return resp, nil
}
