package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	client   *mcpclient.Client
	serverID int64
	name     string
	url      string
	// authKey is an opaque value that changes only when we need to reconnect
	// (e.g. token-auth value changed, connection name changed). For
	// connection-typed servers it intentionally does NOT change on token
	// rotation - the transport handles fresh tokens per request.
	authKey      string
	tools        []mcp.Tool
	failureCount int
	// failedAt is set when the breaker trips. Past upstreamCircuitCooldown the
	// next sync attempt is allowed through.
	failedAt time.Time
}

const (
	maxUpstreamFailures     = 3
	upstreamSyncTimeout     = 10 * time.Second
	upstreamCircuitCooldown = 5 * time.Minute
)

type MCPHandler struct {
	mcpServer       *server.MCPServer
	httpHandler     *server.StreamableHTTPServer
	flowRepo        persistence.FlowRepository
	serverRepo      persistence.MCPServerRepository
	forwarder       RequestForwarder
	aesgcm          *vault.AESGCM
	connManager     *connection.Manager
	stdioSupervisor *StdioSupervisor
	mu              sync.RWMutex
	syncMu          sync.Mutex
	// tool name -> flow ID, used for forwarding via /ingest/{flowID}
	toolFlowMap   map[string]int64
	upstreams     map[string]*upstreamServer
	lastToolsHash string
}

func NewMCPHandler(flowRepo persistence.FlowRepository, serverRepo persistence.MCPServerRepository, secretRepo persistence.SecretRepository, forwarder RequestForwarder, aesgcm *vault.AESGCM, connManager *connection.Manager, version string) *MCPHandler {
	mcpServer := server.NewMCPServer(
		"qaynaq",
		version,
		server.WithToolCapabilities(true),
	)

	httpHandler := server.NewStreamableHTTPServer(mcpServer)

	envResolver := NewEnvResolver(secretRepo, aesgcm)
	stdioSup := NewStdioSupervisor(envResolver)

	h := &MCPHandler{
		mcpServer:       mcpServer,
		httpHandler:     httpHandler,
		flowRepo:        flowRepo,
		serverRepo:      serverRepo,
		forwarder:       forwarder,
		aesgcm:          aesgcm,
		connManager:     connManager,
		stdioSupervisor: stdioSup,
		toolFlowMap:     make(map[string]int64),
		upstreams:       make(map[string]*upstreamServer),
	}

	h.SyncTools()
	return h
}

func (h *MCPHandler) StdioSupervisor() *StdioSupervisor {
	return h.stdioSupervisor
}

func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.httpHandler.ServeHTTP(w, r)
}

