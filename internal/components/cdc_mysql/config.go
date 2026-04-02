package cdc_mysql

import "github.com/warpstreamlabs/bento/public/service"

const (
	cmfHost                  = "host"
	cmfPort                  = "port"
	cmfUser                  = "user"
	cmfPassword              = "password"
	cmfServerID              = "server_id"
	cmfPositionCache         = "position_cache"
	cmfPositionCacheKey      = "position_cache_key"
	cmfIncludeTables         = "include_tables"
	cmfExcludeTables         = "exclude_tables"
	cmfUseSchemaCache        = "use_schema_cache"
	cmfSchemaCacheTTL        = "schema_cache_ttl"
	cmfPositionMode          = "position_mode"
	cmfFlavor                = "flavor"
	cmfCacheSaveInterval     = "cache_save_interval"
	cmfMaxBatchSize          = "max_batch_size"
	cmfMaxPendingCheckpoints = "max_pending_checkpoints"
	cmfRetryInitialInterval  = "retry_initial_interval"
	cmfRetryMaxInterval      = "retry_max_interval"
	cmfRetryMultiplier       = "retry_multiplier"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Services").
		Summary("Consumes change events from a MySQL database binlog.").
		Description(`
This input connects to a MySQL database and monitors the binary log (binlog) for change events.
It outputs structured messages containing information about INSERT, UPDATE, and DELETE operations.

Each message contains:
- database: The name of the database
- table: The name of the table
- type: The operation type (insert, update, delete)
- ts: Unix timestamp when the event was processed
- server_id: Identifier for this input instance
- data: The row data (for all operations)
- old: The previous row data (for update operations only)

The input maintains position information to resume from the last processed event on restart.

Use include_tables and exclude_tables to filter which tables to monitor.
Table names should be in the format: schema_name.table_name`).
		Field(service.NewStringField(cmfHost).
			Description("MySQL server hostname or IP address.").
			Default("localhost")).
		Field(service.NewIntField(cmfPort).
			Description("MySQL server port.").
			Default(3306)).
		Field(service.NewStringField(cmfUser).
			Description("MySQL username for connection.").
			Default("root")).
		Field(service.NewStringField(cmfPassword).
			Description("MySQL password for connection.").
			Secret()).
		Field(service.NewStringField(cmfServerID).
			Description("Unique server ID for this binlog consumer. Corresponds to MySQL's server_id. Can be an integer or string that will be hashed to integer.").
			Default("1000")).
		Field(service.NewStringField(cmfPositionCache).
			Description("Name of the cache resource to use for position tracking.").
			Default("")).
		Field(service.NewStringField(cmfPositionCacheKey).
			Description("Cache key to use for storing position information.").
			Default("")).
		Field(service.NewStringField(cmfPositionMode).
			Description("Position tracking mode: 'gtid' (default) or 'file'.").
			Default("gtid")).
		Field(service.NewStringField(cmfFlavor).
			Description("Database flavor: 'mysql' (default) or 'mariadb'.").
			Default("mysql")).
		Field(service.NewStringField(cmfCacheSaveInterval).
			Description("Interval for saving binlog position to cache (e.g., '10s', '1m'). Position is saved at this interval, on graceful shutdown, or when context is cancelled. Set to '0s' to save immediately on every position change (not recommended for cloud providers).").
			Default("30s")).
		Field(service.NewStringListField(cmfIncludeTables).
			Description("List of tables to monitor in format 'schema_name.table_name'. If empty, all tables are monitored.").
			Optional()).
		Field(service.NewStringListField(cmfExcludeTables).
			Description("List of tables to exclude in format 'schema_name.table_name'.").
			Optional()).
		Field(service.NewBoolField(cmfUseSchemaCache).
			Description("Enable schema caching to get column names without requiring --binlog-row-metadata=FULL. This queries INFORMATION_SCHEMA to map column positions to names.").
			Default(true)).
		Field(service.NewStringField(cmfSchemaCacheTTL).
			Description("TTL for schema cache entries (e.g., '5m', '1h'). Schema is refreshed when cache expires.").
			Default("5m")).
		Field(service.NewIntField(cmfMaxBatchSize).
			Description("Maximum number of messages per batch. Prevents unbounded memory growth for large transactions.").
			Default(1000)).
		Field(service.NewIntField(cmfMaxPendingCheckpoints).
			Description("Maximum number of pending checkpoints in tracker. Provides backpressure when downstream is slow (applies to both GTID and file position modes).").
			Default(100)).
		Field(service.NewStringField(cmfRetryInitialInterval).
			Description("Initial wait time before first retry attempt (e.g., '1s', '500ms'). Doubles with each retry up to max interval.").
			Default("1s")).
		Field(service.NewStringField(cmfRetryMaxInterval).
			Description("Maximum wait time between retry attempts (e.g., '30s', '1m'). Caps the exponential backoff.").
			Default("30s")).
		Field(service.NewFloatField(cmfRetryMultiplier).
			Description("Multiplier for exponential backoff. Each retry waits multiplier times longer than the previous (e.g., 2.0 for doubling).").
			Default(2.0)).
		Version("1.0.0").
		Example("Basic MySQL CDC monitoring",
			`Monitor all changes using GTID position tracking with cache`,
			`
input:
  cdc_mysql:
    host: localhost
    port: 3306
    user: replica_user
    password: replica_password
    server_id: "1001"
    position_cache: positions
    position_cache_key: mysql_db1
    cache_save_interval: 30s
    max_batch_size: 1000
    max_pending_checkpoints: 100
`,
		).
		Example("MySQL CDC with file position",
			`Monitor changes using traditional binlog file and position`,
			`
input:
  cdc_mysql:
    host: localhost
    port: 3306
    user: replica_user
    password: replica_password
    server_id: "1001"
    position_cache: positions
    position_cache_key: mysql_db1
    position_mode: file
    cache_save_interval: 30s
`,
		).
		Example("Monitor specific tables",
			`Monitor only specific tables using include_tables`,
			`
input:
  cdc_mysql:
    host: mysql.example.com
    port: 3306
    user: replica_user
    password: replica_password
    server_id: "1001"
    include_tables:
      - "mydb.users"
      - "mydb.products"
      - "mydb.orders"
    exclude_tables:
      - "mydb.sessions"
      - "mydb.logs"
    position_cache: positions
    position_cache_key: mydb
    cache_save_interval: 1m
`,
		).
		Example("MariaDB CDC monitoring",
			`Monitor changes from a MariaDB database`,
			`
input:
  cdc_mysql:
    host: mariadb.example.com
    port: 3306
    user: replica_user
    password: replica_password
    server_id: "1001"
    flavor: mariadb
    position_cache: positions
    position_cache_key: mariadb_db
    cache_save_interval: 30s
`,
		)
}
