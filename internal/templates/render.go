package templates

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type RenderedFlow struct {
	Name            string
	Kind            string
	InputComponent  string
	InputConfig     string
	Processors      []RenderedProcessor
	OutputComponent string
	OutputConfig    string
}

type RenderedProcessor struct {
	Label     string
	Component string
	Config    string
}

// RenderFlow substitutes {{ key }} placeholders with values and marshals each
// config section to its YAML string form, matching what a user would have
// entered in the flow builder.
func RenderFlow(f Flow, values map[string]string) (RenderedFlow, error) {
	substitute := func(cfg any) any {
		return walkStrings(deepCopy(cfg), func(s string) string {
			return placeholderPattern.ReplaceAllStringFunc(s, func(match string) string {
				key := placeholderPattern.FindStringSubmatch(match)[1]
				if v, ok := values[key]; ok {
					return v
				}
				return match
			})
		})
	}

	inputConfig, err := marshalConfig(substitute(f.Input.Config))
	if err != nil {
		return RenderedFlow{}, fmt.Errorf("flow %q: input config: %w", f.Name, err)
	}
	outputConfig, err := marshalConfig(substitute(f.Output.Config))
	if err != nil {
		return RenderedFlow{}, fmt.Errorf("flow %q: output config: %w", f.Name, err)
	}

	rendered := RenderedFlow{
		Name:            f.Name,
		Kind:            f.Kind,
		InputComponent:  f.Input.Component,
		InputConfig:     inputConfig,
		OutputComponent: f.Output.Component,
		OutputConfig:    outputConfig,
	}
	for _, p := range f.Processors {
		cfg, err := marshalConfig(substitute(p.Config))
		if err != nil {
			return RenderedFlow{}, fmt.Errorf("flow %q: processor %q config: %w", f.Name, p.Label, err)
		}
		label := p.Label
		if label == "" {
			label = p.Component
		}
		rendered.Processors = append(rendered.Processors, RenderedProcessor{
			Label:     label,
			Component: p.Component,
			Config:    cfg,
		})
	}
	return rendered, nil
}

func marshalConfig(cfg any) (string, error) {
	if cfg == nil {
		return "", nil
	}
	if m, ok := cfg.(map[string]any); ok && len(m) == 0 {
		return "", nil
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n") + "\n", nil
}

func deepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		c := make(map[string]any, len(t))
		for k, val := range t {
			c[k] = deepCopy(val)
		}
		return c
	case []any:
		c := make([]any, len(t))
		for i, val := range t {
			c[i] = deepCopy(val)
		}
		return c
	default:
		return v
	}
}
