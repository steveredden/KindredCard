ALTER TABLE addresses ADD COLUMN IF NOT EXISTS extended_street TEXT;
UPDATE phones SET type = array_replace(type, 'mobile', 'cell') WHERE 'mobile' = ANY(type);