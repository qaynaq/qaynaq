package coordinator

import (
	"database/sql"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	_ "github.com/qaynaq/qaynaq/internal/components/all"
	"github.com/qaynaq/qaynaq/internal/persistence"
	"github.com/qaynaq/qaynaq/internal/templates"
)

func setupSecretRepo(t *testing.T) persistence.SecretRepository {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	if err := db.AutoMigrate(&persistence.Secret{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return persistence.NewSecretRepository(db)
}

func testTemplate() *templates.Manifest {
	return &templates.Manifest{
		ID: "demo",
		Variables: []templates.Variable{
			{Key: "shop_name", Type: templates.VariableTypeString, Required: true},
			{Key: "api_token", Type: templates.VariableTypeSecret, Required: true},
			{Key: "conn", Type: templates.VariableTypeConnection},
			{Key: "region", Type: templates.VariableTypeString, Default: "eu"},
		},
		Flows: []templates.Flow{
			{Name: "flow_a", Kind: templates.FlowKindTool},
			{Name: "flow_b", Kind: templates.FlowKindAutomation},
		},
	}
}

func TestResolveTemplateVariables(t *testing.T) {
	secretRepo := setupSecretRepo(t)
	if _, err := secretRepo.Create(&persistence.Secret{Key: "SHOPIFY_TOKEN", EncryptedValue: "enc"}); err != nil {
		t.Fatalf("seed secret: %v", err)
	}
	api := &CoordinatorAPI{secretRepo: secretRepo}
	tmpl := testTemplate()

	values, err := api.resolveTemplateVariables(tmpl, map[string]string{
		"shop_name": "mystore",
		"api_token": "SHOPIFY_TOKEN",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if values["shop_name"] != "mystore" {
		t.Errorf("shop_name = %q", values["shop_name"])
	}
	if values["api_token"] != "${SHOPIFY_TOKEN}" {
		t.Errorf("api_token = %q, want secret reference", values["api_token"])
	}
	if values["conn"] != "" {
		t.Errorf("conn = %q, want empty for unset optional", values["conn"])
	}
	if values["region"] != "eu" {
		t.Errorf("region = %q, want default eu", values["region"])
	}

	if _, err := api.resolveTemplateVariables(tmpl, map[string]string{"api_token": "SHOPIFY_TOKEN"}); err == nil || !strings.Contains(err.Error(), `"shop_name" is required`) {
		t.Errorf("missing required variable: err = %v", err)
	}

	if _, err := api.resolveTemplateVariables(tmpl, map[string]string{
		"shop_name": "mystore",
		"api_token": "MISSING_SECRET",
	}); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("missing secret: err = %v", err)
	}
}

func TestSelectTemplateFlows(t *testing.T) {
	tmpl := testTemplate()

	all, err := selectTemplateFlows(tmpl, nil)
	if err != nil || len(all) != 2 {
		t.Errorf("empty selection should return all flows: %d, %v", len(all), err)
	}

	one, err := selectTemplateFlows(tmpl, []string{"flow_b"})
	if err != nil || len(one) != 1 || one[0].Name != "flow_b" {
		t.Errorf("named selection failed: %v, %v", one, err)
	}

	if _, err := selectTemplateFlows(tmpl, []string{"missing"}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unknown flow name: err = %v", err)
	}
}

func TestRenderedFlowToProto(t *testing.T) {
	rendered := templates.RenderedFlow{
		Name:            "demo_tool",
		Kind:            templates.FlowKindTool,
		InputComponent:  "mcp_tool",
		InputConfig:     "name: demo_tool\n",
		OutputComponent: "sync_response",
		Processors: []templates.RenderedProcessor{
			{Label: "shopify", Component: "shopify", Config: "action: list_orders\n"},
			{Label: "error_handler", Component: "catch", Config: "- mapping: root = this\n"},
		},
	}
	flow := renderedFlowToProto(rendered, "shopify")

	if flow.GetManagedBy() != "shopify" {
		t.Errorf("managed_by = %q", flow.GetManagedBy())
	}
	if flow.GetStatus() != "active" || !flow.GetIsReady() {
		t.Errorf("status = %q, is_ready = %v, want active/true", flow.GetStatus(), flow.GetIsReady())
	}
	if flow.GetInputLabel() != "demo_tool" {
		t.Errorf("input_label = %q", flow.GetInputLabel())
	}
	if flow.GetOutputLabel() != templateOutputLabelTool {
		t.Errorf("output_label = %q, want %q for tool kind", flow.GetOutputLabel(), templateOutputLabelTool)
	}
	if len(flow.GetProcessors()) != 2 {
		t.Fatalf("processors = %d, want 2", len(flow.GetProcessors()))
	}
	if err := flow.Validate(); err != nil {
		t.Errorf("proto validation failed: %v", err)
	}

	rendered.Kind = templates.FlowKindAutomation
	rendered.OutputComponent = "http_client"
	if got := renderedFlowToProto(rendered, "shopify").GetOutputLabel(); got != "http_client" {
		t.Errorf("automation output_label = %q, want component name", got)
	}
}

// TestEmbeddedTemplatesPassBentoValidation renders every flow of every embedded
// template tmpl with representative values and runs it through the same Bento
// linting that gates flow activation. This is the port-fidelity contract: a
// template that fails here would fail at install time.
func TestEmbeddedTemplatesPassBentoValidation(t *testing.T) {
	catalog, err := templates.Default()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	for _, tmpl := range catalog.List() {
		values := make(map[string]string, len(tmpl.Variables))
		for _, v := range tmpl.Variables {
			switch v.Type {
			case templates.VariableTypeSecret:
				values[v.Key] = "${TEST_SECRET}"
			case templates.VariableTypeConnection:
				values[v.Key] = "test_connection"
			default:
				values[v.Key] = "test-value"
			}
		}

		for _, tmplFlow := range tmpl.Flows {
			rendered, err := templates.RenderFlow(tmplFlow, values)
			if err != nil {
				t.Errorf("tmpl %q flow %q: render: %v", tmpl.ID, tmplFlow.Name, err)
				continue
			}
			in := renderedFlowToProto(rendered, tmpl.ID)
			if err := in.Validate(); err != nil {
				t.Errorf("tmpl %q flow %q: proto validation: %v", tmpl.ID, tmplFlow.Name, err)
				continue
			}

			flow := persistence.Flow{Processors: make([]persistence.FlowProcessor, len(in.Processors))}
			flow.FromProto(in)
			for i, processor := range in.Processors {
				flow.Processors[i] = persistence.FlowProcessor{
					Label:     processor.GetLabel(),
					Component: processor.GetComponent(),
					Config:    []byte(processor.GetConfig()),
				}
			}
			if err := validateFlowConfig(flow); err != nil {
				t.Errorf("tmpl %q flow %q: bento validation:\n%v", tmpl.ID, tmplFlow.Name, err)
			}
		}
	}
}