func (h *MCPHandler) SyncTools() {
	h.syncMu.Lock()
	defer h.syncMu.Unlock()

	nativeTools, nativeToolMap := h.syncNativeTools()
	upstreamTools, reconnected := h.syncUpstreamServers(nativeToolMap)
	allTools := append(nativeTools, upstreamTools...)

	// Skip SetTools when nothing changed - it would needlessly notify clients.
	// But if any upstream rotated its client, we MUST re-register tool handlers
	// even if tool names are identical, otherwise the previously-registered
	// handlers still close over the old (now closed) client.
	hash := computeToolsHash(allTools)
	h.mu.Lock()
	skip := !reconnected && hash == h.lastToolsHash
	h.lastToolsHash = hash
	h.mu.Unlock()
	if skip {
		return
	}

	h.mcpServer.SetTools(allTools...)

	log.Debug().
		Int("native_count", len(nativeTools)).
		Int("upstream_count", len(upstreamTools)).
		Int("total_count", len(allTools)).
		Bool("reconnected", reconnected).
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

// syncUpstreamServers returns the registered upstream tools and a flag
// indicating whether any upstream client was created or rotated this pass.
// The flag forces SetTools even when tool names are unchanged, since the
// previously-registered handlers close over the now-closed client.
func (h *MCPHandler) syncUpstreamServers(nativeToolNames map[string]int64) ([]server.ServerTool, bool) {
	if h.serverRepo == nil {
		return nil, false
	}

	servers, err := h.serverRepo.ListMonitored()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list MCP servers for upstream sync")
		return nil, false
	}

	if len(servers) == 0 {
		h.mu.Lock()
		removed := len(h.upstreams) > 0
		for name, us := range h.upstreams {
			h.releaseUpstream(us)
			delete(h.upstreams, name)
		}
		h.mu.Unlock()
		return nil, removed
	}

	activeNames := make(map[string]bool, len(servers))
	for _, s := range servers {
		activeNames[s.Name] = true
	}

	// Drop upstream entries for servers that have been removed from the DB.
	h.mu.Lock()
	rotated := false
	for name, us := range h.upstreams {
		if !activeNames[name] {
			h.releaseUpstream(us)
			delete(h.upstreams, name)
			rotated = true
		}
	}
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), upstreamSyncTimeout)
	defer cancel()

	type serverResult struct {
		name        string
		tools       []server.ServerTool
		reconnected bool
	}

	var resultMu sync.Mutex
	var results []serverResult

	g, gCtx := errgroup.WithContext(ctx)

	for _, srv := range servers {
		g.Go(func() error {
			tools, reconnected := h.syncOneUpstream(gCtx, &srv, nativeToolNames)
			resultMu.Lock()
			results = append(results, serverResult{name: srv.Name, tools: tools, reconnected: reconnected})
			resultMu.Unlock()
			// Never fail the group - one bad server shouldn't block the rest.
			return nil
		})
	}

	_ = g.Wait()

	var allUpstreamTools []server.ServerTool
	for _, r := range results {
		allUpstreamTools = append(allUpstreamTools, r.tools...)
		if r.reconnected {
			rotated = true
		}
	}

	return allUpstreamTools, rotated
}

// syncOneUpstream returns the tools registered for srv plus a flag that's
// true when this call rotated the underlying mcp client (handlers registered
// elsewhere now close over a stale pointer).
func (h *MCPHandler) syncOneUpstream(ctx context.Context, srv *persistence.MCPServer, nativeToolNames map[string]int64) ([]server.ServerTool, bool) {
	h.mu.RLock()
	existing := h.upstreams[srv.Name]
	h.mu.RUnlock()

	// Circuit breaker: while inside the cooldown window we skip; past it we
	// let one attempt through. Success clears the counter, failure restarts
	// the cooldown.
	if existing != nil && existing.failureCount >= maxUpstreamFailures {
		if !existing.failedAt.IsZero() && time.Since(existing.failedAt) < upstreamCircuitCooldown {
			log.Debug().Str("server", srv.Name).Msg("Upstream server circuit-broken, skipping")
			return nil, false
		}
		log.Debug().Str("server", srv.Name).Msg("Upstream server cooldown elapsed, allowing retry")
	}

	if srv.Transport == persistence.MCPTransportStdio {
		return h.syncOneStdioUpstream(ctx, srv, existing, nativeToolNames)
	}

	auth := h.resolveAuth(srv)
	needReconnect := existing == nil || existing.url != srv.URL || existing.authKey != auth.key

	var client *mcpclient.Client

	if needReconnect {
		closeUpstreamClient(existing)

		newClient, err := ConnectMCPClient(ctx, srv.URL, auth.headers, auth.httpClient)
		if err != nil {
			log.Warn().Err(err).Str("server", srv.Name).Str("url", srv.URL).Msg("Failed to connect to upstream MCP server")
			h.recordUpstreamFailure(srv, err)
			return nil, false
		}
		client = newClient
	} else {
		client = existing.client
	}

	toolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Warn().Err(err).Str("server", srv.Name).Msg("Failed to list tools from upstream MCP server")
		h.recordUpstreamFailure(srv, err)
		return nil, needReconnect
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
		client:       client,
		serverID:     srv.ID,
		name:         srv.Name,
		url:          srv.URL,
		authKey:      auth.key,
		tools:        upstreamToolDefs,
		failureCount: 0,
	}
	h.mu.Unlock()

	if h.serverRepo != nil {
		_ = h.serverRepo.UpdateSyncStatus(srv.ID, len(tools), "")
		if srv.Status == "error" {
			_ = h.serverRepo.UpdateStatus(srv.ID, "active")
		}
	}

	log.Debug().Str("server", srv.Name).Int("tool_count", len(tools)).Msg("Upstream MCP server synced")

	return tools, needReconnect
}

