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

:::tip
Combine the AI Gateway processor with [Mapping](/docs/components/processors/mapping) processors to pre-process data before sending to the AI, or post-process the AI response further.
:::
