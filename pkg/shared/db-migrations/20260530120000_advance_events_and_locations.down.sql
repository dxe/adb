ALTER TABLE events
  DROP FOREIGN KEY fk_events_location,
  DROP COLUMN location_id,
  DROP COLUMN is_online,
  DROP COLUMN description,
  DROP COLUMN start_time,
  DROP COLUMN end_time,
  DROP COLUMN timezone,
  DROP COLUMN is_public;

DROP TABLE locations;
