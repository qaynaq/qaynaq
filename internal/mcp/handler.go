package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/qaynaq/qaynaq/internal/connection"
	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/vault"
)

type RequestForwarder interface {
	ForwardRequestToWorker(ctx context.Context, r *http.Request) (int32, []byte, error)
}

type toolConfig struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	InputSchema json.RawMessage `yaml:"input_schema"`
}

type upstreamServer struct {
	client         *mcpclient.Client
	serverID       int64
	name           string
	url            string
	authHeader     string
	authValue      string
	connectionName string
	tools          []mcp.Tool
	failureCount   int
	// failedAt is set when the breaker trips. Past upstreamCircuitCooldown the
	// next sync attempt is allowed through.
	failedAt time.Time
}

const (
	maxUpstreamFailures      = 3
	upstreamSyncTimeout      = 10 * time.Second
	upstreamCircuitCooldown  = 5 * time.Minute
)

type MCPHandler struct {
	mcpServer   *server.MCPServer
	httpHandler *server.StreamableHTTPServer
	flowRepo    persistence.FlowRepository
	serverRepo  persistence.MCPServerRepository
	forwarder   RequestForwarder
	aesgcm      *vault.AESGCM
	connManager *connection.Manager
	mu sync.RWMutex
	// tool name -> flow ID, used for forwarding via /ingest/{flowID}
	toolFlowMap   map[string]int64
	upstreams     map[string]*upstreamServer
	lastToolsHash string
}

func NewMCPHandler(flowRepo persistence.FlowRepository, serverRepo persistence.MCPServerRepository, forwarder RequestForwarder, aesgcm *vault.AESGCM, connManager *connection.Manager, version string) *MCPHandler {
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
		serverRepo:  serverRepo,
		forwarder:   forwarder,
		aesgcm:      aesgcm,
		connManager: connManager,
		toolFlowMap: make(map[string]int64),
		upstreams:   make(map[string]*upstreamServer),
	}

	if connManager != nil {
		connManager.OnTokenRefreshed(h.resetUpstreamsForConnection)
	}

	h.SyncTools()
	return h
}

func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.httpHandler.ServeHTTP(w, r)
}

func (h *MCPHandler) SyncTools() {
	nativeTools, nativeToolMap := h.syncNativeTools()
	upstreamTools := h.syncUpstreamServers(nativeToolMap)
	allTools := append(nativeTools, upstreamTools...)

	// Skip SetTools when nothing changed - it would needlessly notify clients.
	hash := computeToolsHash(allTools)
	if hash == h.lastToolsHash {
		return
	}

	h.mu.Lock()
	h.lastToolsHash = hash
	h.mu.Unlock()

	h.mcpServer.SetTools(allTools...)

	log.Debug().
		Int("native_count", len(nativeTools)).
		Int("upstream_count", len(upstreamTools)).
		Int("total_count", len(allTools)).
		Msg("MCP tools synced")
}

func (h *MCPHandler) syncNativeTools() ([]server.ServerTool, map[string]int64) {
	flows, err := h.flowRepo.ListAllByStatuses(persistence.FlowStatusActive)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list flows for MCP tool sync")
		return nil, nil
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
			Handler: h.createNativeToolHandler(flowID),
		})
	}

	h.mu.Lock()
	h.toolFlowMap = newToolMap
	h.mu.Unlock()

	return newTools, newToolMap
}

func (h *MCPHandler) syncUpstreamServers(nativeToolNames map[string]int64) []server.ServerTool {
	if h.serverRepo == nil {
		return nil
	}

	servers, err := h.serverRepo.ListByStatus("active")
	if err != nil {
		log.Error().Err(err).Msg("Failed to list MCP servers for upstream sync")
		return nil
	}

	if len(servers) == 0 {
		h.mu.Lock()
		for name, us := range h.upstreams {
			if us.client != nil {
				us.client = nil
			}
			delete(h.upstreams, name)
		}
		h.mu.Unlock()
		return nil
	}

	activeNames := make(map[string]bool, len(servers))
	for _, s := range servers {
		activeNames[s.Name] = true
	}

	// Drop upstream entries for servers that have been removed from the DB.
	h.mu.Lock()
	for name, us := range h.upstreams {
		if !activeNames[name] {
			if us.client != nil {
				us.client = nil
			}
			delete(h.upstreams, name)
		}
	}
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), upstreamSyncTimeout)
	defer cancel()

	type serverResult struct {
		name  string
		tools []server.ServerTool
	}

	var resultMu sync.Mutex
	var results []serverResult

	g, gCtx := errgroup.WithContext(ctx)

	for _, srv := range servers {
		g.Go(func() error {
			tools := h.syncOneUpstream(gCtx, &srv, nativeToolNames)
			if len(tools) > 0 {
				resultMu.Lock()
				results = append(results, serverResult{name: srv.Name, tools: tools})
				resultMu.Unlock()
			}
			// Never fail the group - one bad server shouldn't block the rest.
			return nil
		})
	}

	_ = g.Wait()

	var allUpstreamTools []server.ServerTool
	for _, r := range results {
		allUpstreamTools = append(allUpstreamTools, r.tools...)
	}

	return allUpstreamTools
}

