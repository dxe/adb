# adb
Activist Database Project

# Run

First, download all the needed dependencies. Then start the server by running `make run_all` and going to localhost:8080.

# Dependencies

You will need the following to run this project:

 * go
 * node
 * mysql

After installing the above, download all the go and node dependencies by running `make deps`.

## Set up mysql locally for development

First, install mysql server. Then, create a user and database like this:

```
CREATE USER adb_user@localhost IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to adb_user@localhost;
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE DATABASE adb_test_db CHARACTER SET utf8 COLLATE utf8_general_ci;
```

Then run `make dev_db`.

# JS

This project uses webpack to compile our frontend files. Frontend
files that need to be compiled are in `frontend/`, and the compiled
outputs are in `dist/`.

 * package.json: file with all frontend dependencies

Run `npm install --save some-module-name` to install a new dependency.

 * webpack.config.js: configuration file for webpack, which builds the js

If you want to add a new page, you'll have to add it as an entry in
webpack.js.config.

 * `make watch`: watch the frontend folder for changes and
   automatically build the file if anything changes.

The most convenient workflow is to run `make watch` in one terminal
and `make run` in another one. Then your JS changes will automatically
be built as you edit them.

# Deploy

To deploy, you need a user account that has the "adb" group and also has passwordless sudo enabled.

The server uses daemontools on the server to run. See the Makefile for more info on how to deploy.

## How To Give Someone New Deploy Access

First, make sure they're trusted because they will have sudo access to the machine.

```
NEW_USER=their-username
sudo adduser --disabled-password $NEW_USER
sudo usermod -a -G sudo $NEW_USER

# Add public key
sudo mkdir /home/$NEW_USER/.ssh
sudo vim /home/$NEW_USER/.ssh/authorized_keys
sudo chown -R $NEW_USER:$NEW_USER /home/$NEW_USER/.ssh
```

Edit /etc/sudoers so the username has passwordless sudo. Use the command `sudo visudo` to edit it.

```
their-username ALL=(ALL) NOPASSWD:ALL
```

Finally, add a deploy target to the Makefile.
