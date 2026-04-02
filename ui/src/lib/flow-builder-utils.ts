import * as yaml from "js-yaml";
import type { McpToolTemplate, SharedConfigField } from "./mcp-tool-templates";

export function buildFlowFromTemplate(
  template: McpToolTemplate,
  sharedConfig: Record<string, string>,
  sharedConfigFields: SharedConfigField[],
  packId: string,
) {
  const inputSchema = template.parameters.map((p) => ({
    name: p.name,
    type: p.type,
    required: p.required,
    description: p.description,
  }));

  const mcpInputConfig: Record<string, any> = {
    name: template.name,
    description: template.description,
  };
  if (inputSchema.length > 0) {
    mcpInputConfig.input_schema = inputSchema;
  }

  const interpolationKeys = new Set(
    sharedConfigFields.filter((f) => f.interpolationOnly).map((f) => f.key),
  );

  const processorConfig: Record<string, any> = {};
  for (const [key, value] of Object.entries(template.processor.config)) {
    if (typeof value === "string") {
      let resolved = value;
      for (const [ik, iv] of Object.entries(sharedConfig)) {
        if (interpolationKeys.has(ik)) {
          resolved = resolved.split(`{{${ik}}}`).join(iv);
        }
      }
      processorConfig[key] = resolved;
    } else {
      processorConfig[key] = value;
    }
  }

  for (const [key, value] of Object.entries(sharedConfig)) {
    if (value && !interpolationKeys.has(key)) {
      processorConfig[key] = value;
    }
  }

  return {
    name: template.name,
    status: "active",
    input_component: "mcp_tool",
    input_label: template.name,
    input_config: yaml.dump(mcpInputConfig, { lineWidth: -1, noRefs: true }),
    output_component: "sync_response",
    output_label: "mcp_tool_response",
    output_config: "",
    is_ready: true,
    builder_state: "",
    managed_by: packId,
    processors: [
      {
        label: template.processor.component,
        component: template.processor.component,
        config: yaml.dump(processorConfig, { lineWidth: -1, noRefs: true }),
      },
      {
        label: "error_handler",
        component: "catch",
        config: yaml.dump(
          [
            {
              mapping:
                'meta status_code = error().re_find_all("^\\\\[[0-9]+\\\\]").index(0).trim("[]").or("500")\nroot.error = error().re_replace_all("^\\\\[[0-9]+\\\\] ", "")',
            },
          ],
          { lineWidth: -1, noRefs: true },
        ),
      },
    ],
  };
}