func (h *MCPHandler) syncOneUpstream(ctx context.Context, srv *persistence.MCPServer, nativeToolNames map[string]int64) []server.ServerTool {
	h.mu.RLock()
	existing := h.upstreams[srv.Name]
	h.mu.RUnlock()

	// Circuit breaker: while inside the cooldown window we skip; past it we
	// let one attempt through. Success clears the counter, failure restarts
	// the cooldown.
	if existing != nil && existing.failureCount >= maxUpstreamFailures {
		if !existing.failedAt.IsZero() && time.Since(existing.failedAt) < upstreamCircuitCooldown {
			log.Debug().Str("server", srv.Name).Msg("Upstream server circuit-broken, skipping")
			return nil
		}
		log.Debug().Str("server", srv.Name).Msg("Upstream server cooldown elapsed, allowing retry")
	}

	needReconnect := existing == nil || existing.url != srv.URL

	authHeader, authValue := h.resolveAuth(ctx, srv)

	if existing != nil && existing.authValue != authValue {
		needReconnect = true
	}

	var client *mcpclient.Client

	if needReconnect {
		if existing != nil && existing.client != nil {
			existing.client = nil
		}

		headers := BuildAuthHeaders(authHeader, authValue)
		newClient, err := ConnectMCPClient(ctx, srv.URL, headers)
		if err != nil {
			log.Warn().Err(err).Str("server", srv.Name).Str("url", srv.URL).Msg("Failed to connect to upstream MCP server")
			h.recordUpstreamFailure(srv, err)
			return nil
		}
		client = newClient
	} else {
		client = existing.client
	}

	toolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Warn().Err(err).Str("server", srv.Name).Msg("Failed to list tools from upstream MCP server")
		h.recordUpstreamFailure(srv, err)
		return nil
	}

	var tools []server.ServerTool
	var upstreamToolDefs []mcp.Tool
	seenNames := make(map[string]bool)

	for _, t := range toolsResult.Tools {
		namespacedName := srv.Name + "__" + t.Name

		// Native tools win on name collision.
		if _, exists := nativeToolNames[namespacedName]; exists {
			log.Warn().Str("tool", namespacedName).Str("server", srv.Name).Msg("Upstream tool name collides with native tool, skipping")
			continue
		}

		if seenNames[namespacedName] {
			continue
		}
		seenNames[namespacedName] = true

		toolDef := t
		toolDef.Name = namespacedName
		upstreamToolDefs = append(upstreamToolDefs, toolDef)

		tools = append(tools, server.ServerTool{
			Tool:    toolDef,
			Handler: h.createUpstreamToolHandler(client, t.Name),
		})
	}

	h.mu.Lock()
	h.upstreams[srv.Name] = &upstreamServer{
		client:         client,
		serverID:       srv.ID,
		name:           srv.Name,
		url:            srv.URL,
		authHeader:     srv.AuthHeader,
		authValue:      authValue,
		connectionName: srv.ConnectionName,
		tools:          upstreamToolDefs,
		failureCount:   0,
	}
	h.mu.Unlock()

	if h.serverRepo != nil {
		_ = h.serverRepo.UpdateSyncStatus(srv.ID, len(tools), "")
	}

	log.Debug().Str("server", srv.Name).Int("tool_count", len(tools)).Msg("Upstream MCP server synced")

	return tools
}

