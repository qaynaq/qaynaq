# Switch

Routes messages to different outputs based on conditions.

| Field | Type | Description |
|-------|------|-------------|
| Cases | array | List of condition/output pairs |
| Retry Until Success | boolean | Retry failed outputs |
| Strict Mode | boolean | Error if no case matches |

Each case has a **Check** condition (Bloblang expression) and an **Output** to route matching messages to.
