package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/warpstreamlabs/bento/public/service"
	"gopkg.in/yaml.v3"
)

const tryFlowTimeout = 60 * time.Second

type TryMessage struct {
	Content string `json:"content"`
}

type TryResult struct {
	Outputs []TryOutput `json:"outputs"`
	Error   string      `json:"error,omitempty"`
}

type TryOutput struct {
	Content string `json:"content"`
}

type TryFlowOptions struct {
	EnvVarLookupFn func(string) (string, bool)
	FileRepo       persistence.FileRepository
}

func TryStream(ctx context.Context, processors []persistence.FlowProcessor, messages []TryMessage, opts TryFlowOptions) *TryResult {
	if len(messages) == 0 {
		return &TryResult{Error: "at least one test message is required"}
	}

	builder := service.NewStreamBuilder()
	builder.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

	if opts.EnvVarLookupFn != nil {
		builder.SetEnvVarLookupFunc(opts.EnvVarLookupFn)
	}

	b := &configBuilder{}

	processorMaps := make([]map[string]any, 0, len(processors))
	for _, proc := range processors {
		procMap, err := b.buildProcessorConfig(proc)
		if err != nil {
			return &TryResult{Error: fmt.Sprintf("failed to build processor %q config: %s", proc.Label, err)}
		}
		processorMaps = append(processorMaps, procMap)
	}

	if opts.FileRepo != nil {
		var allFileKeys []string
		for _, procMap := range processorMaps {
			allFileKeys = append(allFileKeys, collectFileRefs(procMap)...)
		}
		if len(allFileKeys) > 0 {
			files, err := opts.FileRepo.FindByKeys(allFileKeys)
			if err != nil {
				return &TryResult{Error: fmt.Sprintf("failed to fetch referenced files: %s", err)}
			}

			tmpDir, err := os.MkdirTemp("", "qaynaq-try-*")
			if err != nil {
				return &TryResult{Error: fmt.Sprintf("failed to create temp directory: %s", err)}
			}
			defer os.RemoveAll(tmpDir)

			for _, f := range files {
				filePath := filepath.Join(tmpDir, f.Key)
				if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
					return &TryResult{Error: fmt.Sprintf("failed to create directory for file %q: %s", f.Key, err)}
				}
				if err := os.WriteFile(filePath, f.Content, 0o644); err != nil {
					return &TryResult{Error: fmt.Sprintf("failed to write file %q: %s", f.Key, err)}
				}
			}

			for _, procMap := range processorMaps {
				resolveFileRefsToDir(procMap, tmpDir)
			}
		}
	}

	for i, procMap := range processorMaps {
		yamlBytes, err := yaml.Marshal(procMap)
		if err != nil {
			return &TryResult{Error: fmt.Sprintf("failed to marshal processor %q config: %s", processors[i].Label, err)}
		}
		if err := builder.AddProcessorYAML(string(yamlBytes)); err != nil {
			return &TryResult{Error: fmt.Sprintf("invalid processor %q config: %s", processors[i].Label, err)}
		}
	}

	sendFn, err := builder.AddProducerFunc()
	if err != nil {
		return &TryResult{Error: fmt.Sprintf("failed to set up test input: %s", err)}
	}

	var (
		mu        sync.Mutex
		outputs   []TryOutput
		received  int
		outputsCh = make(chan struct{}, 1)
	)
	expectedCount := len(messages)

	if err := builder.AddConsumerFunc(func(_ context.Context, m *service.Message) error {
		bytes, err := m.AsBytes()
		if err != nil {
			return err
		}

		content := string(bytes)
		if json.Valid(bytes) {
			var structured any
			if err := json.Unmarshal(bytes, &structured); err == nil {
				if formatted, err := json.MarshalIndent(structured, "", "  "); err == nil {
					content = string(formatted)
				}
			}
		}

		mu.Lock()
		outputs = append(outputs, TryOutput{Content: content})
		received++
		done := received >= expectedCount
		mu.Unlock()

		if done {
			select {
			case outputsCh <- struct{}{}:
			default:
			}
		}
		return nil
	}); err != nil {
		return &TryResult{Error: fmt.Sprintf("failed to set up test output: %s", err)}
	}

	flow, err := builder.Build()
	if err != nil {
		return &TryResult{Error: fmt.Sprintf("failed to build flow: %s", err)}
	}

	tryCtx, cancel := context.WithTimeout(ctx, tryFlowTimeout)
	defer cancel()

	flowDone := make(chan error, 1)
	go func() {
		flowDone <- flow.Run(tryCtx)
	}()

	for _, msg := range messages {
		if err := sendFn(tryCtx, service.NewMessage([]byte(msg.Content))); err != nil {
			cancel()
			<-flowDone
			return &TryResult{Error: fmt.Sprintf("failed to send test message: %s", err)}
		}
	}

	// Wait for all outputs to arrive, the flow to exit, or the timeout.
	select {
	case <-outputsCh:
	case <-time.After(3 * time.Second):
	}

	if err := flow.StopWithin(5 * time.Second); err != nil {
		cancel()
		<-flowDone
		return &TryResult{Error: fmt.Sprintf("flow did not stop gracefully: %s", err)}
	}

	<-flowDone

	return &TryResult{Outputs: outputs}
}

func resolveFileRefsToDir(v any, dir string) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if s, ok := child.(string); ok && len(s) > len(FileRefPrefix) && s[:len(FileRefPrefix)] == FileRefPrefix {
				key := s[len(FileRefPrefix):]
				val[k] = "file://" + filepath.Join(dir, key)
			} else {
				resolveFileRefsToDir(child, dir)
			}
		}
	case []any:
		for i, child := range val {
			if s, ok := child.(string); ok && len(s) > len(FileRefPrefix) && s[:len(FileRefPrefix)] == FileRefPrefix {
				key := s[len(FileRefPrefix):]
				val[i] = "file://" + filepath.Join(dir, key)
			} else {
				resolveFileRefsToDir(child, dir)
			}
		}
	}
}
