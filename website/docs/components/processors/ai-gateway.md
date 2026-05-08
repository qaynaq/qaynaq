# AI Gateway

Calls an AI chat completion API and maps the response into the message.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Provider | select | — | The AI provider: `openai` or `anthropic` (required) |
| Model | string | — | Model identifier, e.g. `gpt-4o`, `claude-sonnet-4-6` (required) |
| API Key | string (secret) | — | API key for the provider (required) |
| Base URL | string | — | Custom API endpoint. When empty, uses the provider's default |
| System Prompt | string | — | Optional system message to set model behavior |
| Prompt | string | — | User prompt template with `%v` placeholders (required) |
| Args Mapping | bloblang | — | Bloblang mapping that evaluates to an array of values for `%v` placeholders |
| Unsafe Dynamic Prompt | boolean | `false` | Enables `${!this.field}` interpolation in prompt fields |
| Max Tokens | integer | `1024` | Maximum tokens to generate |
| Temperature | float | `1.0` | Sampling temperature (higher = more random) |
| MCP Tools | boolean | `true` | Discover and call MCP tools registered in Qaynaq |
| MCP URL | string | `http://localhost:8080/mcp` | Qaynaq MCP endpoint for tool discovery and execution. Supports `${ENV_VAR}` substitution |
| MCP Token | string (secret) | — | Bearer token for the Qaynaq MCP endpoint when auth is enabled. Supports `${ENV_VAR}` substitution |
| Max Tool Rounds | integer | `5` | Maximum tool-calling rounds before forcing a final response |
| Include Tools | string | — | Comma-separated allowlist of MCP tool names; when set, only these tools are exposed |
| Exclude Tools | string | — | Comma-separated blocklist of MCP tool names to hide from the model |
| Result Map | bloblang | — | Mapping to apply the AI response to the original message (required) |

## Providers

**OpenAI** — Connects to the OpenAI Chat Completions API. Supports all OpenAI-compatible endpoints via the Base URL field (e.g. Azure OpenAI, local models with OpenAI-compatible APIs).

**Anthropic** — Connects to the Anthropic Messages API. Uses API version `2023-06-01`.

## Prompt Template

The Prompt field is a string template. Use `%v` placeholders that are substituted at runtime with values from the Args Mapping field. The Args Mapping is a Bloblang mapping that must evaluate to an array of values, one for each `%v` placeholder in the prompt.

For example, a prompt of `Summarize the text about %v: %v` with an args_mapping of `root = [this.topic, this.content]` will substitute `this.topic` for the first `%v` and `this.content` for the second.

## Dynamic Prompt (Advanced)

When Unsafe Dynamic Prompt is enabled, the Prompt and System Prompt fields support Bento interpolation functions. Use `${!this.field_name}` to inject message fields directly into the prompt string. Both interpolation and args_mapping can be used together — interpolation resolves first, then `%v` placeholders are substituted.

## Result Mapping

The Result Map is a Bloblang mapping where:
- `this` refers to the AI response object
- `root` refers to the original message (fields are preserved)

The AI response object contains:
- **content** — The generated text response
- **model** — The model that was used
- **finish_reason** — Why generation stopped (e.g. `stop`, `end_turn`, `length`)
- **usage.input_tokens** — Number of input tokens consumed
- **usage.output_tokens** — Number of output tokens generated

## MCP Tools

When MCP Tools is enabled the processor lists every tool registered with the Qaynaq MCP endpoint and exposes it to the model. The model can issue tool calls across up to Max Tool Rounds rounds; results are fed back as `tool` messages and the model is re-prompted.

If your Qaynaq instance protects the MCP endpoint with an access token, set MCP Token to a valid token. The processor sends it as `Authorization: Bearer <token>` on every MCP request. Both MCP URL and MCP Token accept `${ENV_VAR}` references so you can keep the values out of your YAML, e.g. `mcp_token: ${QAYNAQ_MCP_TOKEN}`.

Use Include Tools to limit the surface area to a known set, or Exclude Tools to drop tools you do not want exposed. Names are comma-separated and matched exactly. When Include Tools is non-empty only listed tools are exposed; everything else is filtered out. When Include Tools is empty every discovered tool is allowed except those listed in Exclude Tools. A name appearing in both lists is dropped.

### Avoiding Self-Recursion

When this processor lives inside an `mcp_tool` flow, the MCP endpoint exposes the flow's own tool back to the model. If the model picks that tool, the worker forwards the call into itself, the inner request blocks waiting for the outer pipeline to finish, and the call eventually fails with `408 Request timed out`.

Add the flow's own tool name to Exclude Tools to prevent this. For example, if your `mcp_tool` flow defines `name: top_customer_details`, set `exclude_tools: top_customer_details` on the AI Gateway processor inside that flow.

:::tip
Combine the AI Gateway processor with [Mapping](/docs/components/processors/mapping) processors to pre-process data before sending to the AI, or post-process the AI response further.
:::
