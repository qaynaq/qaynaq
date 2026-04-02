export interface McpToolParameter {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

export interface McpToolTemplate {
  id: string;
  name: string;
  description: string;
  parameters: McpToolParameter[];
  processor: {
    component: string;
    config: Record<string, string | number | boolean>;
  };
}

export interface SharedConfigField {
  key: string;
  title: string;
  description: string;
  type: "input" | "dynamic_select";
  required: boolean;
  default?: string;
  dataSource?: "secrets" | "connections";
  interpolationOnly?: boolean;
  placeholder?: string;
  secret?: boolean;
}

export interface TemplatePack {
  id: string;
  name: string;
  description: string;
  sharedConfig: SharedConfigField[];
  templates: McpToolTemplate[];
}

export function param(name: string, description: string, required = false, type = "string"): McpToolParameter {
  return { name, type, required, description };
}
