ALTER TABLE flows ADD COLUMN last_error TEXT NOT NULL DEFAULT '';
ALTER TABLE flows ADD COLUMN last_error_at datetime;
