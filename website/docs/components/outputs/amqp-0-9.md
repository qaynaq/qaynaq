# AMQP 0.9

Publishes messages to an AMQP 0.9 exchange (e.g. RabbitMQ).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URLs | array | — | List of broker URLs to connect to. The first successful connection is used |
| Exchange | string | — | Name of the AMQP exchange to publish to. Supports interpolation functions |
| Routing Key | string | `""` | Binding key attached to each message. Supports interpolation functions |
| Type | string | `""` | Type property attached to each message. Supports interpolation functions |
| Content Type | string | `application/octet-flow` | Content type attribute for published messages |
| Persistent | boolean | `false` | Send messages with persistent delivery mode so they survive broker restarts |
| Max In Flight | integer | `64` | Maximum number of messages to have in flight at a given time |
| Exchange Declare — Enabled | boolean | `false` | Declare the exchange during connection if it does not exist |
| Exchange Declare — Type | string | `direct` | Type of exchange to declare: `direct`, `fanout`, `topic`, or `x-custom` |
| Exchange Declare — Durable | boolean | `true` | Whether the declared exchange survives broker restarts |
