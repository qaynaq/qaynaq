# Schema Registry Decode

Decodes Avro-encoded messages using a Confluent Schema Registry.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | Schema Registry URL (required) |
| Avro Raw JSON | boolean | `false` | Output raw JSON instead of Avro logical types |

Supports authentication: **Basic Auth**, **OAuth**, and **JWT**.

:::tip
This processor is commonly paired with the [Kafka](/docs/components/inputs/kafka) input for Avro-encoded topic consumption. See the [Kafka to PostgreSQL](/playbooks/kafka-to-postgresql) playbook for a complete example.
:::
