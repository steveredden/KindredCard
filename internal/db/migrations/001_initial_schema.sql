-- KindredCard Personal CRM - Complete Database Schema
-- Drop existing database and recreate for clean slate
-- Usage: psql -d postgres -c "DROP DATABASE IF EXISTS kindredcard; CREATE DATABASE kindredcard;"
--        psql -d kindredcard -f schema.sql

-- Enable UUID extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- CORE TABLES
-- =============================================================================

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_setup_complete BOOLEAN DEFAULT FALSE,
    theme VARCHAR(20) DEFAULT 'winter', -- light, dark, system
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    addressbook_sync_token BIGINT NOT NULL DEFAULT 1,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for authentication
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,

    -- Device/Browser Information
    user_agent TEXT,
    browser VARCHAR(50),
    browser_version VARCHAR(20),
    os VARCHAR(50),
    device VARCHAR(50),
    is_mobile BOOLEAN DEFAULT FALSE,
    referer TEXT,
    language VARCHAR(50),
    
    -- Network Information
    ip_address VARCHAR(45), -- IPv6 support

    -- Session Metadata
    login_time TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),

    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_token UNIQUE(token)
);

-- API Tokens for authentication
CREATE TABLE IF NOT EXISTS api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,   -- SHA-256 hash
    token_signature VARCHAR(64),              -- HMAC-SHA256 with APP_KEY
    name VARCHAR(255) NOT NULL,               -- "Mobile App", "CI/CD"
    expires_at TIMESTAMP,                     -- Optional expiration
    is_active BOOLEAN NOT NULL DEFAULT true,  -- Revocation flag
    last_used_at TIMESTAMP,                   -- Track usage
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Notification settings for Discord webhooks and/or SMTP
CREATE TABLE IF NOT EXISTS notification_settings (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    webhook_url TEXT,
    days_look_ahead INTEGER DEFAULT 1,
    notification_time VARCHAR(5) DEFAULT '09:00', -- HH:MM format
    include_birthdays BOOLEAN DEFAULT TRUE,
    include_anniversaries BOOLEAN DEFAULT TRUE,
    include_event_dates BOOLEAN DEFAULT FALSE,
    other_event_regex TEXT,
    enabled BOOLEAN DEFAULT FALSE,
    last_sent_at TIMESTAMP NULL,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Relationship types (predefined and custom)
CREATE TABLE IF NOT EXISTS relationship_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    reverse_name_male VARCHAR(100),
    reverse_name_female VARCHAR(100),
    reverse_name_neutral VARCHAR(100),
    is_system BOOLEAN DEFAULT FALSE
);

-- =============================================================================
-- CONTACTS AND RELATED DATA
-- =============================================================================

-- Main contacts table
CREATE TABLE IF NOT EXISTS contacts (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Name fields (full_name is computed from other fields)
    full_name VARCHAR(255),
    given_name VARCHAR(100),
    family_name VARCHAR(100),
    middle_name VARCHAR(100),
    prefix VARCHAR(50),
    suffix VARCHAR(50),
    nickname VARCHAR(100),
    
    -- Demographics
    gender VARCHAR(1), -- M, F, O, N, U (Male, Female, Other, None/Not applicable, Unknown)
    
    -- Birthday - supports both full dates and partial (month/day only)
    birthday DATE,
    birthday_month INTEGER, -- 1-12, for when year is unknown
    birthday_day INTEGER,   -- 1-31, for when year is unknown
    
    -- Anniversary - supports both full dates and partial (month/day only)
    anniversary DATE,
    anniversary_month INTEGER, -- 1-12, for when year is unknown
    anniversary_day INTEGER,   -- 1-31, for when year is unknown
    
    -- Metadata
    notes TEXT,
    
    -- Avatar stored as base64
    avatar_base64 TEXT,
    avatar_mime_type VARCHAR(50),
    
    -- CardDAV sync control
    exclude_from_sync BOOLEAN DEFAULT FALSE,
    card_origin VARCHAR(255),

    -- The token value when this contact was last modified
    last_modified_token BIGINT NOT NULL DEFAULT 1,

    -- ETag (must change whenever vcard changes)
    etag VARCHAR(255) NOT NULL,
    
    -- Timestamps and versioning
    deleted_at TIMESTAMP NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (user_id, uid)
);