func (h *MCPHandler) syncOneStdioUpstream(ctx context.Context, srv *persistence.MCPServer, existing *upstreamServer, nativeToolNames map[string]int64) ([]server.ServerTool, bool) {
	if h.stdioSupervisor == nil {
		log.Warn().Str("server", srv.Name).Msg("Stdio supervisor not configured, skipping")
		return nil, false
	}

	client, err := h.stdioSupervisor.Get(ctx, srv)
	if err != nil {
		switch {
		case errors.Is(err, ErrStarting):
			log.Debug().Str("server", srv.Name).Msg("Stdio MCP server starting, will retry next sync")
			return nil, false
		case errors.Is(err, ErrFailed):
			if h.serverRepo != nil {
				_ = h.serverRepo.UpdateSyncStatus(srv.ID, 0, err.Error())
				_ = h.serverRepo.UpdateStatus(srv.ID, "error")
			}
			return nil, false
		case errors.Is(err, ErrCapExceeded):
			log.Warn().Str("server", srv.Name).Msg("Stdio MCP process cap reached, skipping spawn")
			return nil, false
		default:
			log.Warn().Err(err).Str("server", srv.Name).Msg("Stdio MCP supervisor returned error")
			h.recordUpstreamFailure(srv, err)
			return nil, false
		}
	}

	needReconnect := existing == nil || existing.client != client

	toolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Warn().Err(err).Str("server", srv.Name).Msg("Failed to list tools from stdio MCP server")
		h.recordUpstreamFailure(srv, err)
		return nil, needReconnect
	}

	var tools []server.ServerTool
	var upstreamToolDefs []mcp.Tool
	seenNames := make(map[string]bool)

	for _, t := range toolsResult.Tools {
		namespacedName := srv.Name + "__" + t.Name
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
			Handler: h.createStdioToolHandler(srv.ID, srv.Name, t.Name),
		})
	}

	h.mu.Lock()
	h.upstreams[srv.Name] = &upstreamServer{
		client:       client,
		serverID:     srv.ID,
		name:         srv.Name,
		url:          "",
		authKey:      "stdio",
		tools:        upstreamToolDefs,
		failureCount: 0,
	}
	h.mu.Unlock()

	if h.serverRepo != nil {
		_ = h.serverRepo.UpdateSyncStatus(srv.ID, len(tools), "")
		_ = h.serverRepo.UpdateProcessState(srv.ID, h.stdioSupervisor.State(srv.ID))
		if srv.Status == "error" {
			_ = h.serverRepo.UpdateStatus(srv.ID, "active")
		}
	}

	log.Debug().Str("server", srv.Name).Int("tool_count", len(tools)).Msg("Stdio MCP server synced")
	return tools, needReconnect
}

