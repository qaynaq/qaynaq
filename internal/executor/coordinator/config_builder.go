package coordinator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

var envVarPattern = regexp.MustCompile(`^\$\{[^!][^}]*\}$`)

const (
	FileRefPrefix  = "qaynaq://"
	WorkerFilesDir = "/tmp/qaynaq/files"
)

type BuildResult struct {
	Config string
	Files  []persistence.File
}

type ConfigBuilder interface {
	BuildFlowConfig(flow persistence.Flow) (*BuildResult, error)
}

type configBuilder struct {
	flowCacheRepo     persistence.FlowCacheRepository
	flowRateLimitRepo persistence.FlowRateLimitRepository
	fileRepo            persistence.FileRepository
}

func NewConfigBuilder(flowCacheRepo persistence.FlowCacheRepository, flowRateLimitRepo persistence.FlowRateLimitRepository, fileRepo persistence.FileRepository) ConfigBuilder {
	return &configBuilder{
		flowCacheRepo:     flowCacheRepo,
		flowRateLimitRepo: flowRateLimitRepo,
		fileRepo:            fileRepo,
	}
}

func (b *configBuilder) BuildFlowConfig(flow persistence.Flow) (*BuildResult, error) {
	configMap := make(map[string]any)

	cacheResources, err := b.buildCacheResourcesConfig(flow.ID)
	if err != nil {
		return nil, err
	}
	if len(cacheResources) > 0 {
		configMap["cache_resources"] = cacheResources
	}

	rateLimitResources, err := b.buildRateLimitResourcesConfig(flow.ID)
	if err != nil {
		return nil, err
	}
	if len(rateLimitResources) > 0 {
		configMap["rate_limit_resources"] = rateLimitResources
	}

	input, err := b.buildInputConfig(flow)
	if err != nil {
		return nil, err
	}
	configMap["input"] = input

	if len(flow.Processors) > 0 {
		pipeline, err := b.buildProcessorPipelineConfig(flow.Processors)
		if err != nil {
			return nil, err
		}
		configMap["pipeline"] = pipeline
	}

	output, err := b.buildOutputConfig(flow)
	if err != nil {
		return nil, err
	}
	configMap["output"] = output

	fileKeys := collectFileRefs(configMap)
	var files []persistence.File
	if len(fileKeys) > 0 {
		files, err = b.fileRepo.FindByKeys(fileKeys)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch referenced files: %w", err)
		}
		if len(files) != len(fileKeys) {
			found := make(map[string]bool)
			for _, f := range files {
				found[f.Key] = true
			}
			for _, key := range fileKeys {
				if !found[key] {
					return nil, fmt.Errorf("referenced file not found: %s", key)
				}
			}
		}
		resolveFileRefs(configMap)
	}

	configYAML, err := marshalWithLiteralEnvVars(configMap)
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		Config: configYAML,
		Files:  files,
	}, nil
}

func collectFileRefs(v any) []string {
	var keys []string
	switch val := v.(type) {
	case map[string]any:
		for _, child := range val {
			keys = append(keys, collectFileRefs(child)...)
		}
	case []any:
		for _, child := range val {
			keys = append(keys, collectFileRefs(child)...)
		}
	case string:
		if strings.HasPrefix(val, FileRefPrefix) {
			key := strings.TrimPrefix(val, FileRefPrefix)
			keys = append(keys, key)
		}
	}
	return keys
}

func resolveFileRefs(v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if s, ok := child.(string); ok && strings.HasPrefix(s, FileRefPrefix) {
				key := strings.TrimPrefix(s, FileRefPrefix)
				val[k] = "file://" + filepath.Join(WorkerFilesDir, key)
			} else {
				resolveFileRefs(child)
			}
		}
	case []any:
		for i, child := range val {
			if s, ok := child.(string); ok && strings.HasPrefix(s, FileRefPrefix) {
				key := strings.TrimPrefix(s, FileRefPrefix)
				val[i] = "file://" + filepath.Join(WorkerFilesDir, key)
			} else {
				resolveFileRefs(child)
			}
		}
	}
}

