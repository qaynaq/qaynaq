import { lazy } from "react";
import { z } from "zod";
import type { FlowComponent } from "../../types";
import { parseYaml, serializeYaml } from "../../utils/yaml";

const paramSchema = z.object({
  name: z.string(),
  type: z.string(),
  required: z.boolean(),
  description: z.string(),
});

const annotationsSchema = z.object({
  title: z.string().optional(),
  read_only_hint: z.boolean(),
  destructive_hint: z.boolean(),
  idempotent_hint: z.boolean(),
  open_world_hint: z.boolean(),
});

const configSchema = z.object({
  name: z.string().min(1, "Required").regex(/^[a-zA-Z0-9_-]+$/, "Only letters, numbers, _ and -"),
  description: z.string().min(1, "Required"),
  input_schema: z.array(paramSchema),
  annotations: annotationsSchema,
});

type Config = z.infer<typeof configSchema>;

const defaultConfig: Config = {
  name: "",
  description: "",
  input_schema: [],
  annotations: {
    read_only_hint: false,
    destructive_hint: false,
    idempotent_hint: false,
    open_world_hint: true,
  },
};

const component: FlowComponent<Config> = {
  id: "mcp_tool",
  name: "MCP Tool",
  category: "input",
  description:
    "Expose a flow as a tool for MCP-compatible AI clients (Claude, Cursor, etc).",
  configSchema,
  defaultConfig,
  parse: (s) => parseYaml(configSchema, s, defaultConfig),
  serialize: serializeYaml,
  Editor: lazy(() => import("./editor")),
};

export default component;
