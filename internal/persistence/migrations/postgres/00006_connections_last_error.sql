ALTER TABLE connections ADD COLUMN last_error TEXT NOT NULL DEFAULT '';
ALTER TABLE connections ADD COLUMN last_error_at timestamptz;
ALTER TABLE connections ADD COLUMN first_failed_at timestamptz;
ALTER TABLE connections ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_connections_last_error_at ON connections(last_error_at);
