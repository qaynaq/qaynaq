CREATE TABLE IF NOT EXISTS oauth_clients (
    id text PRIMARY KEY,
    secret_hash text NOT NULL,
    name text NOT NULL,
    redirect_uris text NOT NULL DEFAULT '[]',
    created_at datetime NOT NULL,
    last_used_at datetime
);

CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
    id integer PRIMARY KEY,
    token_hash text NOT NULL UNIQUE,
    client_id text NOT NULL,
    user_email text NOT NULL,
    expires_at datetime NOT NULL,
    revoked_at datetime,
    created_at datetime NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_client_id ON oauth_refresh_tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_tokens_user_email ON oauth_refresh_tokens(user_email);

CREATE TABLE IF NOT EXISTS oauth_consents (
    id integer PRIMARY KEY,
    user_email text NOT NULL,
    client_id text NOT NULL,
    scope text NOT NULL DEFAULT '',
    approved_at datetime NOT NULL,
    UNIQUE(user_email, client_id, scope)
);
CREATE INDEX IF NOT EXISTS idx_oauth_consents_lookup ON oauth_consents(user_email, client_id);
