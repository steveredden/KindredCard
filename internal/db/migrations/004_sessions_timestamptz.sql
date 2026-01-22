-- Ensure the session has UTC as the base for this migration
SET TIME ZONE 'UTC';

-- Convert the column. 
-- USING clause ensures that the 'wall clock' time 07:06:00 is 
-- interpreted as 07:06:00+00 (UTC) rather than using the system's local time.
ALTER TABLE sessions 
    ALTER COLUMN last_activity TYPE TIMESTAMPTZ 
    USING last_activity AT TIME ZONE 'UTC';

-- Ensure any future default values are UTC-aware
ALTER TABLE sessions 
    ALTER COLUMN last_activity SET DEFAULT CURRENT_TIMESTAMP;