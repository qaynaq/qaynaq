# Kafka

Consumes messages from one or more Kafka topics.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Addresses | array | — | Kafka broker addresses |
| Topics | array | — | Topics to consume |
| Consumer Group | string | — | Consumer group ID |
| Checkpoint Limit | integer | `1024` | Max unprocessed messages before committing |
| Auto Replay Nacks | boolean | `true` | Automatically replay rejected messages |

Supports **SASL** authentication (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512) and **TLS**.
