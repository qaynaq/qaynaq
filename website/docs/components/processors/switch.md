# Switch

Conditional processing — routes messages to different processors based on conditions.

| Field | Type | Description |
|-------|------|-------------|
| Cases | array | List of condition/processor pairs |

Each case has:
- **Check** — A Bloblang condition (e.g., `this.type == "order"`).
- **Processors** — Processors to apply when the condition is true.

Add multiple cases in the UI. Messages are evaluated against each case in order.
