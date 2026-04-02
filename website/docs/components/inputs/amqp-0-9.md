# AMQP 0.9

Consumes messages from an AMQP 0.9 queue (e.g. RabbitMQ).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URLs | array | — | List of broker URLs to connect to. The first successful connection is used |
| Queue | string | — | Name of the AMQP queue to consume from |
| Consumer Tag | string | `""` | Identifier for the consumer connection |
| Auto Ack | boolean | `false` | Acknowledge messages on receipt, skipping downstream acknowledgment |
| Prefetch Count | integer | `10` | Maximum number of pending messages to have in-flight at a time |
| Nack Reject Patterns | array | `[]` | Regular expressions matched against processing errors. Messages whose errors match are dropped instead of requeued |
| Queue Declare — Enabled | boolean | `false` | Declare the queue during connection if it does not exist |
| Queue Declare — Durable | boolean | `true` | Whether the declared queue survives broker restarts |
| Queue Declare — Auto Delete | boolean | `false` | Whether the declared queue is deleted when the last consumer disconnects |
