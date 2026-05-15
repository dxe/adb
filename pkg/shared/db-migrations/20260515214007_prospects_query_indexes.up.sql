CREATE INDEX idx_interactions_activist_timestamp ON interactions(activist_id, timestamp);

CREATE INDEX idx_chapter_level ON activists(chapter_id, activist_level);
