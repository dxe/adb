CREATE USER 'adb_user'@'%' IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to 'adb_user'@'%';
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE DATABASE adb_test_db CHARACTER SET utf8 COLLATE utf8_general_ci;
