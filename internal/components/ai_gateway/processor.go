package ai_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	err := service.RegisterProcessor(
		"ai_gateway", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	provider      Provider
	model         string
	promptStatic  string
	promptDynamic *service.InterpolatedString
	sysMsgStatic  string
	sysMsgDynamic *service.InterpolatedString
	argsMapping   *bloblang.Executor
	maxTokens     int
	temperature   float64
	mcpTools      bool
	mcpURL        string
	maxToolRounds int
	logger        *service.Logger

	mcpClient     *mcpclient.Client
	mcpOnce       sync.Once
	mcpInitErr    error
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	providerName, err := conf.FieldString(agfProvider)
	if err != nil {
		return nil, err
	}

	model, err := conf.FieldString(agfModel)
	if err != nil {
		return nil, err
	}

	apiKey, err := conf.FieldString(agfAPIKey)
	if err != nil {
		return nil, err
	}

	var baseURL string
	if conf.Contains(agfBaseURL) {
		baseURL, err = conf.FieldString(agfBaseURL)
		if err != nil {
			return nil, err
		}
	}

	unsafeDynamic, err := conf.FieldBool(agfUnsafeDynamicPrompt)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		model:  model,
		logger: mgr.Logger(),
	}

	if unsafeDynamic {
		if p.promptDynamic, err = conf.FieldInterpolatedString(agfPrompt); err != nil {
			return nil, err
		}
	} else {
		if p.promptStatic, err = conf.FieldString(agfPrompt); err != nil {
			return nil, err
		}
	}

	if conf.Contains(agfSystemPrompt) {
		if unsafeDynamic {
			if p.sysMsgDynamic, err = conf.FieldInterpolatedString(agfSystemPrompt); err != nil {
				return nil, err
			}
		} else {
			if p.sysMsgStatic, err = conf.FieldString(agfSystemPrompt); err != nil {
				return nil, err
			}
		}
	}

	if conf.Contains(agfArgsMapping) {
		if p.argsMapping, err = conf.FieldBloblang(agfArgsMapping); err != nil {
			return nil, err
		}
	}

	p.maxTokens, err = conf.FieldInt(agfMaxTokens)
	if err != nil {
		return nil, err
	}

	p.temperature, err = conf.FieldFloat(agfTemperature)
	if err != nil {
		return nil, err
	}

	p.mcpTools, err = conf.FieldBool(agfMCPTools)
	if err != nil {
		return nil, err
	}

	p.mcpURL, err = conf.FieldString(agfMCPURL)
	if err != nil {
		return nil, err
	}

	p.maxToolRounds, err = conf.FieldInt(agfMaxToolRounds)
	if err != nil {
		return nil, err
	}

	p.provider, err = NewProvider(providerName, apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Processor) initMCPClient() (*mcpclient.Client, error) {
	p.mcpOnce.Do(func() {
		p.logger.Debugf("Connecting to MCP server at %s", p.mcpURL)
		c, err := mcpclient.NewStreamableHttpClient(p.mcpURL)
		if err != nil {
			p.mcpInitErr = fmt.Errorf("failed to create MCP client: %w", err)
			return
		}

		ctx := context.Background()
		if _, err = c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
			p.mcpInitErr = fmt.Errorf("failed to initialize MCP session: %w", err)
			return
		}

		p.mcpClient = c
	})
	return p.mcpClient, p.mcpInitErr
}

func (p *Processor) fetchMCPTools(ctx context.Context) ([]ToolDefinition, error) {
	client, err := p.initMCPClient()
	if err != nil {
		return nil, err
	}

	result, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}

	tools := make([]ToolDefinition, 0, len(result.Tools))
	for _, t := range result.Tools {
		schemaBytes, err := json.Marshal(t.InputSchema)
		if err != nil {
			p.logger.Warnf("Failed to marshal input schema for tool %s: %v", t.Name, err)
			continue
		}
		tools = append(tools, ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schemaBytes,
		})
	}

	return tools, nil
}

