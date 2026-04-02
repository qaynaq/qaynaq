---
sidebar_position: 5
---

# Flow Validation

Qaynaq validates flow configurations before saving them as **active**. This prevents broken flows from being dispatched to workers, and gives you immediate, per-component error feedback so you can fix problems before they reach production.

## How Validation Works

When you save a flow with status **active**, the coordinator validates every component in the pipeline:

1. **Input** — checked against Bento's component schema (field names, types, required fields).
2. **Processors** — each processor is checked individually:
   - For **mapping** processors, the Bloblang script is compiled and any syntax errors are reported.
   - For all other processors, the component schema is validated (unknown fields, missing required fields, etc.).
3. **Output** — checked against Bento's component schema.

Validation runs per-component, so errors in multiple components are all reported at once rather than stopping at the first failure.

## Error Format

Errors are labeled by component position and type so you can locate the problem immediately:

```
[input/http_server] lint errors: (3,5) unknown field 'bad_field'
[processor/transform (mapping)] expected assignment, got end of input
[output/kafka] lint errors: (2,3) field 'addresses' is required
```

Each line identifies:
- The component role (`input`, `processor`, or `output`)
- The label you assigned to the component
- For processors, the component type in parentheses
- The line and column of the error where applicable

## Saving with Active Status

If the configuration is invalid and you attempt to save with status **active**, the save is **rejected** — nothing is written to the database. The full validation error is returned so you can correct it.

:::tip
Use the **Validate** button in the flow builder to check your configuration at any time before saving. This runs the same validation without attempting to save. To also verify that your processors produce the expected output, see [Testing Flows](./testing-flows).
:::

## Saving an Invalid Configuration

If you want to save a flow that is not yet fully configured or has known issues, set the status to **paused**. A paused flow is saved to the database but is never dispatched to a worker, so it cannot cause runtime failures.

Once you have fixed the configuration, change the status to **active** and save again to put the flow back into the work queue.

## Validate Button

The flow builder includes a **Validate** button in the toolbar. Clicking it sends the current configuration to the coordinator for validation without saving. Results appear inline:

- A green confirmation message if all components pass.
- A red error panel listing every component-level error if any fail.

The result is cleared automatically whenever you change the flow name, status, or any node configuration.

## What Is and Is Not Validated

| Checked at validation time | Not checked at validation time |
|---------------------------|-------------------------------|
| Unknown or misspelled field names | Runtime connectivity (broker reachability, DB access) |
| Missing required fields | Secret values being present or correct |
| Bloblang syntax errors in mapping processors | File references existing on disk |
| Invalid field types | Permission errors on external systems |

Runtime errors that only manifest when a worker executes the flow — such as a Kafka broker being unreachable — are captured as flow events and surfaced in the flow detail view.
