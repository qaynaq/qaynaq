CREATE TABLE IF NOT EXISTS mcp_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    auth_type TEXT NOT NULL DEFAULT 'none',
    auth_header TEXT DEFAULT '',
    encrypted_auth_value TEXT DEFAULT '',
    connection_name TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    tool_count INTEGER DEFAULT 0,
    last_sync_at DATETIME,
    last_error TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
