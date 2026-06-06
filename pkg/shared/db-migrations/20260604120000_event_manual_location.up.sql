-- Manual location for events that aren't a Google Place (e.g. an intersection
-- or a patch of public land). Stored directly on the event rather than in the
-- place-keyed locations table, since there's no google_place_id to dedupe on.
-- All nullable/empty so existing rows are unaffected; an event uses either a
-- location_id (Google Place) or these manual fields, never both.
ALTER TABLE events
  ADD COLUMN manual_location_name VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN manual_location_lat  DECIMAL(10, 7) NULL,
  ADD COLUMN manual_location_lng  DECIMAL(10, 7) NULL;
