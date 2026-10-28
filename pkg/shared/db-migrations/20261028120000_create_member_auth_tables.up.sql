-- Sessions and single-use login tokens for member sign-in (Google or emailed
-- magic link) on members.directactioneverywhere.com. Only SHA-256 hashes of
-- the random tokens are stored; the raw values live in the cookie / email.

CREATE TABLE member_sessions (
  id_hash BINARY(32) NOT NULL,
  email VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  last_seen_at DATETIME NOT NULL,
  PRIMARY KEY (id_hash),
  INDEX member_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
