ALTER TABLE connections ADD COLUMN expires_at datetime;
CREATE INDEX IF NOT EXISTS idx_connections_expires_at ON connections(expires_at);
