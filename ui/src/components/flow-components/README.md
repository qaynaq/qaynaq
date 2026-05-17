# Flow components

Frontend definitions for everything that appears as a node in the flow builder, or as a resource in the cache / buffer / rate-limit pages. The registry auto-discovers all components at build time via `import.meta.glob`.

To add a new component, drop a folder under the matching category (`inputs/`, `processors/`, `outputs/`, `caches/`, `buffers/`, `rate-limits/`) with an `index.ts` and an `editor.tsx`. Copy from any existing component as a template - `inputs/generate/` is a good starting point.

## Contract

```ts
export interface FlowComponent<TConfig> {
  id: string;                                    // matches the YAML key, e.g. "kafka"
  name: string;                                  // friendly label shown in pickers
  category: "input" | "processor" | "output"
          | "cache" | "buffer" | "rate_limit";
  description?: string;

  configSchema: z.ZodType<TConfig>;              // single source of truth for shape + validation
  defaultConfig: TConfig;                        // initial state for a blank form

  parse(yamlStr: string): TConfig;               // YAML -> typed config
  serialize(config: TConfig): string;            // typed config -> YAML

  toListItem?(config: TConfig): unknown;         // optional, see "Nested list semantics"
  fromListItem?(raw: unknown): TConfig;

  Editor: ComponentType<EditorProps<TConfig>>;   // React.lazy-loaded
}
```

A few rules:

- `id` must match the YAML key the backend expects.
- Export `Config` from `index.ts` so the editor can import it.
- `defaultConfig` does not have to validate. Required fields will fail validation until the user fills them in.
- Do not use Zod `.default()` - it makes input and output types diverge and breaks the generic. Initial values go in `defaultConfig`.

## Form primitives

In `@/components/form-primitives`. Every primitive supports `label`, `description`, `required`, `error`, and `size: "sm" | "default"`.

| Primitive                | Purpose                                                                 |
| ------------------------ | ----------------------------------------------------------------------- |
| `TextField`              | Single-line input. `type="password"` for secrets.                        |
| `TextAreaField`          | Multi-line input.                                                       |
| `NumberField`            | Numeric input with min/max/step.                                        |
| `CheckboxField`          | Boolean checkbox with description.                                      |
| `SwitchField`            | Boolean as a toggle switch.                                             |
| `SelectField`            | Dropdown with a fixed list of options.                                  |
| `ArrayField`             | Editable list of strings or numbers.                                    |
| `KeyValueField`          | Editable map of string keys to string values.                            |
| `CodeField`              | Monaco-backed code editor (lazy-loaded). `language="bloblang"`, `sql`, `python`, etc. |
| `ConnectionPickerField`  | Picks a secret / connection / cache / rate-limit / file. Handles the `${SECRET}` / `${QAYNAQ_CONN_X}` / `qaynaq://` wrap and unwrap. |
| `EmptyEditor`            | Reuse for components with empty config.                                 |

## Shared config helpers

In `flow-components/shared/`. When the same nested config block recurs, compose it via these. Each helper exports a Zod schema, a typed default, and an editor.

| Helper                                         | Purpose                                              |
| ---------------------------------------------- | ---------------------------------------------------- |
| `tls.ts` + `TlsEditor`                         | TLS settings.                                        |
| `retries.ts` + `RetriesEditor`                 | Retry intervals.                                     |
| `batching.ts` + `BatchingEditor`               | Batching policy.                                     |
| `auth.ts` + `BasicAuthEditor` etc.             | Basic auth, OAuth, OAuth2, JWT blocks.               |
| `ComponentListField`                           | Typed list of components from the registry (branch processors, broker outputs, etc.). |
| `OutputCasesField`                             | Switch-style output cases.                           |
| `ProcessorCasesField`                          | Switch-style processor cases.                        |

Use the `errorPathPrefix` prop on the editors to thread Zod error paths correctly when composed.

## Errors

The editor receives `errors?: Record<string, string>` keyed by dot-separated Zod paths. Pass each one to the matching primitive via its `error` prop. The host (`NodeConfigPanel` or `ResourceForm`) runs `configSchema.safeParse` on every change; you don't call Zod yourself.

## Nested list semantics

Most components nest as `{ <id>: <config-object> }`. The `mapping` processor is an exception - its nested form is a bare string. Components like that override `toListItem`/`fromListItem`:

```ts
toListItem: (c) => c.mapping ?? "",
fromListItem: (raw) => ({ mapping: typeof raw === "string" ? raw : "" }),
```

See `processors/mapping/index.ts` for the canonical example.

## Custom parse/serialize

Use `parseYaml` and `serializeYaml` from `utils/yaml.ts` for components whose YAML is a plain object. Components with bare top-level arrays (`catch`, `switch` processor) implement their own parse/serialize using `js-yaml` directly - see `processors/catch/index.ts`.

## Tests

Every registered component is automatically picked up by `__tests__/contract.test.tsx`. The contract test asserts:

1. `parse(serialize(defaultConfig))` is deeply equal to `defaultConfig`.
2. `parse("")` returns `defaultConfig`.
3. The editor renders without throwing when given `defaultConfig`.

Run `pnpm test` after adding a component.
