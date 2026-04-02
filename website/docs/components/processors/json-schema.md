# JSON Schema

Validates messages against a JSON Schema. Messages that fail validation are rejected.

You can provide the schema in two ways — pick one:

| Field | Type | Description |
|-------|------|-------------|
| Schema File | file | A JSON Schema file uploaded via the [File Manager](/docs/concepts/files). Takes priority over Inline Schema if both are set. |
| Inline Schema | string | A JSON Schema pasted directly into the field. |

### Schema File

Select a file from the File Manager containing your JSON Schema. This is useful when the same schema is shared across multiple flows or when schemas are large.

### Inline Schema

Paste the full JSON Schema directly. For example, to require `id` and `email` fields, use a schema with `"required": ["id", "email"]` and define the property types.