func marshalWithLiteralEnvVars(configMap map[string]any) (string, error) {
	yamlBytes, err := yaml.Marshal(configMap)
	if err != nil {
		return "", err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &doc); err != nil {
		return "", err
	}

	setLiteralStyleForEnvVars(&doc)

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func setLiteralStyleForEnvVars(node *yaml.Node) {
	if node.Kind == yaml.ScalarNode && envVarPattern.MatchString(node.Value) {
		node.Style = yaml.LiteralStyle
	}
	for _, child := range node.Content {
		setLiteralStyleForEnvVars(child)
	}
}

func (b *configBuilder) buildInputConfig(flow persistence.Flow) (map[string]any, error) {
	input := make(map[string]any)

	if flow.InputComponent == "mcp_tool" {
		input["http_server"] = map[string]any{
			"path":          "/",
			"allowed_verbs": []string{"POST"},
			"timeout":       "60s",
			"sync_response": map[string]any{
				"status": `${! metadata("status_code").or("200") }`,
			},
		}
	} else {
		input[flow.InputComponent] = make(map[string]any)
		if err := yaml.Unmarshal(flow.InputConfig, input[flow.InputComponent]); err != nil {
			return nil, err
		}
	}

	input["label"] = flow.InputLabel
	return input, nil
}

func (b *configBuilder) buildOutputConfig(flow persistence.Flow) (map[string]any, error) {
	output := make(map[string]any)
	output[flow.OutputComponent] = make(map[string]any)

	if err := yaml.Unmarshal(flow.OutputConfig, output[flow.OutputComponent]); err != nil {
		return nil, err
	}

	output["label"] = flow.OutputLabel
	return output, nil
}

func (b *configBuilder) buildProcessorPipelineConfig(processors []persistence.FlowProcessor) (map[string]any, error) {
	processorList := make([]any, len(processors))

	for i, processor := range processors {
		processorConfig, err := b.buildProcessorConfig(processor)
		if err != nil {
			return nil, err
		}
		processorList[i] = processorConfig
	}

	return map[string]any{
		"processors": processorList,
	}, nil
}

func (b *configBuilder) buildProcessorConfig(processor persistence.FlowProcessor) (map[string]any, error) {
	processorConfig := make(map[string]any)

	switch processor.Component {
	case "mapping":
		processorConfig[processor.Component] = string(processor.Config)
	case "catch", "switch":
		var processorList []any
		if err := yaml.Unmarshal(processor.Config, &processorList); err != nil {
			return nil, err
		}
		processorConfig[processor.Component] = processorList
	default:
		processorConfig[processor.Component] = make(map[string]any)
		if err := yaml.Unmarshal(processor.Config, processorConfig[processor.Component]); err != nil {
			return nil, err
		}
	}

	processorConfig["label"] = processor.Label
	return processorConfig, nil
}


func (b *configBuilder) buildCacheResourcesConfig(flowID int64) ([]map[string]any, error) {
	flowCaches, err := b.flowCacheRepo.FindByFlowID(flowID)
	if err != nil {
		return nil, err
	}

	if len(flowCaches) == 0 {
		return nil, nil
	}

	cacheResources := make([]map[string]any, 0, len(flowCaches))
	for _, flowCache := range flowCaches {
		cacheResource := make(map[string]any)
		cacheResource["label"] = flowCache.Cache.Label

		cacheConfig := make(map[string]any)
		if err := yaml.Unmarshal(flowCache.Cache.Config, &cacheConfig); err != nil {
			return nil, err
		}

		cacheResource[flowCache.Cache.Component] = cacheConfig
		cacheResources = append(cacheResources, cacheResource)
	}

	return cacheResources, nil
}

func (b *configBuilder) buildRateLimitResourcesConfig(flowID int64) ([]map[string]any, error) {
	flowRateLimits, err := b.flowRateLimitRepo.FindByFlowID(flowID)
	if err != nil {
		return nil, err
	}

	if len(flowRateLimits) == 0 {
		return nil, nil
	}

	rateLimitResources := make([]map[string]any, 0, len(flowRateLimits))
	for _, flowRateLimit := range flowRateLimits {
		rateLimitResource := make(map[string]any)
		rateLimitResource["label"] = flowRateLimit.RateLimit.Label

		rateLimitConfig := make(map[string]any)
		if err := yaml.Unmarshal(flowRateLimit.RateLimit.Config, &rateLimitConfig); err != nil {
			return nil, err
		}

		rateLimitResource[flowRateLimit.RateLimit.Component] = rateLimitConfig
		rateLimitResources = append(rateLimitResources, rateLimitResource)
	}

	return rateLimitResources, nil
}
