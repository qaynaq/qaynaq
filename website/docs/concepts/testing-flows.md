---
sidebar_position: 6
---

# Testing Flows

The **Try** feature lets you test your processor pipeline with sample messages before activating a flow. You can verify that a given input produces the expected output without deploying the flow to a worker or triggering side effects on external systems.

## How It Works

When you click the **Try** button in the flow builder:

1. You provide one or more test messages as input.
2. Qaynaq builds a temporary, in-memory flow using only your **processors** — no real input or output connections are made.
3. Each test message is pushed through the processor pipeline.
4. The transformed results are returned and displayed in the dialog.

The temporary flow is discarded immediately after the test completes.

## What Is Tested

The Try feature exercises the **processor chain** of your flow. This includes:

- **Mapping** processors — Bloblang transformations are executed against the test messages.
- **JSON Schema** validation — messages are validated against the configured schema, including file-based schemas from the File Manager.
- **Catch** processors — error handling logic is triggered when an upstream processor fails.
- **Switch** processors — routing conditions are evaluated and the matching branch is executed.
- **AI Gateway** and other processors that call external services — these run live, using your configured secrets.

## What Is Not Tested

| Included | Not included |
|----------|--------------|
| All configured processors | Input component (Kafka, HTTP, CDC, etc.) |
| Secret resolution (`${SECRET_KEY}` syntax) | Output component (database, Kafka, etc.) |
| File references from the File Manager | Buffer configuration |
| External API calls made by processors | Worker assignment and scheduling |

Input and output components are excluded because inputs require real connections to receive data, and outputs would produce side effects on external systems.

## Secrets and File References

The Try feature resolves secrets and file references the same way a running flow does:

- **Secrets**: Any `${SECRET_KEY}` references in processor configuration are resolved from the secrets store. If a secret is missing, the processor will fail with an appropriate error.
- **File references**: Components that reference files from the File Manager (e.g., JSON Schema with a schema file) have their files fetched from the database and made available during the test.

## Test Messages

You can provide multiple test messages in a single test run. Each message is pushed through the pipeline independently, and the results are displayed in order.

Messages should be formatted as the processor chain expects to receive them — typically JSON for pipelines that parse structured data, or plain text for processors that operate on raw strings.

## Timeout

Each test run has a **10-second timeout**. If all processors complete before the timeout, results are returned immediately. If processors filter or delete messages (e.g., a `mapping` processor that calls `root = deleted()`), the test waits briefly for any remaining output before returning.

## Try Button

The **Try** button appears in the flow builder toolbar between the **Validate** and **Save** buttons. It is enabled whenever at least one processor is configured. You do not need to configure an input or output to use it.

:::tip
Use **Validate** to check that your configuration is syntactically correct, then use **Try** to confirm that your processors produce the expected output for real data.
:::
