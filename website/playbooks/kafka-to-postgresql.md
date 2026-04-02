---
description: Set up a flowing pipeline from Kafka to PostgreSQL with Avro schema decoding.
---

# Kafka to PostgreSQL

This guide walks through setting up a complete flowing pipeline from Kafka to PostgreSQL with Avro schema decoding.

## Prerequisites

- Docker installed
- Qaynaq coordinator and worker running ([Installation](/docs/getting-started/installation))

## 1. Start the Development Environment

Create a `docker-compose.yml` with Redpanda (Kafka-compatible), Schema Registry, and PostgreSQL:

```yaml
services:
  redpanda:
    image: docker.redpanda.com/redpandadata/redpanda:v24.3.9
    command:
      - redpanda start
      - --smp 1
      - --overprovisioned
      - --node-id 0
      - --kafka-addr PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:9092
      - --advertise-kafka-addr PLAINTEXT://redpanda:29092,OUTSIDE://localhost:9092
      - --pandaproxy-addr 0.0.0.0:8082
      - --advertise-pandaproxy-addr localhost:8082
    ports:
      - "8081:8081"
      - "8082:8082"
      - "9092:9092"
      - "29092:29092"

  postgres:
    image: postgres:17.2-alpine
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
```

```bash
docker compose up -d
```

This starts:
- **Redpanda** (Kafka-compatible) on `localhost:9092`
- **Schema Registry** on `localhost:8081`
- **PostgreSQL** on `localhost:5432`

## 2. Create the PostgreSQL Table

```bash
docker exec -it postgres psql -U postgres -d mydb
```

```sql
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 3. Create a Kafka Topic

```bash
docker exec -it redpanda rpk topic create application-events --partitions 1
```

## 4. Register an Avro Schema

```bash
curl -X POST http://localhost:8081/subjects/application-events-value/versions \
  -H "Content-Type: application/vnd.schemaregistry.v1+json" \
  -d '{
    "schema": "{\"type\":\"record\",\"name\":\"Event\",\"fields\":[{\"name\":\"event_type\",\"type\":\"string\"},{\"name\":\"user_id\",\"type\":\"string\"},{\"name\":\"message\",\"type\":\"string\"}]}"
  }'
```

## 5. Create the Qaynaq Flow

Open the Qaynaq UI, click **Create New Flow**, and configure each section:

### Input — select **Kafka**

| Field | Value |
|-------|-------|
| Addresses | `localhost:9092` |
| Topics | `application-events` |
| Consumer Group | `event-flow-group` |

### Processor — select **Schema Registry Decode**

| Field | Value |
|-------|-------|
| URL | `http://localhost:8081` |

### Output — select **SQL Insert**

| Field | Value |
|-------|-------|
| Driver | `postgres` |
| DSN | `postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable` |
| Table | `events` |
| Columns | `event_type`, `user_id`, `message` |
| Args Mapping | `root = [this.event_type, this.user_id, this.message]` |

Click **Save** and then **Start** the flow.

## 6. Send Test Data

The example below uses Python, but any Kafka producer that writes Confluent-framed Avro will work — use whatever language or tool fits your stack.

Install the Python dependencies:

```bash
pip install confluent-kafka fastavro requests
```

Then run a quick test producer:

```python
import io, struct, json, random, requests
from confluent_kafka import Producer
from fastavro import parse_schema, schemaless_writer

SCHEMA_REGISTRY_URL = "http://localhost:8081"
TOPIC = "application-events"

# Fetch schema
resp = requests.get(f"{SCHEMA_REGISTRY_URL}/subjects/{TOPIC}-value/versions/latest")
schema_id = resp.json()["id"]
parsed = parse_schema(json.loads(resp.json()["schema"]))

# Produce sample events
producer = Producer({"bootstrap.servers": "localhost:9092"})
for event in [
    {"event_type": "user_signup",        "user_id": "user_42",  "message": "New premium signup"},
    {"event_type": "button_click",       "user_id": "user_99",  "message": "Clicked checkout on cart page"},
    {"event_type": "purchase_completed", "user_id": "user_77",  "message": "Order #12345 — $49.99 USD"},
]:
    buf = io.BytesIO()
    buf.write(struct.pack("b", 0))          # magic byte
    buf.write(struct.pack(">I", schema_id)) # schema ID
    schemaless_writer(buf, parsed, event)
    producer.produce(TOPIC, value=buf.getvalue())

producer.flush()
print("Done - 3 events sent.")
```

## 7. Verify

```bash
docker exec -it postgres psql -U postgres -d mydb \
  -c "SELECT * FROM events ORDER BY created_at DESC;"
```

## Cleanup

```bash
docker compose down -v
```
