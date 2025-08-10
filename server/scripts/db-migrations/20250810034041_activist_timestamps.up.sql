-- *_updated timestamps may be used to resolve conflicts when merging activist records

-- A sentinel value is used as the default in order to indicate unknown modification times for
-- existing table rows in production.
ALTER TABLE activists
    ADD COLUMN created TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    ADD COLUMN phone_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    ADD COLUMN email_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    ADD COLUMN address_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    ADD COLUMN coords_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01';

-- Change default for new records to CURRENT_TIMESTAMP.
ALTER TABLE activists
    MODIFY COLUMN created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE activists
    MODIFY COLUMN phone_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE activists
    MODIFY COLUMN email_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE activists
    MODIFY COLUMN address_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE activists
    MODIFY COLUMN coords_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
