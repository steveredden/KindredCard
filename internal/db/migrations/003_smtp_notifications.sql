ALTER TABLE notification_settings 
    ADD COLUMN IF NOT EXISTS provider_type VARCHAR(50) DEFAULT 'discord',
    ADD COLUMN IF NOT EXISTS target_address VARCHAR(255) NULL;

COMMENT ON COLUMN notification_settings.provider_type IS 'Discord Webhooks or SMTP Server';
COMMENT ON COLUMN notification_settings.target_address IS 'SMTP TO: email address';