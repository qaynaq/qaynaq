# SQL Insert

Inserts rows into a SQL database.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Driver | string | — | Database driver |
| DSN | string | — | Connection string |
| Table | string | — | Target table |
| Columns | array | — | Column names |
| Args Mapping | Bloblang | — | Maps message fields to column values |
| Suffix | string | — | Optional SQL suffix (e.g., `ON CONFLICT DO NOTHING`) |
| Max In Flight | integer | `64` | Maximum parallel inserts |
| Batching | object | — | Batching policy |

Supported drivers: `mysql`, `postgres`, `clickhouse`, `mssql`, `sqlite`, `oracle`, `snowflake`, `trino`, `cosmos`, `spanner`.
