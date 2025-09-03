CREATE USER 'adb_user'@'%' IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to 'adb_user'@'%';
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE adb_test_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
