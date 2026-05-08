ALTER TABLE mcp_servers ADD COLUMN transport TEXT NOT NULL DEFAULT 'http';
ALTER TABLE mcp_servers ADD COLUMN catalog_id TEXT NOT NULL DEFAULT '';
ALTER TABLE mcp_servers ADD COLUMN command TEXT NOT NULL DEFAULT '';
ALTER TABLE mcp_servers ADD COLUMN args TEXT NOT NULL DEFAULT '';
ALTER TABLE mcp_servers ADD COLUMN encrypted_env TEXT NOT NULL DEFAULT '';
ALTER TABLE mcp_servers ADD COLUMN process_state TEXT NOT NULL DEFAULT 'stopped';

CREATE TABLE IF NOT EXISTS mcp_servers__new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    url TEXT,
    auth_type TEXT NOT NULL DEFAULT 'none',
    auth_header TEXT DEFAULT '',
    encrypted_auth_value TEXT DEFAULT '',
    connection_name TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    tool_count INTEGER DEFAULT 0,
    last_sync_at DATETIME,
    last_error TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    transport TEXT NOT NULL DEFAULT 'http',
    catalog_id TEXT NOT NULL DEFAULT '',
    command TEXT NOT NULL DEFAULT '',
    args TEXT NOT NULL DEFAULT '',
    encrypted_env TEXT NOT NULL DEFAULT '',
    process_state TEXT NOT NULL DEFAULT 'stopped'
);

INSERT INTO mcp_servers__new (
    id, name, url, auth_type, auth_header, encrypted_auth_value, connection_name,
    status, tool_count, last_sync_at, last_error, created_at, updated_at,
    transport, catalog_id, command, args, encrypted_env, process_state
)
SELECT
    id, name, url, auth_type, auth_header, encrypted_auth_value, connection_name,
    status, tool_count, last_sync_at, last_error, created_at, updated_at,
    transport, catalog_id, command, args, encrypted_env, process_state
FROM mcp_servers;

DROP TABLE mcp_servers;
ALTER TABLE mcp_servers__new RENAME TO mcp_servers;