func (h *MCPHandler) recordUpstreamFailure(srv *persistence.MCPServer, err error) {
	h.mu.Lock()
	existing := h.upstreams[srv.Name]
	if existing == nil {
		existing = &upstreamServer{name: srv.Name, serverID: srv.ID, connectionName: srv.ConnectionName}
		h.upstreams[srv.Name] = existing
	}
	existing.connectionName = srv.ConnectionName
	existing.failureCount++
	failCount := existing.failureCount
	if failCount >= maxUpstreamFailures {
		existing.failedAt = time.Now()
	}
	h.mu.Unlock()

	if h.serverRepo != nil {
		errMsg := err.Error()
		_ = h.serverRepo.UpdateSyncStatus(srv.ID, 0, errMsg)
		if failCount >= maxUpstreamFailures {
			_ = h.serverRepo.UpdateStatus(srv.ID, "error")
		}
	}
}

func (h *MCPHandler) createNativeToolHandler(flowID int64) server.ToolHandlerFunc {
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

func (h *MCPHandler) createUpstreamToolHandler(client *mcpclient.Client, originalToolName string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      originalToolName,
				Arguments: request.GetArguments(),
			},
		})

		latency := time.Since(start)
		log.Debug().
			Str("tool", originalToolName).
			Dur("latency", latency).
			Err(err).
			Msg("Upstream MCP tool call")

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("upstream tool call failed: %v", err)), nil
		}

		return result, nil
	}
}

// ResetUpstreamServer resets the circuit breaker for a server, allowing it to be re-synced.
func (h *MCPHandler) ResetUpstreamServer(serverName string) {
	h.mu.Lock()
	if us, ok := h.upstreams[serverName]; ok {
		us.failureCount = 0
		if us.client != nil {
			us.client = nil
		}
	}
	delete(h.upstreams, serverName)
	h.mu.Unlock()
}

// resetUpstreamsForConnection clears the in-memory circuit breaker and flips
// errored DB rows back to active for any MCP server tied to the given
// connection. Wired to connection.Manager's token-refreshed event so a
// successful refresh re-arms servers that were stuck on stale-credential 401s.
func (h *MCPHandler) resetUpstreamsForConnection(connectionName string) {
	if connectionName == "" {
		return
	}

	h.mu.Lock()
	for name, us := range h.upstreams {
		if us.connectionName != connectionName {
			continue
		}
		log.Debug().Str("server", name).Str("connection", connectionName).Msg("Connection refreshed - clearing upstream circuit breaker")
		if us.client != nil {
			us.client = nil
		}
		delete(h.upstreams, name)
	}
	h.mu.Unlock()

	if h.serverRepo != nil {
		if n, err := h.serverRepo.ReactivateByConnection(connectionName); err != nil {
			log.Warn().Err(err).Str("connection", connectionName).Msg("Failed to reactivate MCP servers after connection refresh")
		} else if n > 0 {
			log.Debug().Int64("count", n).Str("connection", connectionName).Msg("Reactivated MCP servers after connection refresh")
		}
	}
}

// resolveAuth returns the (header, value) pair for the upstream server's
// auth_type. "connection" goes through connection.Manager.GetAccessToken,
// which auto-refreshes near expiry; "token" decrypts the stored static value.
func (h *MCPHandler) resolveAuth(ctx context.Context, srv *persistence.MCPServer) (header, value string) {
	switch srv.AuthType {
	case "token":
		if srv.EncryptedAuthValue != "" && h.aesgcm != nil {
			decrypted, err := h.aesgcm.Decrypt(srv.EncryptedAuthValue)
			if err != nil {
				log.Warn().Err(err).Str("server", srv.Name).Msg("Failed to decrypt auth value")
				return "", ""
			}
			return srv.AuthHeader, decrypted
		}
	case "connection":
		if srv.ConnectionName != "" && h.connManager != nil {
			tok, err := h.connManager.GetAccessToken(ctx, srv.ConnectionName, false)
			if err != nil {
				log.Warn().Err(err).Str("server", srv.Name).Str("connection", srv.ConnectionName).Msg("Failed to get access token for MCP server auth")
				return "", ""
			}
			// Empty header -> sent as Authorization: Bearer.
			return "", tok.AccessToken
		}
	}
	return "", ""
}

func computeToolsHash(tools []server.ServerTool) string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Tool.Name
	}
	sort.Strings(names)
	return strings.Join(names, ",")
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
