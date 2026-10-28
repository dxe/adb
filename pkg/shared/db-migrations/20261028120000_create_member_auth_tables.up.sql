-- Single-use login tokens for emailed magic links on
-- members.directactioneverywhere.com. Only the SHA-256 hash of the random token
-- is stored; the raw value lives in the email. (Sessions are stateless signed
-- cookies, so there is no sessions table.)

CREATE TABLE member_login_tokens (
  token_hash BINARY(32) NOT NULL,
  email VARCHAR(255) NOT NULL,
  next_url VARCHAR(512) NOT NULL,
  created_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  used_at DATETIME NULL,
  PRIMARY KEY (token_hash),
  INDEX member_login_tokens_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
