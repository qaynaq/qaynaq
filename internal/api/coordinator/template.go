package coordinator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"github.com/qaynaq/qaynaq/internal/templates"
)

const templateOutputLabelTool = "mcp_tool_response"

func (c *CoordinatorAPI) ListTemplates(_ context.Context, _ *emptypb.Empty) (*pb.ListTemplatesResponse, error) {
	catalog, err := templates.Default()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load template catalog")
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.ListTemplatesResponse{}
	for _, m := range catalog.List() {
		resp.Data = append(resp.Data, c.templateToProto(m))
	}
	return resp, nil
}

func (c *CoordinatorAPI) GetTemplate(_ context.Context, in *pb.GetTemplateRequest) (*pb.TemplateResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	catalog, err := templates.Default()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load template catalog")
		return nil, status.Error(codes.Internal, err.Error())
	}
	m, ok := catalog.Get(in.GetId())
	if !ok {
		return nil, status.Error(codes.NotFound, "Template not found")
	}
	return &pb.TemplateResponse{Data: c.templateToProto(m)}, nil
}

func (c *CoordinatorAPI) InstallTemplate(ctx context.Context, in *pb.InstallTemplateRequest) (*pb.InstallTemplateResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	catalog, err := templates.Default()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load template catalog")
		return nil, status.Error(codes.Internal, err.Error())
	}
	tmpl, ok := catalog.Get(in.GetId())
	if !ok {
		return nil, status.Error(codes.NotFound, "Template not found")
	}

	values, err := c.resolveTemplateVariables(tmpl, in.GetVariables())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	selected, err := selectTemplateFlows(tmpl, in.GetFlowNames())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	installed := c.installedTemplateFlows(tmpl.ID)
	resp := &pb.InstallTemplateResponse{}
	for _, flow := range selected {
		result := &pb.InstallTemplateResponse_Result{Name: flow.Name}
		resp.Results = append(resp.Results, result)

		if installed[flow.Name] && !in.GetOverride() {
			result.Skipped = true
			continue
		}

		rendered, err := templates.RenderFlow(flow, values)
		if err != nil {
			result.Error = err.Error()
			continue
		}

		flowResp, err := c.CreateFlow(ctx, renderedFlowToProto(rendered, tmpl.ID))
		if err != nil {
			result.Error = status.Convert(err).Message()
			continue
		}
		result.Success = true
		result.FlowId = flowResp.GetData().GetId()
	}
	return resp, nil
}

func (c *CoordinatorAPI) templateToProto(m *templates.Manifest) *pb.Template {
	installed := c.installedTemplateFlows(m.ID)

	tmpl := &pb.Template{
		Id:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Version:     m.Version,
	}
	for _, v := range m.Variables {
		tmpl.Variables = append(tmpl.Variables, &pb.Template_Variable{
			Key:         v.Key,
			Title:       v.Title,
			Description: v.Description,
			Type:        v.Type,
			Required:    v.Required,
			Placeholder: v.Placeholder,
			Default:     v.Default,
		})
	}
	for _, f := range m.Flows {
		tmpl.Flows = append(tmpl.Flows, &pb.Template_FlowSummary{
			Name:        f.Name,
			Kind:        f.Kind,
			Description: f.Description,
			Installed:   installed[f.Name],
		})
	}
	return tmpl
}

func (c *CoordinatorAPI) installedTemplateFlows(templateID string) map[string]bool {
	installed := make(map[string]bool)
	flows, err := c.flowRepo.ListAllByManagedBy(templateID)
	if err != nil {
		log.Warn().Err(err).Str("template_id", templateID).Msg("Failed to list managed flows")
		return installed
	}
	for _, f := range flows {
		installed[f.Name] = true
	}
	return installed
}

func (c *CoordinatorAPI) resolveTemplateVariables(tmpl *templates.Manifest, provided map[string]string) (map[string]string, error) {
	values := make(map[string]string, len(tmpl.Variables))
	for _, v := range tmpl.Variables {
		raw := provided[v.Key]
		if raw == "" {
			raw = v.Default
		}
		if raw == "" {
			if v.Required {
				return nil, fmt.Errorf("variable %q is required", v.Key)
			}
			values[v.Key] = ""
			continue
		}

		switch v.Type {
		case templates.VariableTypeSecret:
			secret, err := c.secretRepo.GetByKey(raw)
			if err != nil || secret == nil {
				return nil, fmt.Errorf("variable %q: secret %q not found, create it first", v.Key, raw)
			}
			values[v.Key] = "${" + raw + "}"
		case templates.VariableTypeConnection:
			if err := c.ensureConnectionExists(raw); err != nil {
				return nil, fmt.Errorf("variable %q: %w", v.Key, err)
			}
			values[v.Key] = "${QAYNAQ_CONN_" + raw + "}"
		default:
			values[v.Key] = raw
		}
	}
	return values, nil
}

func (c *CoordinatorAPI) ensureConnectionExists(name string) error {
	connections, err := c.connManager.ListConnections()
	if err != nil {
		return fmt.Errorf("listing connections: %w", err)
	}
	for _, conn := range connections {
		if conn.Name == name {
			return nil
		}
	}
	return fmt.Errorf("connection %q not found, create it first", name)
}

func selectTemplateFlows(tmpl *templates.Manifest, names []string) ([]templates.Flow, error) {
	if len(names) == 0 {
		return tmpl.Flows, nil
	}
	byName := make(map[string]templates.Flow, len(tmpl.Flows))
	for _, f := range tmpl.Flows {
		byName[f.Name] = f
	}
	selected := make([]templates.Flow, 0, len(names))
	for _, name := range names {
		f, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("flow %q does not exist in template %q", name, tmpl.ID)
		}
		selected = append(selected, f)
	}
	return selected, nil
}

func renderedFlowToProto(r templates.RenderedFlow, templateID string) *pb.Flow {
	outputLabel := r.OutputComponent
	if r.Kind == templates.FlowKindTool {
		outputLabel = templateOutputLabelTool
	}

	flow := &pb.Flow{
		Name:            r.Name,
		Status:          string(persistence.FlowStatusActive),
		InputComponent:  r.InputComponent,
		InputLabel:      r.Name,
		InputConfig:     r.InputConfig,
		OutputComponent: r.OutputComponent,
		OutputLabel:     outputLabel,
		OutputConfig:    r.OutputConfig,
		IsReady:         true,
		BuilderState:    "",
		ManagedBy:       &templateID,
	}
	for _, p := range r.Processors {
		flow.Processors = append(flow.Processors, &pb.Flow_Processor{
			Label:     p.Label,
			Component: p.Component,
			Config:    p.Config,
		})
	}
	return flow
}
