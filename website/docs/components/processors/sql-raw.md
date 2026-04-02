# SQL Raw

Runs an arbitrary SQL query against a database and replaces the message with the rows returned. If the query fails, the message is marked as having failed and will not be sent to downstream processors or outputs.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Driver | string | — | Database driver (required) |
| DSN | string | — | Connection string for the target database (required) |
| Query | string | — | The SQL query to execute. Placeholder style depends on the driver (required) |
| Unsafe Dynamic Query | boolean | `false` | Enable interpolation functions in the query. WARNING: may be susceptible to SQL injection attacks |
| Args Mapping | Bloblang | — | Maps message fields to query placeholder arguments |
| Exec Only | boolean | `false` | Discard the query result and leave the message unchanged. Useful for INSERT, UPDATE, or DELETE statements |
| Init Statement | string | — | SQL statement to execute on the first connection (e.g., to create tables) |
| Conn Max Idle Time | string | — | Maximum idle duration before a connection is closed |
| Conn Max Life Time | string | — | Maximum total lifetime of a connection |
| Conn Max Idle | integer | `2` | Maximum number of idle connections in the pool |
| Conn Max Open | integer | `0` | Maximum number of open connections (`0` = unlimited) |

Supported drivers: `mysql`, `postgres`, `clickhouse`, `mssql`, `sqlite`, `oracle`, `snowflake`, `trino`, `gocosmos`, `spanner`.

## Placeholder Styles

Each driver uses a specific placeholder style in queries:

- **mysql, mssql, sqlite, snowflake, spanner, trino**: `?` (question mark)
- **postgres, clickhouse**: `$1`, `$2`, ... (dollar sign)
- **oracle, gocosmos**: `:1`, `:2`, ... (colon)

## Exec Only Mode

When Exec Only is enabled, the query result is discarded and the message content is left unchanged. This is useful when executing INSERT, UPDATE, or DELETE statements where you don't need the result.

## Dynamic Queries

By default, the query is treated as a static string with placeholder arguments. Enable Unsafe Dynamic Query to use interpolation functions within the query itself. Use this with caution as it can expose the query to SQL injection attacks.

:::tip
Use Args Mapping to safely pass message field values as query parameters instead of embedding them directly in the query string.
:::