-- Other Dates table
CREATE TABLE IF NOT EXISTS other_dates (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    event_name VARCHAR(255),
        
    -- Dates - supports both full dates and partial (month/day only)
    event_date DATE,
    event_date_month INTEGER, -- 1-12, for when year is unknown
    event_date_day INTEGER   -- 1-31, for when year is unknown
);

-- Email addresses
CREATE TABLE IF NOT EXISTS emails (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    type TEXT[], -- home, work, other
    is_primary BOOLEAN DEFAULT FALSE
);

-- Phone numbers
CREATE TABLE IF NOT EXISTS phones (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    phone VARCHAR(50) NOT NULL,
    type TEXT[], -- home, work, cell, fax, other
    is_primary BOOLEAN DEFAULT FALSE
);

-- Physical addresses
CREATE TABLE IF NOT EXISTS addresses (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    street TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    type TEXT[], -- home, work, other
    is_primary BOOLEAN DEFAULT FALSE
);

-- Organizations/Companies
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    name VARCHAR(255),
    title VARCHAR(255),
    department VARCHAR(255),
    is_primary BOOLEAN DEFAULT FALSE
);

-- URLs/Websites
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    type TEXT[]  -- website, social, other
);

-- Relationships between contacts
CREATE TABLE IF NOT EXISTS relationships (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    related_contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    relationship_type_id INTEGER REFERENCES relationship_types(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(contact_id, related_contact_id, relationship_type_id)
);

-- Other Relationships that came from a CardDAV server and didn't match KindredCard relationship_types and/or contact full_names
CREATE TABLE IF NOT EXISTS other_relationships (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    related_contact_name VARCHAR(255),
    relationship_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Contact indexes
CREATE INDEX IF NOT EXISTS idx_contacts_uid ON contacts(uid);
CREATE INDEX IF NOT EXISTS idx_contacts_full_name ON contacts(full_name);
CREATE INDEX IF NOT EXISTS idx_contacts_given_name ON contacts(given_name);
CREATE INDEX IF NOT EXISTS idx_contacts_family_name ON contacts(family_name);
CREATE INDEX IF NOT EXISTS idx_contacts_birthday ON contacts(birthday);
CREATE INDEX IF NOT EXISTS idx_contacts_anniversary ON contacts(anniversary);
CREATE INDEX IF NOT EXISTS idx_contacts_exclude_from_sync ON contacts(exclude_from_sync);
CREATE INDEX IF NOT EXISTS idx_contacts_sync_check ON contacts (user_id, last_modified_token);
CREATE INDEX IF NOT EXISTS idx_contacts_active ON contacts (user_id, deleted_at) WHERE deleted_at IS NULL;

-- Partial date indexes for event queries
CREATE INDEX IF NOT EXISTS idx_contacts_birthday_month_day ON contacts(birthday_month, birthday_day) 
    WHERE birthday_month IS NOT NULL AND birthday_day IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contacts_anniversary_month_day ON contacts(anniversary_month, anniversary_day)
    WHERE anniversary_month IS NOT NULL AND anniversary_day IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_contacts_event_month_day ON other_dates(event_date_month, event_date_day)
    WHERE event_date_month IS NOT NULL AND event_date_day IS NOT NULL;

-- Related data indexes
CREATE INDEX IF NOT EXISTS idx_emails_contact_id ON emails(contact_id);
CREATE INDEX IF NOT EXISTS idx_emails_email ON emails(email);
CREATE INDEX IF NOT EXISTS idx_phones_contact_id ON phones(contact_id);
CREATE INDEX IF NOT EXISTS idx_addresses_contact_id ON addresses(contact_id);
CREATE INDEX IF NOT EXISTS idx_organizations_contact_id ON organizations(contact_id);
CREATE INDEX IF NOT EXISTS idx_urls_contact_id ON urls(contact_id);
CREATE INDEX IF NOT EXISTS idx_other_dates_contact_id ON other_dates(contact_id);

-- Relationship indexes
CREATE INDEX IF NOT EXISTS idx_relationships_contact_id ON relationships(contact_id);
CREATE INDEX IF NOT EXISTS idx_relationships_related_contact_id ON relationships(related_contact_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(relationship_type_id);
CREATE INDEX IF NOT EXISTS idx_other_relationships_contact_id ON other_relationships(contact_id);

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions(user_id, expires_at);

-- Notification settings indexes
CREATE INDEX IF NOT EXISTS idx_notification_settings_user_id ON notification_settings(user_id);

-- API Tokens
CREATE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_active ON api_tokens(token_hash) WHERE is_active = true;

-- =============================================================================
-- TRIGGERS AND FUNCTIONS
-- =============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
DROP TRIGGER IF EXISTS update_contacts_updated_at ON contacts;
CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_notification_settings_updated_at ON notification_settings;
CREATE TRIGGER update_notification_settings_updated_at BEFORE UPDATE ON notification_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- DEFAULT DATA
-- =============================================================================

-- Insert core relationship types
INSERT INTO relationship_types (name, reverse_name_male, reverse_name_female, reverse_name_neutral, is_system) VALUES
    -- Immediate family
    ('Mother', 'Son', 'Daughter', 'Child', TRUE),
    ('Father', 'Son', 'Daughter', 'Child', TRUE),
    ('Son', 'Father', 'Mother', 'Parent', TRUE),
    ('Daughter', 'Father', 'Mother', 'Parent', TRUE),
    ('Parent', 'Son', 'Daughter', 'Child', TRUE),
    ('Child', 'Father', 'Mother', 'Parent', TRUE),
    
    -- Siblings
    ('Brother', 'Brother', 'Sister', 'Sibling', TRUE),
    ('Sister', 'Brother', 'Sister', 'Sibling', TRUE),
    ('Sibling', 'Brother', 'Sister', 'Sibling', TRUE),
    
    -- Marital
    ('Husband', 'Husband', 'Wife', 'Spouse', TRUE),
    ('Wife', 'Husband', 'Wife', 'Spouse', TRUE),
    ('Spouse', 'Husband', 'Wife', 'Spouse', TRUE),
    ('Partner', 'Partner', 'Partner', 'Partner', TRUE),
    ('Ex-Husband', 'Ex-Husband', 'Ex-Wife', 'Ex-Spouse', TRUE),
    ('Ex-Wife', 'Ex-Husband', 'Ex-Wife', 'Ex-Spouse', TRUE),
    ('Ex-Spouse', 'Ex-Husband', 'Ex-Wife', 'Ex-Spouse', TRUE),
    ('Ex-Partner', 'Ex-Husband', 'Ex-Wife', 'Ex-Partner', TRUE),

    -- Other Relationships
    ('Boyfriend', 'Boyfriend', 'Girlfriend', 'Significant Other', TRUE),
    ('Girlfriend', 'Boyfriend', 'Girlfriend', 'Significant Other', TRUE),
    ('Significant Other', 'Significant Other', 'Significant Other', 'Significant Other', TRUE),
    
    -- Step family
    ('Step-Mother', 'Step-Son', 'Step-Daughter', 'Step-Child', TRUE),
    ('Step-Father', 'Step-Son', 'Step-Daughter', 'Step-Child', TRUE),
    ('Step-Son', 'Step-Father', 'Step-Mother', 'Step-Parent', TRUE),
    ('Step-Daughter', 'Step-Father', 'Step-Mother', 'Step-Parent', TRUE),
    
    -- Extended family
    ('Grandfather', 'Grandson', 'Granddaughter', 'Grandchild', TRUE),
    ('Grandmother', 'Grandson', 'Granddaughter', 'Grandchild', TRUE),
    ('Grandson', 'Grandfather', 'Grandmother', 'Grandparent', TRUE),
    ('Granddaughter', 'Grandfather', 'Grandmother', 'Grandparent', TRUE),
    ('Grandparent', 'Grandfather', 'Grandmother', 'Grandchild', TRUE),
    ('Grandchild', 'Grandfather', 'Grandmother', 'Grandparent', TRUE),
    
    -- Aunts/Uncles
    ('Uncle', 'Nephew', 'Neice', 'Nibling', TRUE),
    ('Aunt', 'Nephew', 'Neice', 'Nibling', TRUE),
    ('Nephew', 'Uncle', 'Aunt', 'Pibling', TRUE),
    ('Niece', 'Uncle', 'Aunt', 'Pibling', TRUE),
    
    -- Cousins
    ('Cousin', 'Cousin', 'Cousin', 'Cousin', TRUE),
    
    -- In-laws
    ('Mother-in-Law', 'Son-in-Law', 'Daughter-in-Law', 'Child-in-Law', TRUE),
    ('Father-in-Law', 'Son-in-Law', 'Daughter-in-Law', 'Child-in-Law', TRUE),
    ('Son-in-Law', 'Father-in-Law', 'Mother-in-Law', 'Parent-in-Law', TRUE),
    ('Daughter-in-Law', 'Father-in-Law', 'Mother-in-Law', 'Parent-in-Law', TRUE),
    ('Brother-in-Law', 'Brother-in-Law', 'Sister-in-Law', 'Sibling-in-Law', TRUE),
    ('Sister-in-Law', 'Brother-in-Law', 'Sister-in-Law', 'Sibling-in-Law', TRUE),
    
    -- Social
    ('Friend', 'Friend', 'Friend', 'Friend', TRUE),
    ('Best Friend', 'Best Friend', 'Best Friend', 'Best Friend', TRUE),
    ('Acquaintance', 'Acquaintance', 'Acquaintance', 'Acquaintance', TRUE),
    ('Neighbor', 'Neighbor', 'Neighbor', 'Neighbor', TRUE),
    ('Roommate', 'Roommate', 'Roommate', 'Roommate', TRUE),
    
    -- Professional
    ('Manager', 'Direct Report', 'Direct Report', 'Direct Report', TRUE),
    ('Direct Report', 'Manager', 'Manager', 'Manager', TRUE),
    ('Assistant', 'Manager', 'Manager', 'Manager', TRUE),
    ('Colleague', 'Colleague', 'Colleague', 'Colleague', TRUE),
    ('Mentor', 'Mentee', 'Mentee', 'Mentee', TRUE),
    ('Coworker', 'Coworker', 'Coworker', 'Coworker', TRUE),
    ('Boss', 'Employee', 'Employee', 'Employee', TRUE),
    ('Client', 'Service Provider', 'Service Provider', 'Service Provider', TRUE),
    ('Customer', 'Vendor', 'Vendor', 'Vendor', TRUE),
    
    -- Other
    ('Emergency Contact', 'Person', 'Person', 'Person', TRUE),
    ('Referred By', 'Referral', 'Referral', 'Referral', TRUE)
ON CONFLICT (name) DO NOTHING;

-- =============================================================================
-- COMMENTS
-- =============================================================================

COMMENT ON COLUMN users.timezone IS 'IANA Timezone string (e.g., America/New_York)';

COMMENT ON TABLE contacts IS 'Main contacts table with support for partial dates and computed full names';
COMMENT ON COLUMN contacts.full_name IS 'Computed from prefix, given_name, middle_name, family_name, suffix';
COMMENT ON COLUMN contacts.birthday_month IS 'Month (1-12) when full birthday date is unknown';
COMMENT ON COLUMN contacts.birthday_day IS 'Day (1-31) when full birthday date is unknown';
COMMENT ON COLUMN contacts.anniversary_month IS 'Month (1-12) when full anniversary date is unknown';
COMMENT ON COLUMN contacts.anniversary_day IS 'Day (1-31) when full anniversary date is unknown';
COMMENT ON COLUMN contacts.avatar_base64 IS 'Base64 encoded avatar image';
COMMENT ON COLUMN contacts.exclude_from_sync IS 'If true, exclude from CardDAV sync';

COMMENT ON TABLE notification_settings IS 'User preferences for webhook notifications';
COMMENT ON COLUMN notification_settings.notification_time IS 'Time of day to send notifications (HH:MM format)';

COMMENT ON TABLE sessions IS 'Stores active user sessions with device and network information';
COMMENT ON COLUMN sessions.token IS 'Secure token stored in HTTP-only cookie';
COMMENT ON COLUMN sessions.expires_at IS 'When the session expires (typically 24 hours from login)';
COMMENT ON COLUMN sessions.last_activity IS 'Last time the user made a request (updated on each page load)';

COMMENT ON TABLE api_tokens IS 'API tokens for programmatic access to user accounts';
COMMENT ON COLUMN api_tokens.token_hash IS 'SHA-256 hash of the raw token - never store raw tokens';
COMMENT ON COLUMN api_tokens.token_signature IS 'HMAC-SHA256 signature using APP_KEY for additional verification';
COMMENT ON COLUMN api_tokens.name IS 'User-friendly name to identify the token';
COMMENT ON COLUMN api_tokens.last_used_at IS 'Last time this token was used for authentication';
COMMENT ON COLUMN api_tokens.expires_at IS 'Optional expiration date - NULL means no expiration';
COMMENT ON COLUMN api_tokens.is_active IS 'Whether the token is active - allows soft deletion/revocation';

-- =============================================================================
-- VERIFICATION
-- =============================================================================

SELECT 
    'Schema created successfully!' as status,
    (SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public') as table_count,
    (SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public') as column_count,
    (SELECT count(*) FROM relationship_types) as relationship_types_count;
