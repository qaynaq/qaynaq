# Catch

Error handling processor. Applies a list of processors when a previous processing step fails.

| Field | Type | Description |
|-------|------|-------------|
| Processors | array | Processors to apply on error |

Add child processors inside Catch that run when upstream processing fails. Commonly used with a Mapping processor to capture the error message via `error()` and the raw content via `content()`.
