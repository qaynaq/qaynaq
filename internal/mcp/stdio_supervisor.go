package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

var (
	// ErrStarting tells callers the process is being spawned in the background.
	// The sync loop treats this as "skip, retry next pass" without burning a
	// circuit-breaker slot.
	ErrStarting = errors.New("stdio mcp server is starting")

	// ErrCapExceeded means the global concurrent-process cap was hit.
	ErrCapExceeded = errors.New("stdio mcp process cap exceeded")

	// ErrFailed means the server is in a terminal failed state until the user
	// fixes the cause and triggers a manual restart.
	ErrFailed = errors.New("stdio mcp server is in failed state")

	// ErrUnknownCatalog is returned when a stored catalog ID is no longer in
	// the allowlist (catalog entry was removed in a release).
	ErrUnknownCatalog = errors.New("stdio catalog entry not found")
)

const (
	procStateStopped     = "stopped"
	procStateStarting    = "starting"
	procStateRunning     = "running"
	procStateIdleStopped = "idle_stopped"
	procStateBackoff     = "backoff"
	procStateFailed      = "failed"
)

const (
	defaultMaxTotal       = 100
	defaultIdleTimeout    = 15 * time.Minute
	stdioHandshakeTimeout = 30 * time.Second
	stderrRingSize        = 4 * 1024
	maxCrashesInWindow    = 3
	crashWindow           = 5 * time.Minute
	livenessInterval      = 15 * time.Second
	livenessTimeout       = 5 * time.Second
)

type stdioProc struct {
	serverID    int64
	state       string
	client      *mcpclient.Client
	stderrRing  *ringBuffer
	idleTimer   *time.Timer
	intentional bool
	watchCancel context.CancelFunc

	failureTimes []time.Time
	lastErr      string
}

type StdioSupervisor struct {
	envResolver *EnvResolver

	mu    sync.Mutex
	procs map[int64]*stdioProc
	total int

	maxTotal    int
	idleTimeout time.Duration
}

func NewStdioSupervisor(envResolver *EnvResolver) *StdioSupervisor {
	return &StdioSupervisor{
		envResolver: envResolver,
		procs:       make(map[int64]*stdioProc),
		maxTotal:    defaultMaxTotal,
		idleTimeout: defaultIdleTimeout,
	}
}

// Get returns a live MCP client for srv. If the process is not running it is
// spawned in the background and ErrStarting is returned immediately so the
// sync loop never blocks on a slow npx download.
func (s *StdioSupervisor) Get(ctx context.Context, srv *persistence.MCPServer) (*mcpclient.Client, error) {
	s.mu.Lock()
	proc, ok := s.procs[srv.ID]
	if !ok {
		proc = &stdioProc{
			serverID:   srv.ID,
			state:      procStateStopped,
			stderrRing: newRingBuffer(stderrRingSize),
		}
		s.procs[srv.ID] = proc
	}

	switch proc.state {
	case procStateRunning:
		client := proc.client
		s.mu.Unlock()
		return client, nil
	case procStateStarting:
		s.mu.Unlock()
		return nil, ErrStarting
	case procStateFailed:
		err := proc.lastErr
		s.mu.Unlock()
		if err == "" {
			return nil, ErrFailed
		}
		return nil, fmt.Errorf("%w: %s", ErrFailed, err)
	}

	if s.total >= s.maxTotal {
		s.mu.Unlock()
		return nil, ErrCapExceeded
	}

	proc.state = procStateStarting
	proc.intentional = false
	s.total++
	srvCopy := *srv
	s.mu.Unlock()

	go s.spawn(&srvCopy) //nolint:gosec // supervisor manages spawned process lifecycle independently
	return nil, ErrStarting
}

func (s *StdioSupervisor) Stop(serverID int64) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok {
		s.mu.Unlock()
		return
	}
	proc.intentional = true
	if proc.idleTimer != nil {
		proc.idleTimer.Stop()
		proc.idleTimer = nil
	}
	if proc.watchCancel != nil {
		proc.watchCancel()
		proc.watchCancel = nil
	}
	c := proc.client
	wasActive := proc.state == procStateRunning || proc.state == procStateStarting
	proc.state = procStateStopped
	proc.client = nil
	if wasActive {
		s.total--
	}
	s.mu.Unlock()

	if c != nil {
		_ = c.Close()
	}
}

