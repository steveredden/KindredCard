ALTER TABLE phones ADD COLUMN IF NOT EXISTS last_formatted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_phones_last_formatted ON phones(last_formatted_at);