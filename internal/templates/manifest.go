package templates

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v3"
)

const (
	VariableTypeString     = "string"
	VariableTypeSecret     = "secret"
	VariableTypeConnection = "connection"

	FlowKindTool       = "tool"
	FlowKindAutomation = "automation"
)

type Manifest struct {
	ID               string     `yaml:"id"`
	Name             string     `yaml:"name"`
	Description      string     `yaml:"description"`
	Version          string     `yaml:"version"`
	MinQaynaqVersion string     `yaml:"min_qaynaq_version"`
	Variables        []Variable `yaml:"variables"`
	Flows            []Flow     `yaml:"flows"`
}

type Variable struct {
	Key         string `yaml:"key"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Placeholder string `yaml:"placeholder"`
	Default     string `yaml:"default"`
}

type Flow struct {
	Name        string      `yaml:"name"`
	Kind        string      `yaml:"kind"`
	Description string      `yaml:"description"`
	File        string      `yaml:"file"`
	Input       Component   `yaml:"input"`
	Processors  []Processor `yaml:"processors"`
	Output      Component   `yaml:"output"`
}

type Component struct {
	Component string `yaml:"component"`
	Config    any    `yaml:"config"`
}

type Processor struct {
	Label     string `yaml:"label"`
	Component string `yaml:"component"`
	Config    any    `yaml:"config"`
}

var (
	idPattern          = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	variableKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)
)

func ParseManifest(data []byte) (*Manifest, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if err := dec.Decode(new(Manifest)); err != io.EOF {
		return nil, fmt.Errorf("manifest must contain a single YAML document")
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) validate() error {
	if m.ID == "" {
		return fmt.Errorf("manifest: id is required")
	}
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("manifest %q: id must match %s", m.ID, idPattern)
	}
	if m.Name == "" {
		return fmt.Errorf("manifest %q: name is required", m.ID)
	}
	if m.Version == "" {
		return fmt.Errorf("manifest %q: version is required", m.ID)
	}
	if len(m.Flows) == 0 {
		return fmt.Errorf("manifest %q: at least one flow is required", m.ID)
	}

	varKeys := make(map[string]bool, len(m.Variables))
	for i, v := range m.Variables {
		if v.Key == "" {
			return fmt.Errorf("manifest %q: variables[%d]: key is required", m.ID, i)
		}
		if !variableKeyPattern.MatchString(v.Key) {
			return fmt.Errorf("manifest %q: variable %q: key must match %s", m.ID, v.Key, variableKeyPattern)
		}
		if varKeys[v.Key] {
			return fmt.Errorf("manifest %q: duplicate variable key %q", m.ID, v.Key)
		}
		varKeys[v.Key] = true
		switch v.Type {
		case VariableTypeString, VariableTypeSecret, VariableTypeConnection:
		default:
			return fmt.Errorf("manifest %q: variable %q: type must be one of string, secret, connection", m.ID, v.Key)
		}
		if v.Title == "" {
			return fmt.Errorf("manifest %q: variable %q: title is required", m.ID, v.Key)
		}
	}

	flowNames := make(map[string]bool, len(m.Flows))
	for i, f := range m.Flows {
		if f.File != "" {
			return fmt.Errorf("manifest %q: flows[%d]: file references are not supported yet, define the flow inline", m.ID, i)
		}
		if f.Name == "" {
			return fmt.Errorf("manifest %q: flows[%d]: name is required", m.ID, i)
		}
		if flowNames[f.Name] {
			return fmt.Errorf("manifest %q: duplicate flow name %q", m.ID, f.Name)
		}
		flowNames[f.Name] = true
		switch f.Kind {
		case FlowKindTool, FlowKindAutomation:
		default:
			return fmt.Errorf("manifest %q: flow %q: kind must be one of tool, automation", m.ID, f.Name)
		}
		if f.Input.Component == "" {
			return fmt.Errorf("manifest %q: flow %q: input.component is required", m.ID, f.Name)
		}
		if f.Output.Component == "" {
			return fmt.Errorf("manifest %q: flow %q: output.component is required", m.ID, f.Name)
		}
		for j, p := range f.Processors {
			if p.Component == "" {
				return fmt.Errorf("manifest %q: flow %q: processors[%d]: component is required", m.ID, f.Name, j)
			}
		}
		for _, key := range collectPlaceholders(f) {
			if !varKeys[key] {
				return fmt.Errorf("manifest %q: flow %q: placeholder {{ %s }} does not match any declared variable", m.ID, f.Name, key)
			}
		}
	}
	return nil
}

func collectPlaceholders(f Flow) []string {
	var keys []string
	seen := make(map[string]bool)
	collect := func(cfg any) {
		walkStrings(cfg, func(s string) string {
			for _, match := range placeholderPattern.FindAllStringSubmatch(s, -1) {
				if !seen[match[1]] {
					seen[match[1]] = true
					keys = append(keys, match[1])
				}
			}
			return s
		})
	}
	collect(f.Input.Config)
	for _, p := range f.Processors {
		collect(p.Config)
	}
	collect(f.Output.Config)
	return keys
}

// walkStrings applies fn to every string value nested in maps and slices,
// replacing the value with fn's return. Map keys are left untouched.
func walkStrings(v any, fn func(string) string) any {
	switch t := v.(type) {
	case string:
		return fn(t)
	case map[string]any:
		for k, val := range t {
			t[k] = walkStrings(val, fn)
		}
		return t
	case []any:
		for i, val := range t {
			t[i] = walkStrings(val, fn)
		}
		return t
	default:
		return v
	}
}
