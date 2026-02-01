-- Indexes to optimize timestamp-based filtering
-- The chapter_hidden index reduces the initial scan set
-- Individual timestamp indexes allow MySQL to use index merge for OR conditions

CREATE INDEX idx_chapter_hidden ON activists(chapter_id, hidden);
CREATE INDEX idx_name_updated ON activists(name_updated);
CREATE INDEX idx_email_updated ON activists(email_updated);
CREATE INDEX idx_phone_updated ON activists(phone_updated);
