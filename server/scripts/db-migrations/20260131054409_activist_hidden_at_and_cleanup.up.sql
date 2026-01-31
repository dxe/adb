-- Add hidden_at timestamp to track when activists are hidden/deleted
-- This enables sync of deletions/hides via incremental API

-- Add hidden_at column with sentinel value for existing hidden records
ALTER TABLE activists
    ADD COLUMN hidden_at TIMESTAMP NULL DEFAULT NULL;

-- Set hidden_at for currently hidden activists to sentinel value
-- (we don't know when they were actually hidden)
UPDATE activists
SET hidden_at = '1970-01-01 00:00:01'
WHERE hidden = 1 AND hidden_at IS NULL;
