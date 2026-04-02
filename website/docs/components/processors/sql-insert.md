# SQL Insert

Inserts a row into an SQL database for each message, and leaves the message content unchanged.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Driver | string | — | Database driver (required) |
| DSN | string | — | Connection string for the target database (required) |
| Table | string | — | The target table (required) |
| Columns | array | — | Column names to populate (required) |
| Args Mapping | Bloblang | — | Maps message fields to column values. Must evaluate to an array matching the number of columns (required) |
| Prefix | string | — | An optional prefix to prepend before INSERT |
| Suffix | string | — | An optional suffix to append (e.g., `ON CONFLICT (name) DO NOTHING`) |
| Init Statement | string | — | SQL statement to execute on the first connection |
| Conn Max Idle Time | string | — | Maximum idle duration before a connection is closed |
| Conn Max Life Time | string | — | Maximum total lifetime of a connection |
| Conn Max Idle | integer | `2` | Maximum number of idle connections in the pool |
| Conn Max Open | integer | `0` | Maximum number of open connections (`0` = unlimited) |

Supported drivers: `mysql`, `postgres`, `clickhouse`, `mssql`, `sqlite`, `oracle`, `snowflake`, `trino`, `gocosmos`, `spanner`.

## Conflict Handling

Use the Suffix field to handle conflicts. For example, in PostgreSQL you can use `ON CONFLICT (id) DO NOTHING` to skip duplicate rows, or `ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name` to upsert.

:::tip
When using the SQL Insert processor in a pipeline (rather than as an output), the message content remains unchanged after the insert. This allows you to chain additional processors or route the message to other outputs after the database insert.
:::
