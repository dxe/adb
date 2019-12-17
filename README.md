[![Build Status](https://travis-ci.org/dxe/adb.svg?branch=master)](https://travis-ci.org/dxe/adb)

# adb
Activist Database Project

## Run

First, download all the needed dependencies. Then start the server by running `make run_all` and going to localhost:8080.

## Dependencies

You will need the following to run this project:

 * go
 * node
 * mysql

After installing the above, download all the go and node dependencies by running `make deps`.

### Set up mysql locally for development

First, install mysql server. Then, create a user and database like this:

```
CREATE USER adb_user@localhost IDENTIFIED BY 'adbpassword';
GRANT ALL PRIVILEGES ON *.* to adb_user@localhost;
FLUSH PRIVILEGES;

CREATE DATABASE adb_db CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE DATABASE adb_test_db CHARACTER SET utf8 COLLATE utf8_general_ci;
```

Then run `make dev_db`.

### Environment variables required for surveys to be sent
- AWS_ACCESS_KEY_ID
- AWS_SECRET_KEY
- AWS_SES_ENDPOINT (example: https://email.us-west-2.amazonaws.com)
- SURVEY_FROM_EMAIL (address surveys should be sent from)
- SURVEY_MISSING_EMAIL (address to alert is survey recipients are missing email address)

## JS

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
