ALTER TABLE contacts ADD COLUMN version_token INT DEFAULT 0;
UPDATE contacts SET version_token = last_modified_token;