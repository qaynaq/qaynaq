# Kafka

Produces messages to a Kafka topic.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Addresses | array | — | Kafka broker addresses |
| Topic | string | — | Target topic |
| Key | string | — | Message key (Bloblang interpolation) |
| Compression | string | `none` | Compression: `none`, `gzip`, `snappy`, `lz4`, `zstd` |
| Max In Flight | integer | `64` | Maximum parallel produces |
| Batching | object | — | Batching policy |

Supports **SASL** authentication and **TLS**.
