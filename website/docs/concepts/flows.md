---
sidebar_position: 2
---

# Flows

A **flow** is the core building block in Qaynaq. It defines a data pipeline that moves data from an input, through optional processors, to an output.

## Structure

Every flow consists of three parts:

```
Input → Processors → Output
```

- **Input**: Where data comes from (Kafka, HTTP, database CDC, etc.).
- **Processors**: Optional transformations applied to each message (mapping, validation, schema decoding, etc.).
- **Output**: Where data goes (database, Kafka, HTTP endpoint, etc.).

## Creating Flows

Flows are created through the Qaynaq web UI:

1. Open the console at `http://localhost:8080`.
2. Click **Create New Flow**.
3. Give the flow a name and set the desired status.
4. Configure the input, optional processors, and output using the visual builder.
5. Save the flow.

Each component is configured through its form in the visual editor. See [Validation](./validation) for how Qaynaq checks your configuration before saving, and [Testing Flows](./testing-flows) for how to test your processor pipeline with sample data before activating.

## Flow Lifecycle

Flows have the following statuses:

| Status | Description |
|--------|-------------|
| **active** | Flow is queued to run and will be assigned to an available worker |
| **paused** | Flow is saved but not running; no worker will pick it up |
| **completed** | Flow finished processing (e.g., a finite input source was exhausted) |
| **failed** | Flow encountered an unrecoverable runtime error |

You can change a flow's status from the UI at any time. To save a flow with an invalid configuration, set the status to **paused** — Qaynaq will reject saves with status **active** if the configuration does not pass validation.

## Assignment

When a flow is started, the coordinator assigns it to an available worker. The worker then runs the flow's pipeline. If a worker goes down, the coordinator can reassign the flow to another worker.

## Buffering

Flows support buffering to handle backpressure and ensure delivery:

- **Memory buffer** — In-memory with optional spillover.
- **SQLite buffer** — Persistent buffering for at-least-once delivery.
- **System Window** — Time-based windowing (tumbling, sliding, hopping).
