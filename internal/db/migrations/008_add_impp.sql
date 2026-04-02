CREATE TABLE IF NOT EXISTS impps (
    id SERIAL PRIMARY KEY,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    impp TEXT NOT NULL,
    label_type_id INTEGER REFERENCES contact_label_types(id)
);

INSERT INTO contact_label_types (name, category, is_system) VALUES
-- Instant Messaging and Presence Protocol
-- (name, category, is_system)
('matrix', 'impp', false),
('xmpp', 'impp', false),
('irc', 'impp', false),
('sip', 'impp', false),
('jabber', 'impp', false),
('other', 'impp', true);

ALTER TABLE contacts ADD COLUMN IF NOT EXISTS timezone TEXT DEFAULT 'UTC';


INSERT INTO relationship_types (name, reverse_name_male, reverse_name_female, reverse_name_neutral, is_system) VALUES
-- Half-Siblings
('Half-Brother', 'Half-Brother', 'Half-Sister', 'Half-Sibling', TRUE),
('Half-Sister', 'Half-Brother', 'Half-Sister', 'Half-Sibling', TRUE),
('Half-Sibling', 'Half-Brother', 'Half-Sister', 'Half-Sibling', TRUE);