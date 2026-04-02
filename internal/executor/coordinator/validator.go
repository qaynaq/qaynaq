package coordinator

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pubbloblang "github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

// ValidateFlow validates each component of a stream independently using Bento's per-component
// linting, which produces detailed errors (line/column, field names) per failing section.
// Returns nil if all components are valid.
func ValidateFlow(flow persistence.Flow) error {
	b := &configBuilder{}
	var msgs []string

	inputMap, err := b.buildInputConfig(flow)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("[input/%s] %s", flow.InputComponent, err))
	} else if inputErr := lintComponentYAML(inputMap, func(s *service.StreamBuilder, y string) error {
		return s.AddInputYAML(y)
	}); inputErr != nil {
		msgs = append(msgs, fmt.Sprintf("[input/%s] %s", flow.InputComponent, inputErr))
	}

	for _, proc := range flow.Processors {
		procMap, err := b.buildProcessorConfig(proc)
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("[processor/%s] %s", proc.Label, err))
			continue
		}
		if proc.Component == "mapping" {
			if _, parseErr := pubbloblang.GlobalEnvironment().Parse(string(proc.Config)); parseErr != nil {
				msgs = append(msgs, fmt.Sprintf("[processor/%s (mapping)] %s", proc.Label, parseErr))
			}
			continue
		}
		if procErr := lintComponentYAML(procMap, func(s *service.StreamBuilder, y string) error {
			return s.AddProcessorYAML(y)
		}); procErr != nil {
			msgs = append(msgs, fmt.Sprintf("[processor/%s (%s)] %s", proc.Label, proc.Component, procErr))
		}
	}

	outputMap, err := b.buildOutputConfig(flow)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("[output/%s] %s", flow.OutputComponent, err))
	} else if outputErr := lintComponentYAML(outputMap, func(s *service.StreamBuilder, y string) error {
		return s.AddOutputYAML(y)
	}); outputErr != nil {
		msgs = append(msgs, fmt.Sprintf("[output/%s] %s", flow.OutputComponent, outputErr))
	}

	if len(msgs) > 0 {
		return fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}
	return nil
}

func lintComponentYAML(componentMap map[string]any, addFn func(*service.StreamBuilder, string) error) error {
	yamlBytes, err := yaml.Marshal(componentMap)
	if err != nil {
		return err
	}
	return addFn(service.NewStreamBuilder(), string(yamlBytes))
}
