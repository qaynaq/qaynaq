# Regex Match

Splits the stream wherever a regular expression matches. Each match marks the start of a new message. Useful for parsing log formats that don't fit a fixed delimiter, like multi-line stack traces that start with a timestamp.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Pattern | string | required | Regular expression. Each match begins a new message. The pattern must use the [Go regex syntax](https://pkg.go.dev/regexp/syntax) |
| Max Buffer Size | integer | `65536` | Maximum size in bytes a single message can reach before producing an error |
