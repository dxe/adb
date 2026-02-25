-- Add hidden_updated timestamp to track when activists' hidden state was last changed
-- This enables sync of deletions/hides via incremental API

-- Add hidden_updated column with sentinel value for existing hidden records
ALTER TABLE activists
    ADD COLUMN hidden_updated TIMESTAMP NULL DEFAULT NULL;

-- Set hidden_updated for currently hidden activists to sentinel value
-- (we don't know when they were actually hidden)
UPDATE activists
SET hidden_updated = '1970-01-01 00:00:01'
WHERE hidden = 1 AND hidden_updated IS NULL;
