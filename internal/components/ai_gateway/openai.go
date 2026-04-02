package ai_gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type openaiProvider struct {
	client openai.Client
}

func newOpenAIProvider(apiKey, baseURL string) *openaiProvider {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &openaiProvider{
		client: openai.NewClient(opts...),
	}
}

func (o *openaiProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+2)

	if len(req.Messages) > 0 {
		if req.SystemPrompt != "" {
			messages = append(messages, openai.SystemMessage(req.SystemPrompt))
		}
		for _, msg := range req.Messages {
			switch msg.Role {
			case "user":
				messages = append(messages, openai.UserMessage(msg.Content))
			case "assistant":
				if len(msg.ToolCalls) > 0 {
					toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(msg.ToolCalls))
					for _, tc := range msg.ToolCalls {
						toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
							OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
								ID: tc.ID,
								Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
									Name:      tc.Name,
									Arguments: string(tc.Arguments),
								},
							},
						})
					}
					messages = append(messages, openai.ChatCompletionMessageParamUnion{
						OfAssistant: &openai.ChatCompletionAssistantMessageParam{
							Content:   openai.ChatCompletionAssistantMessageParamContentUnion{OfString: param.NewOpt(msg.Content)},
							ToolCalls: toolCalls,
						},
					})
				} else {
					messages = append(messages, openai.AssistantMessage(msg.Content))
				}
			case "tool":
				messages = append(messages, openai.ToolMessage(msg.Content, msg.ToolCallID))
			}
		}
	} else {
		if req.SystemPrompt != "" {
			messages = append(messages, openai.SystemMessage(req.SystemPrompt))
		}
		messages = append(messages, openai.UserMessage(req.Prompt))
	}

	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: messages,
	}
	if req.MaxTokens > 0 {
		params.MaxCompletionTokens = param.NewOpt(int64(req.MaxTokens))
	}
	if req.Temperature >= 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}

	if len(req.Tools) > 0 {
		tools := make([]openai.ChatCompletionToolUnionParam, 0, len(req.Tools))
		for _, t := range req.Tools {
			var funcParams shared.FunctionParameters
			if len(t.InputSchema) > 0 {
				if err := json.Unmarshal(t.InputSchema, &funcParams); err != nil {
					return nil, fmt.Errorf("failed to parse tool input schema for %s: %w", t.Name, err)
				}
			}
			tools = append(tools, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
				Name:        t.Name,
				Description: param.NewOpt(t.Description),
				Parameters:  funcParams,
			}))
		}
		params.Tools = tools
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("OpenAI chat request failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI returned no choices")
	}

	choice := completion.Choices[0]
	resp := &ChatResponse{
		Content:      choice.Message.Content,
		Model:        completion.Model,
		FinishReason: choice.FinishReason,
		Usage: Usage{
			InputTokens:  int(completion.Usage.PromptTokens),
			OutputTokens: int(completion.Usage.CompletionTokens),
		},
	}

	for _, tc := range choice.Message.ToolCalls {
		if tc.Type == "function" {
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(tc.Function.Arguments),
			})
		}
	}

	return resp, nil
}
