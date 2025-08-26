-- *_updated timestamps may be used to resolve conflicts when merging activist records

-- A sentinel value is used as the default in order to indicate unknown modification times for
-- existing table rows in production.
ALTER TABLE activists
    -- Time when the record was created, or sentinel value if unknown.
    ADD COLUMN created TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    -- Time that phone number was last changed or confirmed, or sentinel value if unknown.
    ADD COLUMN phone_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    -- Time that email was last changed or confirmed, or sentinel value if unknown.
    ADD COLUMN email_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    -- Time that street, city or state was last changed or confirmed, or sentinel value if unknown.
    -- Also indicates when lat, lng coordinates are updated, as they are computed from the address.
    ADD COLUMN address_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01',
    -- Time that any address field (street, city, state) or the 'location' field was last changed or confirmed, or
    -- sentinel value if unknown.
    -- location_updated may be updated separately from address_updated, e.g. when we receive only a zip code from a
    -- petition which may not match the activist's existing street / city / state.
    -- It is updated when the address is updated because it is assumed that the process or user who updates the address
    -- will also verify the location field is up to date as well.
    ADD COLUMN location_updated TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:01';

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
    MODIFY COLUMN location_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
