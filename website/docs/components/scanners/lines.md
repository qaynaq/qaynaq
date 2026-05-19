# Lines

Splits the byte stream into one message per line. The default scanner for line-oriented text - log files, NDJSON, plain text.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Custom Delimiter | string | `""` | Override the default newline. Use any string to mark line endings |
| Max Buffer Size | integer | `65536` | Maximum size in bytes a single line can reach before producing an error |
| Omit Empty | boolean | `false` | Skip empty lines instead of emitting empty messages |
