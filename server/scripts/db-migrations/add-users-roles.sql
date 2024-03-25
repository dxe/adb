
CREATE TABLE users_roles (
  user_id INT NOT NULL,
  role VARCHAR(45) NOT NULL,
  PRIMARY KEY (user_id, role),
  CONSTRAINT user_id
    FOREIGN KEY (user_id)
    REFERENCES adb_users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);

INSERT INTO users_roles (user_id, role)
SELECT id, 'admin' FROM adb_users WHERE admin = 1;
