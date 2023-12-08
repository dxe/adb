[![Build Status](https://travis-ci.org/dxe/adb.svg?branch=master)](https://travis-ci.org/dxe/adb)

# ADB

Activist Database Project

## Local development

### Dependencies

You will need the following to run this project:
 * go
 * node.js v16
 * docker

After installing the above (already installed in the Devcontainer), download all
the go and node dependencies:

```bash
make deps
```

#### Set up mysql locally for development

There is now a Docker container to run MySQL locally, so just run the following:

```bash
docker compose up -d
make dev_db
```

### Run

After downloading the dependencies, start the server:

```bash
make run_all
```

Access the web app at http://localhost:8080.

### JS

This project uses webpack to compile our frontend files. Frontend
files that need to be compiled are in `frontend/`, and the compiled
outputs are in `dist/`.

 * package.json: file with all frontend dependencies

 * webpack.config.js: configuration file for webpack, which builds the js

 * `make watch`: watch the frontend folder for changes and
   automatically build the file if anything changes.

The most convenient workflow is to run `make watch` in one terminal
and `make run` in another one. Then your JS changes will automatically
be built as you edit them.

## Required environment variables for running in prod:

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