func (h *MCPHandler) recordUpstreamFailure(srv *persistence.MCPServer, err error) {
	h.mu.Lock()
	existing := h.upstreams[srv.Name]
	if existing == nil {
		existing = &upstreamServer{name: srv.Name, serverID: srv.ID}
		h.upstreams[srv.Name] = existing
	}
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

func (h *MCPHandler) createStdioToolHandler(serverID int64, serverName, originalToolName string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		// Always re-fetch through the supervisor: the client cached when this
		// handler was registered may have been replaced by a respawn or
		// stopped by the idle timer.
		srv, err := h.serverRepo.GetByID(serverID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("MCP server %q lookup failed: %v", serverName, err)), nil
		}
		client, err := h.stdioSupervisor.Get(ctx, srv)
		if err != nil {
			switch {
			case errors.Is(err, ErrStarting):
				return mcp.NewToolResultError(fmt.Sprintf("MCP server %q is starting, retry in 3s", serverName)), nil
			case errors.Is(err, ErrFailed):
				return mcp.NewToolResultError(fmt.Sprintf("MCP server %q is in failed state: %v", serverName, err)), nil
			default:
				return mcp.NewToolResultError(fmt.Sprintf("MCP server %q unavailable: %v", serverName, err)), nil
			}
		}

		result, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      originalToolName,
				Arguments: request.GetArguments(),
			},
		})

		latency := time.Since(start)
		log.Debug().
			Str("tool", originalToolName).
			Str("server", serverName).
			Dur("latency", latency).
			Err(err).
			Msg("Stdio MCP tool call")

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("stdio tool call failed: %v", err)), nil
		}

		h.stdioSupervisor.Touch(serverID)
		return result, nil
	}
}

// ResetUpstreamServer drops the cached HTTP client and circuit-breaker state.
// Skips stdio: releaseUpstream would tear down the supervised child process.
func (h *MCPHandler) ResetUpstreamServer(serverName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	us, ok := h.upstreams[serverName]
	if !ok {
		return
	}
	if us.authKey == "stdio" {
		return
	}
	h.releaseUpstream(us)
	delete(h.upstreams, serverName)
}

// closeUpstreamClient closes the upstream MCP client (terminating its
// listener goroutine and the upstream session) and nils the field. Safe to
// call with a nil us or nil us.client.
func closeUpstreamClient(us *upstreamServer) {
	if us == nil || us.client == nil {
		return
	}
	if err := us.client.Close(); err != nil {
		log.Debug().Err(err).Str("server", us.name).Msg("Failed to close upstream MCP client")
	}
	us.client = nil
}

// releaseUpstream tears down an upstream entry. For stdio servers the
// supervisor owns the client and the child process, so we route through
// Remove rather than closing the client directly.
func (h *MCPHandler) releaseUpstream(us *upstreamServer) {
	if us == nil {
		return
	}
	if us.authKey == "stdio" && h.stdioSupervisor != nil {
		h.stdioSupervisor.Remove(us.serverID)
		us.client = nil
		return
	}
	closeUpstreamClient(us)
}

// upstreamAuth is everything syncOneUpstream needs to connect to an upstream
// MCP server: a stable key for change detection (so we don't reconnect on
// every token rotation), optional static headers, and an optional custom
// HTTP client for connection-typed auth (where token refresh + 401 retry
// happen inside the transport).
type upstreamAuth struct {
	key        string
	headers    map[string]string
	httpClient *http.Client
}

func (h *MCPHandler) resolveAuth(srv *persistence.MCPServer) upstreamAuth {
	switch srv.AuthType {
	case "token":
		if srv.EncryptedAuthValue != "" && h.aesgcm != nil {
			decrypted, err := h.aesgcm.Decrypt(srv.EncryptedAuthValue)
			if err != nil {
				log.Warn().Err(err).Str("server", srv.Name).Msg("Failed to decrypt auth value")
				return upstreamAuth{}
			}
			return upstreamAuth{
				key:     "token:" + srv.AuthHeader + "\x00" + decrypted,
				headers: BuildAuthHeaders(srv.AuthHeader, decrypted),
			}
		}
	case "connection":
		if srv.ConnectionName != "" && h.connManager != nil {
			return upstreamAuth{
				key:        "connection:" + srv.ConnectionName,
				httpClient: newConnectionHTTPClient(h.connManager, srv.ConnectionName),
			}
		}
	}
	return upstreamAuth{}
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
