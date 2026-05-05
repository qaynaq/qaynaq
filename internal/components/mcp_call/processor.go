package mcp_call

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"

	mcphelper "github.com/qaynaq/qaynaq/internal/mcp"
)

func init() {
	err := service.RegisterProcessor(
		"mcp_call", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	serverURL   string
	tool        string
	authHeader  string
	authValue   string
	argsMapping *bloblang.Executor
	logger      *service.Logger

	mcpClient *mcpclient.Client
	mcpMu     sync.Mutex
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	serverURL, err := conf.FieldString(fieldServerURL)
	if err != nil {
		return nil, err
	}

	tool, err := conf.FieldString(fieldTool)
	if err != nil {
		return nil, err
	}

	authHeader, err := conf.FieldString(fieldAuthHeader)
	if err != nil {
		return nil, err
	}

	authValue, err := conf.FieldString(fieldAuthValue)
	if err != nil {
		return nil, err
	}

	var argsMapping *bloblang.Executor
	if conf.Contains(fieldArgsMapping) {
		argsMapping, err = conf.FieldBloblang(fieldArgsMapping)
		if err != nil {
			return nil, err
		}
	}

	return &Processor{
		serverURL:   serverURL,
		tool:        tool,
		authHeader:  authHeader,
		authValue:   authValue,
		argsMapping: argsMapping,
		logger:      mgr.Logger(),
	}, nil
}

func (p *Processor) initClient() (*mcpclient.Client, error) {
	p.mcpMu.Lock()
	defer p.mcpMu.Unlock()

	if p.mcpClient != nil {
		return p.mcpClient, nil
	}

	p.logger.Debugf("Connecting to MCP server at %s", p.serverURL)
	headers := mcphelper.BuildAuthHeaders(p.authHeader, p.authValue)
	client, err := mcphelper.ConnectMCPClient(context.Background(), p.serverURL, headers, nil)
	if err != nil {
		return nil, err
	}

	p.mcpClient = client
	return client, nil
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	start := time.Now()

	client, err := p.initClient()
	if err != nil {
		return nil, fmt.Errorf("mcp_call: connection failed: %w", err)
	}

	var args map[string]any
	if p.argsMapping != nil {
		res, err := msg.BloblangQuery(p.argsMapping)
		if err != nil {
			return nil, fmt.Errorf("mcp_call: args_mapping failed: %w", err)
		}
		structured, err := res.AsStructured()
		if err != nil {
			return nil, fmt.Errorf("mcp_call: args_mapping result is not structured: %w", err)
		}
		mapped, ok := structured.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("mcp_call: args_mapping must produce a JSON object, got %T", structured)
		}
		args = mapped
	} else {
		structured, err := msg.AsStructured()
		if err != nil {
			return nil, fmt.Errorf("mcp_call: failed to read message as structured: %w", err)
		}
		mapped, ok := structured.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("mcp_call: message must be a JSON object when args_mapping is not set, got %T", structured)
		}
		args = mapped
	}

	result, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      p.tool,
			Arguments: args,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("mcp_call: tool call failed: %w", err)
	}

	latency := time.Since(start)

	if result.IsError {
		for _, c := range result.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				return nil, fmt.Errorf("mcp_call: tool error: %s", tc.Text)
			}
		}
		return nil, fmt.Errorf("mcp_call: tool returned an error")
	}

	var textResult string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			textResult = tc.Text
			break
		}
	}

	outMsg := msg.Copy()

	var parsed any
	if err := json.Unmarshal([]byte(textResult), &parsed); err == nil {
		outMsg.SetStructured(parsed)
	} else {
		outMsg.SetBytes([]byte(textResult))
	}

	outMsg.MetaSet("mcp_server", p.serverURL)
	outMsg.MetaSet("mcp_tool", p.tool)
	outMsg.MetaSet("mcp_latency_ms", fmt.Sprintf("%d", latency.Milliseconds()))

	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	p.mcpMu.Lock()
	defer p.mcpMu.Unlock()

	if p.mcpClient != nil {
		p.mcpClient = nil
	}
	return nil
}
