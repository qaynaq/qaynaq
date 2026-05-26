---
description: Build a RAG knowledge base over your documents and expose it to AI assistants as an MCP tool.
---

# RAG Knowledge Base for AI Assistants

This playbook builds a retrieval-augmented generation (RAG) pipeline end-to-end: chunk your documents, embed each chunk, store the embeddings in PostgreSQL with [pgvector](https://github.com/pgvector/pgvector), and expose the result as a tool that any MCP-compatible AI assistant can call.

Two flows, both built in the Qaynaq UI:

1. **Ingest flow** - reads files, chunks them, embeds each chunk, writes rows to Postgres.
2. **Query flow** - exposed as an MCP tool. Embeds the incoming query, runs a cosine similarity search against the same table, returns the top matches.

For MCP client setup and authentication, see the [MCP Server guide](/docs/guides/mcp-server).

## Prerequisites

- Docker installed
- Qaynaq coordinator and worker running ([Installation](/docs/getting-started/installation))

## 1. Start PostgreSQL with pgvector

Create a `docker-compose.yml`:

```yaml
services:
  postgres:
    image: pgvector/pgvector:pg17
    environment:
      POSTGRES_DB: rag
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

## 2. Create the Table

```bash
docker exec -it $(docker compose ps -q postgres) psql -U postgres -d rag
```

The embedding dimension must match the model you plan to use. This playbook uses [`all-MiniLM-L6-v2`](https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2), which outputs 384-dimensional vectors.

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE chunks (
    id          BIGSERIAL PRIMARY KEY,
    source      TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    chunk_text  TEXT NOT NULL,
    embedding   vector(384) NOT NULL
);

CREATE INDEX chunks_embedding_idx
    ON chunks
    USING hnsw (embedding vector_cosine_ops);
```

The `vector_cosine_ops` index pairs with the `<=>` cosine distance operator used in the query flow below.

## 3. Prepare a Sample Document

For this walkthrough, ingest a single file so you can see the round trip end to end. In real use, the source isn't limited to local files - any input that ultimately yields text works. Swap the File input for HTTP, Kafka, an SQL Select, an S3 bucket, a webhook, or anything else that can fetch or receive text, and the rest of this flow stays exactly the same.

Save the snippet below as `/tmp/qaynaq-rag/handbook.md` (create the directory first):

```bash
mkdir -p /tmp/qaynaq-rag
cat > /tmp/qaynaq-rag/handbook.md <<'EOF'
# Acme Coffee Co. Handbook

## Return Policy
Customers may return unopened bags of coffee beans within 30 days of purchase for
a full refund. Opened bags can be returned within 14 days if at least half the
beans remain. Ground coffee is not eligible for return once the seal is broken.

## Shipping
Orders placed before 2pm Berlin time ship the same business day. Standard
shipping within Germany takes 1-2 business days and 3-5 business days across
the rest of the EU. We do not currently ship outside the EU.

## Brewing Guide
For pour-over, use 60 grams of coffee per liter of water at 96 degrees Celsius.
Bloom the grounds with twice their weight in water for 30 seconds before the
main pour. Total brew time should be between three and four minutes.
EOF
```

This gives the chunker three clearly distinct sections (return policy, shipping, brewing) so each query later returns a different chunk and you can see the retrieval working.

## 4. Build the Ingest Flow

Open the Qaynaq UI, click **Create New Flow**, and configure each section:

### Input - select **File**

Point this at the sample file you just created. To ingest a whole directory, use a glob like `/tmp/qaynaq-rag/**/*.md`.

| Field | Value |
|-------|-------|
| Paths | `/tmp/qaynaq-rag/handbook.md` |

Scroll down in the same File input panel to the **Scanner** field and pick **RAG Chunker**, then fill in:

| Field | Value |
|-------|-------|
| Strategy | `Markdown` (or `Recursive` for plain text) |
| Chunk Size | `1000` |
| Overlap | `200` |

The scanner runs inside the input - it controls how the bytes coming out of the file are split into messages before they reach the processors.

### Processors

Three processors in order. The first stashes the chunk text and source metadata so they survive the embedding step (which replaces the message body with the vector).

**Processor 1 - select Mapping**

| Field | Value |
|-------|-------|
| Mapping | `meta chunk_text = content().string()` |

**Processor 2 - select NLP Extract Features**

| Field | Value |
|-------|-------|
| Name | `embedder` |
| Path | `./models` |
| Enable Download | `true` |
| Repository | `sentence-transformers/all-MiniLM-L6-v2` |
| Normalization | `true` |

Normalization is required because we index with `vector_cosine_ops`.

### Output - select **SQL Insert**

| Field | Value |
|-------|-------|
| Driver | `postgres` |
| DSN | `postgres://postgres:postgres@localhost:5432/rag?sslmode=disable` |
| Table | `chunks` |
| Columns | `source`, `chunk_index`, `chunk_text`, `embedding` |
| Args Mapping | `root = [meta("rag_source"), meta("rag_chunk_index").number(), meta("chunk_text"), this.string()]` |

After the embedding processor, `this` refers to the float vector that became the message body. `meta("rag_source")` and `meta("rag_chunk_index")` come from the RAG Chunker; `meta("chunk_text")` is the value the first Mapping processor stashed. The vector must be stringified - pgvector accepts the textual `[0.1,0.2,...]` form on the wire, and the Postgres driver does not know how to bind a raw `[]float32`.

Click **Save** and then **Start** the flow. As the worker processes each file, rows appear in the `chunks` table. Verify:

```bash
docker exec -it $(docker compose ps -q postgres) \
  psql -U postgres -d rag -c "SELECT count(*) FROM chunks;"
```

## 5. Build the Query Flow as an MCP Tool

Create a second flow. This one is exposed as an MCP tool, so AI assistants can call it directly.

### Input - select **MCP Tool**

| Field | Value |
|-------|-------|
| Name | `search_knowledge_base` |
| Description | `Search the knowledge base for passages relevant to a question.` |
| Input Parameters | `query` (string, required) - "The user's question or search phrase" |
| Read-Only | `true` |

### Processors

**Processor 1 - select Mapping**

The MCP Tool input delivers `{ "query": "..." }`. The embedding processor expects the body to be the text it should embed, so reduce the message to just the query string.

| Field | Value |
|-------|-------|
| Mapping | `root = this.query` |

**Processor 2 - select NLP Extract Features**

Use the same configuration as the ingest flow - same model, same Path, same Normalization setting. The downloaded model is reused.

| Field | Value |
|-------|-------|
| Name | `query-embedder` |
| Path | `./models` |
| Enable Download | `true` |
| Repository | `sentence-transformers/all-MiniLM-L6-v2` |
| Normalization | `true` |

**Processor 3 - select SQL Raw**

| Field | Value |
|-------|-------|
| Driver | `postgres` |
| DSN | `postgres://postgres:postgres@localhost:5432/rag?sslmode=disable` |
| Query | `SELECT source, chunk_index, chunk_text, 1 - (embedding <=> $1) AS score FROM chunks ORDER BY embedding <=> $1 LIMIT $2` |
| Args Mapping | `root = [this.string(), 5]` |

`this` is the query embedding (the body, after the previous processor); `.string()` renders it as `[0.1,0.2,...]` which is the format pgvector expects on the wire. `<=>` is pgvector's cosine distance operator; subtracting from 1 turns it into a similarity score that's easier for the AI client to interpret. `LIMIT $2` returns the top 5 matches.

### Output

The output is automatically set to **Sync Response** and locked when using MCP Tool input. The query result (a JSON array of matching rows) is returned to the AI client verbatim.

Click **Save** and then **Start** the flow.

## 6. Test the Tool

Once both flows are running, the `search_knowledge_base` tool appears on the `/mcp` endpoint and MCP-compatible clients pick it up within seconds.

Open your AI assistant (Claude Desktop, Cursor, or any MCP client pointed at your Qaynaq `/mcp` endpoint - see the [MCP Server guide](/docs/guides/mcp-server) for setup) and paste this exact prompt:

> What is Acme Coffee's return policy for opened bags? Quote the relevant passage.

The assistant should pick up `search_knowledge_base` on its own, call it with a query like `"return policy opened bags"`, and answer with the line: *"Opened bags can be returned within 14 days if at least half the beans remain."*

Try a second prompt to see a different chunk win:

> What temperature does Acme Coffee recommend for pour-over brewing?

Expected answer: 96 degrees Celsius, drawn from the Brewing Guide chunk.

If the assistant answers from general knowledge instead of calling the tool, prepend "Use the search_knowledge_base tool to answer:" - some clients need an explicit nudge the first time.

For command-line testing without an AI client, see [MCP Server - Verifying](/docs/guides/mcp-server#verifying).

## Cleanup

```bash
docker compose down -v
rm -rf /tmp/qaynaq-rag
```

## Notes

- **Re-indexing.** Re-running the ingest flow appends new rows. To rebuild from scratch, `TRUNCATE chunks;` before starting it again, or add a `source` filter so each run replaces only the documents it touched.
- **Model choice.** `all-MiniLM-L6-v2` is small, fast, and good enough for most knowledge bases. For higher recall, swap to `BAAI/bge-small-en-v1.5` (also 384 dims, no DDL change) or `sentence-transformers/all-mpnet-base-v2` (768 dims - update `vector(384)` to `vector(768)` and re-create the table).
- **Cosine vs L2.** This playbook uses cosine similarity because the embedding model is L2-normalized. If you turn Normalization off, switch the index to `vector_l2_ops` and the query operator to `<->`.
