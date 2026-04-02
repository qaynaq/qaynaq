# SQL Select

Runs a SELECT query against a database and replaces the message with the rows returned. Each row is represented as a JSON object with column names as keys.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Driver | string | — | Database driver (required) |
| DSN | string | — | Connection string for the target database (required) |
| Table | string | — | The table to query (required) |
| Columns | array | — | A list of column names to select (required) |
| Where | string | — | An optional WHERE clause. Placeholder arguments are populated with Args Mapping |
| Args Mapping | Bloblang | — | Maps message fields to WHERE clause placeholder arguments |
| Prefix | string | — | An optional prefix to prepend before SELECT |
| Suffix | string | — | An optional suffix to append to the query |
| Init Statement | string | — | SQL statement to execute on the first connection |
| Conn Max Idle Time | string | — | Maximum idle duration before a connection is closed |
| Conn Max Life Time | string | — | Maximum total lifetime of a connection |
| Conn Max Idle | integer | `2` | Maximum number of idle connections in the pool |
| Conn Max Open | integer | `0` | Maximum number of open connections (`0` = unlimited) |

Supported drivers: `mysql`, `postgres`, `clickhouse`, `mssql`, `sqlite`, `oracle`, `snowflake`, `trino`, `gocosmos`, `spanner`.

## Result Format

The message payload is replaced with an array of objects, where each object represents a row and column names are used as keys. If the query returns no rows, the message is replaced with an empty array.

## Filtering with Where

Use the Where field to filter rows. Placeholder arguments in the WHERE clause are populated from the Args Mapping field. The placeholder style depends on the driver (e.g., `?` for MySQL, `$1` for PostgreSQL).

:::tip
Use Prefix and Suffix to add clauses like `DISTINCT` (prefix) or `ORDER BY`, `LIMIT` (suffix) to the generated SELECT query.
:::
