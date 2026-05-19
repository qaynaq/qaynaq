# XML Documents

Consumes a stream of XML documents. Each document becomes a separate message, converted to its JSON equivalent.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Cast | boolean | `false` | Try to cast numeric and boolean string values into their native types. Off means every value stays a string |
