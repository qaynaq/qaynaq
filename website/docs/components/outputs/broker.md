# Broker

Routes messages to multiple outputs simultaneously.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Pattern | string | `fan_out` | Routing pattern |
| Outputs | array | — | List of output configurations |

Available patterns:
- `fan_out` — Send to all outputs in parallel.
- `fan_out_sequential` — Send to all outputs in order.
- `round_robin` — Distribute across outputs.
- `greedy` — Send to the first available output.

Variants with `_fail_fast` stop on first error.
