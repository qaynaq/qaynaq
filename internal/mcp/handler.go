package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

type RequestForwarder interface {
	ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error)
}

type toolConfig struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	InputSchema json.RawMessage `yaml:"input_schema"`
}

type MCPHandler struct {
	mcpServer   *server.MCPServer
	httpHandler *server.StreamableHTTPServer
	flowRepo    persistence.FlowRepository
	forwarder   RequestForwarder
	mu          sync.RWMutex
	// tool name -> flow ID used for forwarding via /ingest/{flowID}
	toolFlowMap map[string]int64
}

func NewMCPHandler(flowRepo persistence.FlowRepository, forwarder RequestForwarder, version string) *MCPHandler {
	mcpServer := server.NewMCPServer(
		"qaynaq",
		version,
		server.WithToolCapabilities(true),
	)

	httpHandler := server.NewStreamableHTTPServer(mcpServer)

	h := &MCPHandler{
		mcpServer:   mcpServer,
		httpHandler: httpHandler,
		flowRepo:    flowRepo,
		forwarder:   forwarder,
		toolFlowMap: make(map[string]int64),
	}

	h.SyncTools()
	return h
}

func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.httpHandler.ServeHTTP(w, r)
}

func (h *MCPHandler) SyncTools() {
	flows, err := h.flowRepo.ListAllByStatuses(persistence.FlowStatusActive)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list flows for MCP tool sync")
		return
	}

	newToolMap := make(map[string]int64)
	var newTools []server.ServerTool

	for _, flow := range flows {
		if flow.InputComponent != "mcp_tool" {
			continue
		}

		cfg, err := parseToolConfig(flow.InputConfig)
		if err != nil {
			log.Warn().Err(err).Int64("flow_id", flow.ID).Msg("Failed to parse MCP tool config")
			continue
		}

		if cfg.Name == "" {
			log.Warn().Int64("flow_id", flow.ID).Msg("MCP flow missing tool name")
			continue
		}

		if _, exists := newToolMap[cfg.Name]; exists {
			log.Warn().Str("tool", cfg.Name).Int64("flow_id", flow.ID).Msg("Duplicate MCP tool name, skipping")
			continue
		}

		flowID := flow.ID
		if flow.ParentID != nil {
			flowID = *flow.ParentID
		}
		newToolMap[cfg.Name] = flowID

		tool := mcp.NewToolWithRawSchema(cfg.Name, cfg.Description, cfg.InputSchema)
		newTools = append(newTools, server.ServerTool{
			Tool:    tool,
			Handler: h.createToolHandler(flowID),
		})
	}

	h.mu.Lock()
	h.toolFlowMap = newToolMap
	h.mu.Unlock()

	h.mcpServer.SetTools(newTools...)

	log.Debug().Int("tool_count", len(newTools)).Msg("MCP tools synced")
}

func (h *MCPHandler) createToolHandler(flowID int64) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		payload, err := json.Marshal(args)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal arguments: %v", err)), nil
		}

		path := fmt.Sprintf("/ingest/%d/", flowID)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(payload))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create request: %v", err)), nil
		}
		req.Header.Set("Content-Type", "application/json")

		statusCode, response, err := h.forwarder.ForwardRequestToWorker(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("tool execution failed: %v", err)), nil
		}

		if statusCode >= 400 {
			return mcp.NewToolResultError(fmt.Sprintf("tool returned status %d: %s", statusCode, string(response))), nil
		}

		return mcp.NewToolResultText(string(response)), nil
	}
}

func parseToolConfig(inputConfig []byte) (*toolConfig, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(inputConfig, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input config: %w", err)
	}

	cfg := &toolConfig{}

	if name, ok := raw["name"].(string); ok {
		cfg.Name = name
	}
	if desc, ok := raw["description"].(string); ok {
		cfg.Description = desc
	}

	if schema, ok := raw["input_schema"]; ok {
		jsonSchema, err := propertyListToJSONSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input_schema: %w", err)
		}
		cfg.InputSchema = jsonSchema
	} else {
		cfg.InputSchema = json.RawMessage(`{"type":"object","properties":{}}`)
	}

	return cfg, nil
}

func propertyListToJSONSchema(schema any) (json.RawMessage, error) {
	propList, ok := schema.([]any)
	if !ok {
		schemaJSON, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input_schema to JSON: %w", err)
		}
		return schemaJSON, nil
	}

	properties := make(map[string]any)
	var required []string

	for _, item := range propList {
		prop, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := prop["name"].(string)
		if name == "" {
			continue
		}

		propSchema := map[string]any{}
		if t, ok := prop["type"].(string); ok {
			propSchema["type"] = t
		}
		if desc, ok := prop["description"].(string); ok && desc != "" {
			propSchema["description"] = desc
		}

		properties[name] = propSchema

		if req, ok := prop["required"].(bool); ok && req {
			required = append(required, name)
		}
	}

	jsonSchemaObj := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		jsonSchemaObj["required"] = required
	}

	return json.Marshal(jsonSchemaObj)
}
