# CDC MySQL

Captures changes from MySQL/MariaDB binary log (CDC — Change Data Capture). Supports batched reads, automatic reconnection with exponential backoff, and ordered checkpointing.

**Prerequisites:** MySQL must have `log_bin = ON`, `binlog_format = ROW`, and `binlog_row_image = FULL`. For GTID mode, also enable `gtid_mode = ON` and `enforce_gtid_consistency = ON`. The user needs `REPLICATION SLAVE`, `REPLICATION CLIENT`, and `SELECT` privileges.

## Connection

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Host | string | `localhost` | MySQL server hostname or IP address |
| Port | integer | `3306` | MySQL server port |
| User | string | `root` | MySQL username |
| Password | string | — | MySQL password |
| Server ID | string | `1000` | Unique server ID for this binlog consumer. Corresponds to MySQL's `server_id` |
| Flavor | string | `mysql` | Database flavor (`mysql` or `mariadb`) |

## Position tracking

Requires a [cache resource](/docs/components/caches) to persist binlog position across restarts.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Position Cache | cache | — | Cache resource for storing binlog position |
| Position Cache Key | string | — | Key within the cache to store position data |
| Position Mode | string | `gtid` | Position tracking mode: `gtid` or `file` |
| Cache Save Interval | string | `30s` | How often to persist position to cache. `0s` for immediate saves |

**GTID mode** (default, recommended): stores a GTID set and on first start queries `gtid_purged` from MySQL. **File mode**: stores binlog filename and position. If the stored position is no longer available (e.g., binlogs purged), the component automatically purges and reconnects from the earliest available position.

## Table filtering

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Include Tables | array | — | Tables to monitor in `schema.table` format. If empty, all tables are monitored |
| Exclude Tables | array | — | Tables to exclude in `schema.table` format |

## Schema cache

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Use Schema Cache | boolean | `true` | Query INFORMATION_SCHEMA for column names |
| Schema Cache TTL | string | `5m` | How long cached schema information is valid |

## Batching and backpressure

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Max Batch Size | integer | `1000` | Maximum messages per batch |
| Max Pending Checkpoints | integer | `100` | Maximum unacknowledged batches before applying backpressure |

## Connection retry

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Retry Initial Interval | string | `1s` | Initial wait before first reconnect attempt |
| Retry Max Interval | string | `30s` | Maximum wait between reconnect attempts |
| Retry Multiplier | float | `2.0` | Exponential backoff multiplier |

## Output format

Each message contains:

| Field | Type | Description |
|-------|------|-------------|
| `database` | string | Database name |
| `table` | string | Table name |
| `type` | string | Operation type: `insert`, `update`, or `delete` |
| `ts` | integer | Unix timestamp when the event was processed |
| `server_id` | string | The configured server ID value |
| `new` | object | Row after the change — present for `insert` and `update`, absent for `delete` |
| `old` | object | Row before the change — present for `update` and `delete`, absent for `insert` |
| `gtid` | string | GTID of the transaction (when using GTID mode) |

:::tip
For CDC position tracking, use a persistent cache type like **Redis** or **File**. The **Memory** cache works for development but loses position data on restart, causing the consumer to reprocess events from the earliest available position.
:::
