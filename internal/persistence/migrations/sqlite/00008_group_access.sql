CREATE TABLE IF NOT EXISTS groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'observed',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at DATETIME
);

CREATE TABLE IF NOT EXISTS user_groups (
    email TEXT PRIMARY KEY,
    groups TEXT NOT NULL DEFAULT '[]',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mcp_call_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor TEXT NOT NULL,
    actor_kind TEXT NOT NULL,
    groups TEXT NOT NULL DEFAULT '[]',
    tool_name TEXT NOT NULL,
    target_kind TEXT NOT NULL,
    target_id INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    error TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mcp_call_logs_created_at ON mcp_call_logs (created_at);

ALTER TABLE flows ADD COLUMN allowed_groups TEXT NOT NULL DEFAULT '[]';
ALTER TABLE mcp_servers ADD COLUMN allowed_groups TEXT NOT NULL DEFAULT '[]';
