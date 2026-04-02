CREATE TABLE IF NOT EXISTS events (
    id integer PRIMARY KEY,
    flow_id integer,
    worker_flow_id integer NOT NULL,
    trace_id text,
    section text NOT NULL,
    component_label text,
    type text NOT NULL,
    content text NOT NULL,
    meta blob NOT NULL,
    created_at datetime
);
CREATE INDEX IF NOT EXISTS idx_events_flow_id ON events(flow_id);
CREATE INDEX IF NOT EXISTS idx_events_worker_flow_id ON events(worker_flow_id);
CREATE INDEX IF NOT EXISTS idx_events_trace_id ON events(trace_id);

CREATE TABLE IF NOT EXISTS flows (
    id integer PRIMARY KEY,
    parent_id integer,
    name text NOT NULL,
    input_label text,
    input_component text NOT NULL,
    input_config blob NOT NULL,
    output_label text,
    output_component text NOT NULL,
    output_config blob NOT NULL,
    buffer_id integer,
    is_current numeric DEFAULT true,
    is_ready numeric DEFAULT false,
    builder_state blob,
    status text NOT NULL,
    managed_by text,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_flows_parent_id ON flows(parent_id);

CREATE TABLE IF NOT EXISTS flow_processors (
    id integer PRIMARY KEY,
    flow_id integer NOT NULL,
    component text NOT NULL,
    label text,
    config blob,
    created_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_flow_processors_flow_id ON flow_processors(flow_id);

CREATE TABLE IF NOT EXISTS flow_caches (
    id integer PRIMARY KEY,
    flow_id integer NOT NULL,
    cache_id integer NOT NULL,
    created_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_flow_caches_flow_id ON flow_caches(flow_id);
CREATE INDEX IF NOT EXISTS idx_flow_caches_cache_id ON flow_caches(cache_id);

CREATE TABLE IF NOT EXISTS workers (
    id text PRIMARY KEY,
    address text,
    last_heartbeat datetime,
    status text
);

CREATE TABLE IF NOT EXISTS worker_flows (
    id integer PRIMARY KEY,
    worker_id text NOT NULL,
    flow_id integer NOT NULL,
    input_events integer NOT NULL DEFAULT 0,
    processor_errors integer NOT NULL DEFAULT 0,
    output_events integer NOT NULL DEFAULT 0,
    status text NOT NULL,
    lease_expires_at datetime,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_worker_flows_worker_id ON worker_flows(worker_id);
CREATE INDEX IF NOT EXISTS idx_worker_flows_flow_id ON worker_flows(flow_id);

CREATE TABLE IF NOT EXISTS secrets (
    key text PRIMARY KEY,
    encrypted_value text NOT NULL,
    created_at datetime
);

CREATE TABLE IF NOT EXISTS caches (
    id integer PRIMARY KEY,
    parent_id integer,
    label text NOT NULL,
    component text NOT NULL,
    config blob NOT NULL,
    is_current numeric DEFAULT true,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_caches_parent_id ON caches(parent_id);

CREATE TABLE IF NOT EXISTS buffers (
    id integer PRIMARY KEY,
    parent_id integer,
    label text NOT NULL,
    component text NOT NULL,
    config blob NOT NULL,
    is_current numeric DEFAULT true,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_buffers_parent_id ON buffers(parent_id);

CREATE TABLE IF NOT EXISTS rate_limits (
    id integer PRIMARY KEY,
    parent_id integer,
    label text NOT NULL,
    component text NOT NULL,
    config blob NOT NULL,
    is_current numeric DEFAULT true,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_rate_limits_parent_id ON rate_limits(parent_id);

CREATE TABLE IF NOT EXISTS rate_limit_states (
    id integer PRIMARY KEY,
    rate_limit_label text NOT NULL,
    key text NOT NULL,
    tokens real NOT NULL,
    last_refill_at datetime NOT NULL,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_label_key ON rate_limit_states(rate_limit_label, key);

CREATE TABLE IF NOT EXISTS flow_rate_limits (
    id integer PRIMARY KEY,
    flow_id integer NOT NULL,
    rate_limit_id integer NOT NULL,
    created_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_flow_rate_limits_flow_id ON flow_rate_limits(flow_id);
CREATE INDEX IF NOT EXISTS idx_flow_rate_limits_rate_limit_id ON flow_rate_limits(rate_limit_id);

CREATE TABLE IF NOT EXISTS flow_buffers (
    id integer PRIMARY KEY,
    flow_id integer NOT NULL,
    buffer_id integer NOT NULL,
    created_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_flow_buffers_flow_id ON flow_buffers(flow_id);
CREATE INDEX IF NOT EXISTS idx_flow_buffers_buffer_id ON flow_buffers(buffer_id);

CREATE TABLE IF NOT EXISTS files (
    id integer PRIMARY KEY,
    parent_id integer,
    key text NOT NULL,
    content blob NOT NULL,
    size integer NOT NULL DEFAULT 0,
    is_current boolean NOT NULL DEFAULT true,
    created_at datetime NOT NULL,
    updated_at datetime
);
CREATE INDEX IF NOT EXISTS idx_files_key ON files(key);
CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);

CREATE TABLE IF NOT EXISTS settings (
    key text PRIMARY KEY,
    value text NOT NULL,
    updated_at datetime
);

CREATE TABLE IF NOT EXISTS api_tokens (
    id integer PRIMARY KEY,
    name text NOT NULL UNIQUE,
    token_hash text NOT NULL,
    scopes text NOT NULL DEFAULT '[]',
    last_used_at datetime,
    created_at datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS connections (
    name text PRIMARY KEY,
    provider text NOT NULL,
    encrypted_config text NOT NULL,
    encrypted_token text NOT NULL,
    created_at datetime,
    updated_at datetime
);
