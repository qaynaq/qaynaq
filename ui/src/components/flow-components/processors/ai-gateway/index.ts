import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

export const AI_PROVIDERS = ["openai", "anthropic"] as const;

const configSchema = z.object({
  provider: z.enum(AI_PROVIDERS),
  model: z.string().min(1, "Required"),
  api_key: z.string().min(1, "Required"),
  base_url: z.string(),
  system_prompt: z.string(),
  prompt: z.string().min(1, "Required"),
  args_mapping: z.string(),
  unsafe_dynamic_prompt: z.boolean(),
  max_tokens: z.number().int().min(1),
  temperature: z.number().min(0),
  mcp_tools: z.boolean(),
  mcp_url: z.string(),
  mcp_token: z.string(),
  max_tool_rounds: z.number().int().min(1),
  include_tools: z.string(),
  exclude_tools: z.string(),
});
type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  provider: "openai",
  model: "",
  api_key: "",
  base_url: "",
  system_prompt: "",
  prompt: "",
  args_mapping: "",
  unsafe_dynamic_prompt: false,
  max_tokens: 1024,
  temperature: 1.0,
  mcp_tools: true,
  mcp_url: "http://localhost:8080/mcp",
  mcp_token: "",
  max_tool_rounds: 5,
  include_tools: "",
  exclude_tools: "",
};

const component: FlowComponent<Config> = {
  id: "ai_gateway",
  name: "AI Gateway",
  category: "processor",
  description:
    "Calls an AI chat completion API and maps the response into the message.",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
