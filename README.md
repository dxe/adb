# adb
Activist Database Project

# Run

First, download all the needed dependencies. Then start the server by running `go run main.go` and going to localhost:8080.

# Deploy

NOTE: This project is under active development and shouldn't be deployed in its current state.

# Set up mysql locally for development

First, install mysql server. Then, create a user and database like this:

```
CREATE USER adb_user@localhost IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to adb_user@localhost;
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8 COLLATE utf8_general_ci;
```
