import {
  TextField,
  NumberField,
  SelectField,
  CheckboxField,
  CodeField,
} from "@/components/form-primitives";
import { AI_PROVIDERS } from ".";
import type { EditorProps } from "../../types";

interface Config {
  provider: (typeof AI_PROVIDERS)[number];
  model: string;
  api_key: string;
  base_url: string;
  system_prompt: string;
  prompt: string;
  args_mapping: string;
  unsafe_dynamic_prompt: boolean;
  max_tokens: number;
  temperature: number;
  mcp_tools: boolean;
  mcp_url: string;
  mcp_token: string;
  max_tool_rounds: number;
  include_tools: string;
  exclude_tools: string;
}

export default function AiGatewayEditor({
  value,
  onChange,
  errors,
}: EditorProps<Config>) {
  const set = <K extends keyof Config>(k: K, v: Config[K]) =>
    onChange({ ...value, [k]: v });
  return (
    <div className="space-y-4">
      <SelectField
        label="Provider"
        required
        value={value.provider}
        onChange={(v) => set("provider", v as Config["provider"])}
        options={AI_PROVIDERS as unknown as string[]}
      />
      <TextField
        label="Model"
        description="Model identifier (e.g. gpt-4o, claude-sonnet-4-6)."
        required
        value={value.model}
        onChange={(v) => set("model", v)}
        error={errors?.model}
      />
      <TextField
        label="API Key"
        type="password"
        required
        value={value.api_key}
        onChange={(v) => set("api_key", v)}
        error={errors?.api_key}
      />
      <TextField
        label="Base URL"
        description="Custom API base URL. Leave empty to use the provider's default."
        value={value.base_url}
        onChange={(v) => set("base_url", v)}
      />
      <CodeField
        label="System Prompt"
        description="An optional system prompt to set the behavior of the AI model."
        value={value.system_prompt}
        onChange={(v) => set("system_prompt", v)}
      />
      <CodeField
        label="Prompt"
        description="The user prompt template. Use %v placeholders for values provided by args_mapping."
        required
        value={value.prompt}
        onChange={(v) => set("prompt", v)}
        error={errors?.prompt}
      />
      <CodeField
        label="Args Mapping"
        description="A Bloblang mapping which should evaluate to an array of values matching the %v placeholders."
        value={value.args_mapping}
        onChange={(v) => set("args_mapping", v)}
      />
      <CheckboxField
        label="Unsafe Dynamic Prompt"
        description="Enable interpolation in prompt and system_prompt."
        checked={value.unsafe_dynamic_prompt}
        onChange={(c) => set("unsafe_dynamic_prompt", c)}
      />
      <NumberField
        label="Max Tokens"
        min={1}
        value={value.max_tokens}
        onChange={(v) => set("max_tokens", v)}
      />
      <NumberField
        label="Temperature"
        min={0}
        step={0.1}
        value={value.temperature}
        onChange={(v) => set("temperature", v)}
      />
      <section className="space-y-3 border-t pt-3">
        <h4 className="text-sm font-medium">MCP Tools</h4>
        <CheckboxField
          label="Enabled"
          description="Let the model discover and call MCP tools registered in Qaynaq."
          checked={value.mcp_tools}
          onChange={(c) => set("mcp_tools", c)}
        />
        <TextField
          label="MCP URL"
          size="sm"
          value={value.mcp_url}
          onChange={(v) => set("mcp_url", v)}
        />
        <TextField
          label="MCP Token"
          size="sm"
          type="password"
          value={value.mcp_token}
          onChange={(v) => set("mcp_token", v)}
        />
        <NumberField
          label="Max Tool Rounds"
          size="sm"
          min={1}
          value={value.max_tool_rounds}
          onChange={(v) => set("max_tool_rounds", v)}
        />
        <TextField
          label="Include Tools"
          description="Comma-separated allowlist of MCP tool names."
          size="sm"
          value={value.include_tools}
          onChange={(v) => set("include_tools", v)}
        />
        <TextField
          label="Exclude Tools"
          description="Comma-separated blocklist of MCP tool names."
          size="sm"
          value={value.exclude_tools}
          onChange={(v) => set("exclude_tools", v)}
        />
      </section>
    </div>
  );
}