func (p *Processor) executeMCPTool(ctx context.Context, name string, arguments json.RawMessage) (string, error) {
	client, err := p.initMCPClient()
	if err != nil {
		return "", err
	}

	var args map[string]any
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	if result.IsError {
		for _, c := range result.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				return "", fmt.Errorf("MCP tool error: %s", tc.Text)
			}
		}
		return "", fmt.Errorf("MCP tool returned an error")
	}

	var textResult string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			textResult = tc.Text
			break
		}
	}

	return textResult, nil
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	var promptStr string
	var err error
	if p.promptDynamic != nil {
		promptStr, err = p.promptDynamic.TryString(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate prompt: %w", err)
		}
	} else {
		promptStr = p.promptStatic
	}

	var systemPromptStr string
	if p.sysMsgDynamic != nil {
		systemPromptStr, err = p.sysMsgDynamic.TryString(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate system_prompt: %w", err)
		}
	} else {
		systemPromptStr = p.sysMsgStatic
	}

	if p.argsMapping != nil {
		batch := service.MessageBatch{msg}
		executor := batch.BloblangExecutor(p.argsMapping)

		resMsg, err := executor.Query(0)
		if err != nil {
			return nil, fmt.Errorf("failed to execute args_mapping: %w", err)
		}

		iargs, err := resMsg.AsStructured()
		if err != nil {
			return nil, fmt.Errorf("args_mapping result is not structured: %w", err)
		}

		args, ok := iargs.([]any)
		if !ok {
			return nil, fmt.Errorf("args_mapping must evaluate to an array, got %T", iargs)
		}

		promptStr = fmt.Sprintf(promptStr, args...)
	}

	var tools []ToolDefinition
	if p.mcpTools {
		tools, err = p.fetchMCPTools(ctx)
		if err != nil {
			p.logger.Warnf("Failed to fetch MCP tools, proceeding without tools: %v", err)
			tools = nil
		} else if len(tools) > 0 {
			p.logger.Debugf("Discovered %d MCP tools", len(tools))
		}
	}

	chatReq := &ChatRequest{
		Model:        p.model,
		Prompt:       promptStr,
		SystemPrompt: systemPromptStr,
		MaxTokens:    p.maxTokens,
		Temperature:  p.temperature,
		Tools:        tools,
	}

	p.logger.Debugf("AI gateway request: model=%s, prompt_len=%d, system_prompt_len=%d, max_tokens=%d, temperature=%f, tools=%d",
		chatReq.Model, len(chatReq.Prompt), len(chatReq.SystemPrompt), chatReq.MaxTokens, chatReq.Temperature, len(tools))
	p.logger.Tracef("AI gateway prompt: %s", chatReq.Prompt)
	p.logger.Tracef("AI gateway system_prompt: %s", chatReq.SystemPrompt)

	chatResp, err := p.provider.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("AI chat request failed: %w", err)
	}

	// Tool calling loop
	for round := 0; round < p.maxToolRounds && len(chatResp.ToolCalls) > 0; round++ {
		p.logger.Debugf("Tool calling round %d: %d tool calls", round+1, len(chatResp.ToolCalls))

		messages := []ChatMessage{
			{Role: "user", Content: promptStr},
			{Role: "assistant", Content: chatResp.Content, ToolCalls: chatResp.ToolCalls},
		}

		if round > 0 && len(chatReq.Messages) > 0 {
			messages = chatReq.Messages
			messages = append(messages, ChatMessage{
				Role:      "assistant",
				Content:   chatResp.Content,
				ToolCalls: chatResp.ToolCalls,
			})
		}

		for _, tc := range chatResp.ToolCalls {
			p.logger.Debugf("Executing MCP tool: %s", tc.Name)
			p.logger.Tracef("Tool arguments: %s", string(tc.Arguments))

			result, err := p.executeMCPTool(ctx, tc.Name, tc.Arguments)
			if err != nil {
				p.logger.Warnf("MCP tool %s failed: %v", tc.Name, err)
				result = fmt.Sprintf("Error: %v", err)
			}

			p.logger.Tracef("Tool %s result: %s", tc.Name, result)

			messages = append(messages, ChatMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		chatReq = &ChatRequest{
			Model:        p.model,
			SystemPrompt: systemPromptStr,
			Messages:     messages,
			MaxTokens:    p.maxTokens,
			Temperature:  p.temperature,
			Tools:        tools,
		}

		chatResp, err = p.provider.Chat(ctx, chatReq)
		if err != nil {
			return nil, fmt.Errorf("AI chat request failed during tool calling round %d: %w", round+1, err)
		}
	}

	p.logger.Debugf("AI gateway response: model=%s, finish_reason=%s, content_len=%d, input_tokens=%d, output_tokens=%d",
		chatResp.Model, chatResp.FinishReason, len(chatResp.Content), chatResp.Usage.InputTokens, chatResp.Usage.OutputTokens)
	if chatResp.Content == "" {
		p.logger.Warnf("AI gateway returned empty content: finish_reason=%s, model=%s", chatResp.FinishReason, chatResp.Model)
	}
	p.logger.Tracef("AI gateway response content: %s", chatResp.Content)

	outMsg := msg.Copy()
	outMsg.SetStructured(map[string]any{
		"content":       chatResp.Content,
		"model":         chatResp.Model,
		"finish_reason": chatResp.FinishReason,
		"usage": map[string]any{
			"input_tokens":  chatResp.Usage.InputTokens,
			"output_tokens": chatResp.Usage.OutputTokens,
		},
	})

	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	if p.mcpClient != nil {
		return p.mcpClient.Close()
	}
	return nil
}
