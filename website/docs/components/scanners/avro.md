# Avro

Consumes Avro Object Container Files. Each datum in the file becomes a separate message.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Avro Raw JSON | boolean | `false` | Decode into standard JSON instead of Avro JSON encoding. Standard JSON is friendlier for downstream processors that expect ordinary JSON. Avro JSON wraps union values in `{ "<type>": value }` objects, which is faithful to the spec but awkward to consume |

## Avro JSON vs raw JSON

Avro union schemas like `["null", "string", "Foo"]` encode differently depending on which mode is chosen:

| Value | Avro JSON | Standard JSON |
|-------|-----------|---------------|
| `null` | `null` | `null` |
| string `"a"` | `{"string": "a"}` | `"a"` |
| `Foo` record | `{"Foo": {...}}` | `{...}` |

Enable **Avro Raw JSON** when downstream processors expect ordinary JSON shapes.
