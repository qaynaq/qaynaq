# Chunker

Splits the byte stream into fixed-size chunks. Each chunk becomes one message.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Size | integer | required | Number of bytes per chunk. The final chunk may be smaller if the stream does not divide evenly |

Use this for binary data that has no logical structure, or when you want to bound message size for downstream processing.
