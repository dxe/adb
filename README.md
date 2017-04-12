# adb
Activist Database Project

# Run

First, download all the needed dependencies. Then start the server by running `go run main.go` and going to localhost:8080.

# Deploy

To deploy, you need a user account that has the "adb" group and also has passwordless sudo enabled.

The server uses daemontools on the server to run. See the Makefile for more info on how to deploy.

# Set up mysql locally for development

First, install mysql server. Then, create a user and database like this:

```
CREATE USER adb_user@localhost IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to adb_user@localhost;
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE DATABASE adb_test_db CHARACTER SET utf8 COLLATE utf8_general_ci;
```

Then run `make dev_db`.
