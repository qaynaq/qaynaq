# Mapping

Applies [Bloblang](https://www.benthos.dev/docs/guides/bloblang/about) transformations to messages. This is the most commonly used processor.

| Field | Type | Description |
|-------|------|-------------|
| Mapping | Bloblang | The Bloblang mapping to apply (required) |

Common patterns:

- **Rename and transform fields** — `root.full_name = this.first_name + " " + this.last_name`
- **Convert values** — `root.email = this.email.lowercase()`
- **Add timestamps** — `root.created_at = now()`
- **Filter messages** — `root = if this.status == "active" { this } else { deleted() }`
