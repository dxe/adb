-- Add name_updated timestamp column
-- This tracks when name or preferred_name fields were last modified
-- A sentinel value is used as the default for existing records to indicate unknown modification times

ALTER TABLE activists
    -- Time that name was last changed, or sentinel value if unknown
    ADD COLUMN name_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01';

-- Change default for new records to CURRENT_TIMESTAMP
ALTER TABLE activists
    MODIFY COLUMN name_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
