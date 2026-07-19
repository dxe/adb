-- Richer fields to support creating events in advance. All default to
-- empty/null so existing attendance rows remain valid.
--
-- Location is stored denormalized directly on the event: a free-text name
-- (always editable, never shared between events) plus optional geo data — a
-- Google Place id and/or coordinates. There is intentionally no shared
-- locations table: a place's display name belongs to the event, so fixing a
-- typo or renaming one event's location can never affect another.
ALTER TABLE events
  ADD COLUMN is_online                TINYINT(1)    NOT NULL DEFAULT 0,
  ADD COLUMN description              TEXT,
  ADD COLUMN start_time               TIME          NULL,
  ADD COLUMN end_time                 TIME          NULL,
  ADD COLUMN timezone                 VARCHAR(64)   NOT NULL DEFAULT '',
  ADD COLUMN is_public                TINYINT(1)    NOT NULL DEFAULT 0,
  ADD COLUMN location_name            VARCHAR(255)  NOT NULL DEFAULT '',
  ADD COLUMN location_address         VARCHAR(512)  NOT NULL DEFAULT '',
  ADD COLUMN location_google_place_id VARCHAR(255)  NOT NULL DEFAULT '',
  ADD COLUMN location_lat             DECIMAL(10, 7) NULL,
  ADD COLUMN location_lng             DECIMAL(10, 7) NULL;