func (s *StdioSupervisor) Remove(serverID int64) {
	s.Stop(serverID)
	s.mu.Lock()
	delete(s.procs, serverID)
	s.mu.Unlock()
}

func (s *StdioSupervisor) Restart(serverID int64) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok {
		s.mu.Unlock()
		return
	}
	proc.failureTimes = nil
	proc.lastErr = ""
	proc.intentional = false
	if proc.state == procStateFailed {
		proc.state = procStateStopped
	}
	s.mu.Unlock()
}

func (s *StdioSupervisor) Touch(serverID int64) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok || proc.idleTimer == nil {
		s.mu.Unlock()
		return
	}
	proc.idleTimer.Reset(s.idleTimeout)
	s.mu.Unlock()
}

func (s *StdioSupervisor) LastError(serverID int64) (lastErr, stderr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proc, ok := s.procs[serverID]
	if !ok {
		return "", ""
	}
	if proc.stderrRing != nil {
		stderr = proc.stderrRing.String()
	}
	return proc.lastErr, stderr
}

func (s *StdioSupervisor) State(serverID int64) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	proc, ok := s.procs[serverID]
	if !ok {
		return procStateStopped
	}
	return proc.state
}

func (s *StdioSupervisor) Close() {
	s.mu.Lock()
	ids := make([]int64, 0, len(s.procs))
	for id := range s.procs {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	for _, id := range ids {
		s.Stop(id)
	}
}

func (s *StdioSupervisor) spawn(srv *persistence.MCPServer) {
	if srv.Transport != persistence.MCPTransportStdio {
		s.recordCrash(srv.ID, fmt.Errorf("non-stdio server passed to spawn: %s", srv.Transport))
		return
	}

	entry, ok := LookupCatalogEntry(srv.CatalogID)
	if !ok {
		s.recordCrash(srv.ID, fmt.Errorf("%w: %s", ErrUnknownCatalog, srv.CatalogID))
		return
	}

	rawEnv, err := decodeEnvBlob(srv.EncryptedEnv, s.envResolver)
	if err != nil {
		s.recordCrash(srv.ID, fmt.Errorf("decode env: %w", err))
		return
	}

	resolvedEnv, err := s.envResolver.Resolve(rawEnv)
	if err != nil {
		// Missing-secret failures stay failed without respawn: retrying with the
		// same input would just loop. User must fix the secret first.
		s.markFailedNoRetry(srv.ID, err)
		return
	}

	args, err := SubstituteArgs(entry.ArgsTemplate, rawEnv)
	if err != nil {
		s.markFailedNoRetry(srv.ID, fmt.Errorf("arg substitution: %w", err))
		return
	}

	log.Debug().
		Str("server", srv.Name).
		Str("command", entry.Command).
		Strs("args", args).
		Msg("Spawning stdio MCP process")

	client, err := newStdioMCPClientWithStderr(entry.Command, resolvedEnv, args, s.getRing(srv.ID))
	if err != nil {
		s.recordCrash(srv.ID, fmt.Errorf("spawn: %w", err))
		return
	}

	hctx, cancel := context.WithTimeout(context.Background(), stdioHandshakeTimeout)
	defer cancel()
	if _, err := client.Initialize(hctx, mcp.InitializeRequest{}); err != nil {
		_ = client.Close()
		s.recordCrash(srv.ID, fmt.Errorf("initialize: %w", err))
		return
	}

	watchCtx, watchCancel := context.WithCancel(context.Background())

	s.mu.Lock()
	proc, ok := s.procs[srv.ID]
	if !ok || proc.intentional {
		s.mu.Unlock()
		watchCancel()
		_ = client.Close()
		return
	}
	proc.client = client
	proc.state = procStateRunning
	proc.lastErr = ""
	proc.watchCancel = watchCancel
	proc.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		s.idleStop(srv.ID)
	})
	s.mu.Unlock()

	go s.watchLiveness(watchCtx, srv.ID, client)

	log.Info().Str("server", srv.Name).Msg("Stdio MCP server running")
}

