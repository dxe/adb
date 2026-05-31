-- Locations are an internal, deduped table keyed by (chapter, Google place).
-- The concept is not exposed in the UI; the user just searches an address.
CREATE TABLE locations (
  id                INT NOT NULL AUTO_INCREMENT,
  chapter_id        INT NOT NULL,
  google_place_id   VARCHAR(255) NOT NULL DEFAULT '',
  name              VARCHAR(255) NOT NULL DEFAULT '',
  formatted_address VARCHAR(512) NOT NULL DEFAULT '',
  lat               DECIMAL(10, 7) NULL,
  lng               DECIMAL(10, 7) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uniq_chapter_place (chapter_id, google_place_id)
);

-- Richer fields to support creating events in advance. All default to
-- empty/null so existing attendance rows remain valid.
ALTER TABLE events
  ADD COLUMN location_id INT NULL,
  ADD COLUMN is_online   TINYINT(1) NOT NULL DEFAULT 0,
  ADD COLUMN description TEXT,
  ADD COLUMN start_time  TIME NULL,
  ADD COLUMN end_time    TIME NULL,
  ADD COLUMN timezone    VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN is_public   TINYINT(1) NOT NULL DEFAULT 0,
  ADD CONSTRAINT fk_events_location FOREIGN KEY (location_id) REFERENCES locations (id);
