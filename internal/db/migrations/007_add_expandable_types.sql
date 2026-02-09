-- 1. Create the lookup table
CREATE TABLE IF NOT EXISTS contact_label_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL, -- 'phone', 'email', 'address', 'url', 'date'
    is_system BOOLEAN DEFAULT false,
    UNIQUE(name, category)
);

-- 2. Seed RFC Standard Values (vCard 3.0 & 4.0)
INSERT INTO contact_label_types (name, category, is_system) VALUES
-- Phone Types
('cell', 'phone', true),
('home', 'phone', true),
('work', 'phone', true),
('fax', 'phone', true),
('pager', 'phone', true),
('voice', 'phone', true),
('text', 'phone', true),
-- custom phones
('work cell', 'phone', false),

-- Email Types
('home', 'email', true),
('work', 'email', true),
('other', 'email', true),

-- Address Types
('home', 'address', true),
('work', 'address', true),

-- URL Types
('home', 'url', true),
('work', 'url', true),
('blog', 'url', true),
('profile', 'url', true),
-- custom urls
('immich', 'url', false);

-- 3. Modify existing tables
ALTER TABLE phones ADD COLUMN IF NOT EXISTS label_type_id INTEGER REFERENCES contact_label_types(id);
ALTER TABLE emails ADD COLUMN IF NOT EXISTS label_type_id INTEGER REFERENCES contact_label_types(id);
ALTER TABLE addresses ADD COLUMN IF NOT EXISTS label_type_id INTEGER REFERENCES contact_label_types(id);
ALTER TABLE urls ADD COLUMN IF NOT EXISTS label_type_id INTEGER REFERENCES contact_label_types(id);

-- 4. Data Migration Logic
-- This attempts to map existing array values to the new ID system
-- Note: This handles only the first item in the old array
UPDATE phones 
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = phones.type[1] AND category = 'phone')
WHERE type IS NOT NULL AND array_length(type, 1) > 0;
UPDATE emails
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = emails.type[1] AND category = 'email')
WHERE type IS NOT NULL AND array_length(type, 1) > 0;
UPDATE addresses
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = addresses.type[1] AND category = 'address')
WHERE type IS NOT NULL AND array_length(type, 1) > 0;
UPDATE urls
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = urls.type[1] AND category = 'url')
WHERE type IS NOT NULL AND array_length(type, 1) > 0;

-- now do defaults for anything missing
UPDATE phones
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = 'cell' AND category = 'phone')
WHERE label_type_id IS NULL;
UPDATE emails
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = 'home' AND category = 'email')
WHERE label_type_id IS NULL;
UPDATE addresses
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = 'home' AND category = 'address')
WHERE label_type_id IS NULL;
UPDATE urls
SET label_type_id = (SELECT id FROM contact_label_types WHERE name = 'home' AND category = 'url')
WHERE label_type_id IS NULL;

-- 5. Drop the old array columns after verifying data
ALTER TABLE phones DROP COLUMN type;
ALTER TABLE emails DROP COLUMN type;
ALTER TABLE addresses DROP COLUMN type;
ALTER TABLE urls DROP COLUMN type;

-- 6. Add additional X fields to contacts
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS maiden_name VARCHAR(100); -- X-MAIDENNAME
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS phonetic_first_name VARCHAR(100); -- X-PHONETIC-FIRST-NAME
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS pronunciation_first_name VARCHAR(100); -- X-PRONUNCIATION-FIRST-NAME
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS phonetic_middle_name VARCHAR(100); -- X-PHONETIC-MIDDLE-NAME
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS phonetic_last_name VARCHAR(100); -- X-PHONETIC-LAST-NAME
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS pronunciation_last_name VARCHAR(100); -- X-PRONUNCIATION-LAST-NAME

-- 7. Add additional X field (X-PHONETIC-ORG) to organizations
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS phonetic_name VARCHAR(100);