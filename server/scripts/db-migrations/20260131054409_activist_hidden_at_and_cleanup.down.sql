-- Re-add preferred_name_updated column
ALTER TABLE activists
    ADD COLUMN preferred_name_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01';

ALTER TABLE activists
    MODIFY COLUMN preferred_name_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Remove hidden_at column
ALTER TABLE activists
    DROP COLUMN hidden_at;
