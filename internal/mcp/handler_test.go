package mcp

import (
	"strings"
	"testing"
)

func TestParseToolConfig_NoAnnotations(t *testing.T) {
	yamlCfg := []byte(`
name: get_weather
description: Look up the weather
input_schema:
  - name: city
    type: string
    required: true
`)

	cfg, err := parseToolConfig(yamlCfg)
	if err != nil {
		t.Fatalf("parseToolConfig: %v", err)
	}
	if cfg.Name != "get_weather" {
		t.Errorf("name: got %q", cfg.Name)
	}
	if cfg.Annotations.Title != "" {
		t.Errorf("title should be empty, got %q", cfg.Annotations.Title)
	}
	if cfg.Annotations.ReadOnlyHint != nil ||
		cfg.Annotations.DestructiveHint != nil ||
		cfg.Annotations.IdempotentHint != nil ||
		cfg.Annotations.OpenWorldHint != nil {
		t.Errorf("hints should be nil when annotations block is absent")
	}
}

func TestParseToolConfig_WithAnnotations(t *testing.T) {
	yamlCfg := []byte(`
name: get_weather
description: Look up the weather
input_schema: []
annotations:
  title: Weather Lookup
  read_only_hint: true
  destructive_hint: false
  idempotent_hint: true
  open_world_hint: true
`)

	cfg, err := parseToolConfig(yamlCfg)
	if err != nil {
		t.Fatalf("parseToolConfig: %v", err)
	}
	if cfg.Annotations.Title != "Weather Lookup" {
		t.Errorf("title: got %q", cfg.Annotations.Title)
	}
	checks := []struct {
		name string
		got  *bool
		want bool
	}{
		{"read_only_hint", cfg.Annotations.ReadOnlyHint, true},
		{"destructive_hint", cfg.Annotations.DestructiveHint, false},
		{"idempotent_hint", cfg.Annotations.IdempotentHint, true},
		{"open_world_hint", cfg.Annotations.OpenWorldHint, true},
	}
	for _, c := range checks {
		if c.got == nil {
			t.Errorf("%s: expected non-nil pointer", c.name)
			continue
		}
		if *c.got != c.want {
			t.Errorf("%s: got %v want %v", c.name, *c.got, c.want)
		}
	}
}

func TestParseToolConfig_PartialAnnotations(t *testing.T) {
	yamlCfg := []byte(`
name: search
annotations:
  read_only_hint: true
`)

	cfg, err := parseToolConfig(yamlCfg)
	if err != nil {
		t.Fatalf("parseToolConfig: %v", err)
	}
	if cfg.Annotations.ReadOnlyHint == nil || !*cfg.Annotations.ReadOnlyHint {
		t.Errorf("read_only_hint should be *true")
	}
	if cfg.Annotations.DestructiveHint != nil {
		t.Errorf("destructive_hint should remain nil when omitted")
	}
}

func TestParseToolConfig_NonBoolHintIgnored(t *testing.T) {
	yamlCfg := []byte(`
name: search
annotations:
  read_only_hint: "yes"
`)

	cfg, err := parseToolConfig(yamlCfg)
	if err != nil {
		t.Fatalf("parseToolConfig: %v", err)
	}
	if cfg.Annotations.ReadOnlyHint != nil {
		t.Errorf("non-bool value should produce nil pointer")
	}
}

func TestParseToolConfig_RejectsMalformedYAML(t *testing.T) {
	_, err := parseToolConfig([]byte("name: [unterminated"))
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got %v", err)
	}
}
