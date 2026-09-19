-- Add the activist's preferred and alternate contact methods.
-- Values are the `ContactMethod` enum in pkg/activists; 0 means "not set".
-- The API exchanges labels, which are mapped to these values by the server.

ALTER TABLE activists
    ADD COLUMN preferred_contact_method TINYINT UNSIGNED NOT NULL DEFAULT 0,
    ADD COLUMN alternate_contact_method TINYINT UNSIGNED NOT NULL DEFAULT 0;
