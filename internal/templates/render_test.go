package templates

import (
	"strings"
	"testing"
	"testing/fstest"

	"gopkg.in/yaml.v3"
)

func TestRenderFlowSubstitution(t *testing.T) {
	m, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	values := map[string]string{
		"shop_name": "mystore",
		"api_token": "${SHOPIFY_API_TOKEN}",
	}
	rendered, err := RenderFlow(m.Flows[0], values)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	var cfg map[string]any
	if err := yaml.Unmarshal([]byte(rendered.Processors[0].Config), &cfg); err != nil {
		t.Fatalf("rendered processor config is not valid YAML: %v", err)
	}
	if cfg["shop_name"] != "mystore" {
		t.Errorf("shop_name = %v, want mystore", cfg["shop_name"])
	}
	if cfg["api_access_token"] != "${SHOPIFY_API_TOKEN}" {
		t.Errorf("api_access_token = %v, want secret reference", cfg["api_access_token"])
	}
	if cfg["action"] != "list_orders" {
		t.Errorf("action = %v, want list_orders", cfg["action"])
	}

	if !strings.Contains(rendered.Processors[1].Config, "root.error = error()") {
		t.Errorf("catch config lost: %q", rendered.Processors[1].Config)
	}
	var catchCfg []any
	if err := yaml.Unmarshal([]byte(rendered.Processors[1].Config), &catchCfg); err != nil {
		t.Fatalf("catch config should marshal as a list: %v", err)
	}

	if rendered.OutputConfig != "" {
		t.Errorf("nil output config should render empty, got %q", rendered.OutputConfig)
	}
}

func TestRenderFlowSpecialCharacters(t *testing.T) {
	manifest := strings.Replace(validManifest, `shop_name: "{{ shop_name }}"`, `shop_name: "prefix-{{shop_name}}-suffix"`, 1)
	m, err := ParseManifest([]byte(manifest))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	values := map[string]string{
		"shop_name": "with: colon\nand newline \"quoted\"",
		"api_token": "tok",
	}
	rendered, err := RenderFlow(m.Flows[0], values)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal([]byte(rendered.Processors[0].Config), &cfg); err != nil {
		t.Fatalf("special characters broke YAML output: %v", err)
	}
	want := "prefix-with: colon\nand newline \"quoted\"-suffix"
	if cfg["shop_name"] != want {
		t.Errorf("shop_name = %q, want %q", cfg["shop_name"], want)
	}
}

func TestRenderFlowLeavesBentoInterpolationUntouched(t *testing.T) {
	manifest := strings.Replace(validManifest, "action: list_orders", `action: list_orders
          limit: ${!this.limit.or("50")}`, 1)
	m, err := ParseManifest([]byte(manifest))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	rendered, err := RenderFlow(m.Flows[0], map[string]string{"shop_name": "s", "api_token": "t"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(rendered.Processors[0].Config, `${!this.limit.or("50")}`) {
		t.Errorf("bento interpolation was altered: %q", rendered.Processors[0].Config)
	}
}

func TestRenderFlowDoesNotMutateManifest(t *testing.T) {
	m, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := RenderFlow(m.Flows[0], map[string]string{"shop_name": "first", "api_token": "t"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	rendered, err := RenderFlow(m.Flows[0], map[string]string{"shop_name": "second", "api_token": "t"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(rendered.Processors[0].Config, "second") || strings.Contains(rendered.Processors[0].Config, "first") {
		t.Errorf("manifest was mutated by a previous render: %q", rendered.Processors[0].Config)
	}
}

func TestLoadCatalog(t *testing.T) {
	fsys := fstest.MapFS{
		"a.yaml": {Data: []byte(validManifest)},
		"b.yaml": {Data: []byte(strings.Replace(validManifest, "id: demo_pack", "id: other_pack", 1))},
	}
	c, err := LoadCatalog(fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(c.List()) != 2 {
		t.Errorf("packs = %d, want 2", len(c.List()))
	}
	if _, ok := c.Get("other_pack"); !ok {
		t.Errorf("other_pack not found")
	}

	dup := fstest.MapFS{
		"a.yaml": {Data: []byte(validManifest)},
		"b.yaml": {Data: []byte(validManifest)},
	}
	if _, err := LoadCatalog(dup); err == nil || !strings.Contains(err.Error(), "duplicate template id") {
		t.Errorf("expected duplicate template id error, got %v", err)
	}
}

func TestDefaultCatalogLoads(t *testing.T) {
	c, err := Default()
	if err != nil {
		t.Fatalf("embedded catalog failed to load: %v", err)
	}

	wantFlows := map[string]int{
		"google_calendar": 13,
		"google_drive":    22,
		"google_sheets":   28,
		"shopify":         7,
	}
	if len(c.List()) != len(wantFlows) {
		t.Errorf("packs = %d, want %d", len(c.List()), len(wantFlows))
	}
	for id, want := range wantFlows {
		m, ok := c.Get(id)
		if !ok {
			t.Errorf("pack %q not found", id)
			continue
		}
		if len(m.Flows) != want {
			t.Errorf("pack %q: flows = %d, want %d", id, len(m.Flows), want)
		}
		for _, f := range m.Flows {
			if f.Kind != FlowKindTool {
				t.Errorf("pack %q flow %q: kind = %q, want tool", id, f.Name, f.Kind)
			}
		}
	}
}
