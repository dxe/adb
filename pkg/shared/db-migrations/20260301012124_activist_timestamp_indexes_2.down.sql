DROP INDEX idx_name_updated ON activists;
CREATE INDEX idx_name_updated ON activists(name_updated);

DROP INDEX idx_email_updated ON activists;
CREATE INDEX idx_email_updated ON activists(email_updated);

DROP INDEX idx_phone_updated ON activists;
CREATE INDEX idx_phone_updated ON activists(phone_updated);
