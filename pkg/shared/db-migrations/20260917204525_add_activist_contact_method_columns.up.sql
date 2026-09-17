-- Add the activist's preferred and alternate contact methods.
-- Values are restricted by the API to a fixed set (see `ValidContactMethods` in
-- server); the empty string means "not set".

ALTER TABLE activists
    ADD COLUMN preferred_contact_method VARCHAR(40) NOT NULL DEFAULT '',
    ADD COLUMN alternate_contact_method VARCHAR(40) NOT NULL DEFAULT '';
