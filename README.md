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

## Required enviornment variables for running in prod:

- ADB_URL_PATH: For example, "http://adb.domain.com"
- PORT: The port to run the webserver on
- MEMBERS_PORT: Port to run the Members webserver on
- DB_USER
- DB_PASSWORD
- DB_NAME
- DB_PROTOCOL: For example, "tcp([host]:[port])"
- PROD: [true or false]
- RUN_BACKGROUND_JOBS: [true or false] should only be true on at most one instance if load balancing
- COOKIE_SECRET: [a random string]
- CSRF_AUTH_KEY: [a random string]

## Optional environment variables

### For signing people up to DxE's main mailing list & chapter-specific mailing lists (please reach out to tech@dxe.io to get an API key to sign people up)
- SIGNUP_ENDPOINT
- SIGNUP_KEY

### For syncing with a chapter's internal Google Groups (for example, working group lists):
- SYNC_MAILING_LISTS_CONFIG_FILE: relative path to client_secrets.json if syncing with google groups
- SYNC_MAILING_LISTS_OAUTH_SUBJECT: google account to use to sync

### For sending emails via SMTP:
- SMTP_HOST
- SMTP_PORT  
- SMTP_USER
- SMTP_PASSWORD

### For sending surveys to event attendees:
- SURVEY_FROM_EMAIL
- SURVEY_MISSING_EMAIL: Email to send survey errors to

### Google Cloud client ID/secret for the Members page:
- MEMBERS_CLIENT_ID
- MEMBERS_CLIENT_SECRET

### ipgeolocation.io key for finding nearby upcoming events based on a user's IP address (used w/ public-facing API):
- IPGEOLOCATION_KEY

### Discord config for verifying accounts:
- DISCORD_SECRET
- DISCORD_BOT_BASE_URL
- DISCORD_FROM_EMAIL
- DISCORD_MODERATOR_EMAIL

### Google Places API Key for finding city information on public-facing forms
- GOOGLE_PLACES_API_KEY