func (s *StdioSupervisor) idleStop(serverID int64) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok || proc.state != procStateRunning {
		s.mu.Unlock()
		return
	}
	proc.intentional = true
	proc.state = procStateIdleStopped
	c := proc.client
	proc.client = nil
	if proc.idleTimer != nil {
		proc.idleTimer.Stop()
		proc.idleTimer = nil
	}
	if proc.watchCancel != nil {
		proc.watchCancel()
		proc.watchCancel = nil
	}
	s.total--
	s.mu.Unlock()

	if c != nil {
		_ = c.Close()
	}
	log.Debug().Int64("server_id", serverID).Msg("Stdio MCP server idle-stopped")
}

func (s *StdioSupervisor) watchLiveness(ctx context.Context, serverID int64, client *mcpclient.Client) {
	ticker := time.NewTicker(livenessInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		pingCtx, cancel := context.WithTimeout(ctx, livenessTimeout)
		err := client.Ping(pingCtx)
		cancel()
		if err == nil {
			continue
		}
		if ctx.Err() != nil {
			return
		}
		log.Warn().Int64("server_id", serverID).Err(err).Msg("Stdio MCP liveness ping failed")
		s.recordCrash(serverID, fmt.Errorf("liveness ping: %w", err))
		return
	}
}

func (s *StdioSupervisor) recordCrash(serverID int64, err error) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok {
		s.mu.Unlock()
		return
	}
	// If Stop ran concurrently with the spawn it already cleaned up state and
	// decremented total. Don't double-account.
	if proc.intentional {
		s.mu.Unlock()
		return
	}
	proc.lastErr = err.Error()
	proc.state = procStateBackoff
	if proc.watchCancel != nil {
		proc.watchCancel()
		proc.watchCancel = nil
	}
	if proc.idleTimer != nil {
		proc.idleTimer.Stop()
		proc.idleTimer = nil
	}
	if proc.client != nil {
		_ = proc.client.Close()
		proc.client = nil
	}

	now := time.Now()
	proc.failureTimes = append(proc.failureTimes, now)
	cutoff := now.Add(-crashWindow)
	pruned := proc.failureTimes[:0]
	for _, t := range proc.failureTimes {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	proc.failureTimes = pruned

	if len(proc.failureTimes) >= maxCrashesInWindow {
		proc.state = procStateFailed
		s.total--
		s.mu.Unlock()
		log.Warn().Int64("server_id", serverID).Err(err).Msg("Stdio MCP server entered failed state")
		return
	}
	s.total--
	s.mu.Unlock()

	log.Warn().Int64("server_id", serverID).Err(err).Msg("Stdio MCP server crashed")
}

func (s *StdioSupervisor) markFailedNoRetry(serverID int64, err error) {
	s.mu.Lock()
	proc, ok := s.procs[serverID]
	if !ok {
		s.mu.Unlock()
		return
	}
	if proc.intentional {
		s.mu.Unlock()
		return
	}
	proc.state = procStateFailed
	proc.lastErr = err.Error()
	if proc.client != nil {
		_ = proc.client.Close()
		proc.client = nil
	}
	s.total--
	s.mu.Unlock()
	log.Warn().Int64("server_id", serverID).Err(err).Msg("Stdio MCP server failed (no retry)")
}

// getRing returns a fresh ring buffer for this spawn attempt and replaces the
// proc's reference, so a respawn does not surface the previous run's stderr.
func (s *StdioSupervisor) getRing(serverID int64) *ringBuffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	proc, ok := s.procs[serverID]
	if !ok {
		return nil
	}
	proc.stderrRing = newRingBuffer(stderrRingSize)
	return proc.stderrRing
}

func decodeEnvBlob(blob string, r *EnvResolver) (map[string]string, error) {
	if blob == "" {
		return map[string]string{}, nil
	}
	if r.AESGCM() == nil {
		return nil, fmt.Errorf("vault not configured")
	}
	plain, err := r.AESGCM().Decrypt(blob)
	if err != nil {
		return nil, fmt.Errorf("decrypt env blob: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(plain), &m); err != nil {
		return nil, fmt.Errorf("unmarshal env blob: %w", err)
	}
	return m, nil
}

func EncodeEnvBlob(env map[string]string, r *EnvResolver) (string, error) {
	if len(env) == 0 {
		return "", nil
	}
	if r.AESGCM() == nil {
		return "", fmt.Errorf("vault not configured")
	}
	b, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("marshal env: %w", err)
	}
	return r.AESGCM().Encrypt(string(b))
}
