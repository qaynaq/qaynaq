---
description: Accept webhook data over HTTP and store it in a database with validation.
---

# HTTP Webhooks

This guide shows how to accept webhook data over HTTP and store it in a database.

## Use Case

You want to receive webhook events (e.g., from Stripe, GitHub, or any service) and insert them into PostgreSQL.

## Create the Flow

Open the Qaynaq UI, click **Create New Flow**, and configure each section:

### Input — select **HTTP Server**

| Field | Value |
|-------|-------|
| Path | `/webhooks/events` |
| Allowed Verbs | `POST` |
| Timeout | `10s` |

### Processor — select **Mapping**

Extract and transform the relevant fields:

| Field | Value |
|-------|-------|
| Mapping | `root.event_type = this.type` / `root.payload = this.string()` / `root.received_at = now()` |

### Output — select **SQL Insert**

| Field | Value |
|-------|-------|
| Driver | `postgres` |
| DSN | `postgres://user:pass@localhost:5432/mydb?sslmode=disable` |
| Table | `webhook_events` |
| Columns | `event_type`, `payload`, `received_at` |
| Args Mapping | `root = [this.event_type, this.payload, this.received_at]` |

Click **Save** and then **Start** the flow.

## Test It

Once the flow is running, send a test webhook. All HTTP Server flows are exposed under `/ingest/{flow-id}` followed by the configured path:

```bash
curl -X POST http://localhost:8080/ingest/{flow-id}/webhooks/events \
  -H "Content-Type: application/json" \
  -d '{"type": "payment.completed", "amount": 99.99, "currency": "USD"}'
```

Replace `{flow-id}` with the actual flow ID shown in the Qaynaq UI.

## Returning Custom Responses

To return a custom response to the webhook caller, use the **Sync Response** output with a **Broker**:

Configure the output as a **Broker** with `fan_out` pattern containing both your SQL Insert and a Sync Response output. Use a mapping processor to customize the response body before the Sync Response.

## Adding Validation

Add a **JSON Schema** processor before the mapping to reject invalid payloads. Set the schema to require a `type` field:

| Field | Value |
|-------|-------|
| Schema | `{"type": "object", "required": ["type"], "properties": {"type": {"type": "string"}}}` |

Use a **Catch** processor to handle validation errors gracefully.
