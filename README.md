# ADB

Activist Database Project

## Local development

### Dependencies

The following dependencies are required to run this project and will be already
installed if using the devcontainer:

- go
- nvm
- docker

Running this command is required to download all the go and node dependencies:

```bash
make deps
```

#### Set up mysql locally for development

If you are not using the devcontainer, you can use our Docker Compose configuration
to run MySQL locally:

```bash
( cd server/ && docker compose up -d )
make dev_db
```

If you are using the devcontainer, just run this command in the container:

```bash
make dev_db
```

You can log into the database from the devcontainer by just typing `db` in the terminal (an alias in .bashrc).

### Run

After downloading the dependencies, start the development servers. There are two
ways to do this.

The first way is using the configuration in Makefile. This runs the React app
in the background and then runs the Go server.

```bash
# Convenient but does not allow viewing server output separately.
make run_all
```

The other way is using VS Code launch configurations. This allows you to see
the output of each process individually via the debug console, and to restart
processes individually. To do this, go to the Run and Debug panel and choose
`Launch all`.

Access the new frontend (React app) at http://localhost:8080/v2 - or test with
http://localhost:8080/v2/test.

React pages auto-update in the frontend development server when you save
changes.

Note: The react app can technically be accessed directly on port 3000 locally,
but it won't have access to the cookies from :8080 which could cause issues.

Access the Go API server and old frontend (Vue app) at http://localhost:8080/

Access the Members app at members.dxesf.org at http://localhost:8081/

### Code formatting

Please run `make fmt` before sending a pull request.

Note that `prettier` is configured at the root of the repo and not within individual projects, although project-specifc
configuration can be applied in the configuration file. This allows formatting the repo with one invocation of prettier,
and is necessary as VS Code only accepts one prettier configuration. It also ensures consistency by avoiding multiple
versions or configurations of prettier.

### Test

Please run `make test` before sending a pull request.

Note the tests take several minutes to run, even though some of them
finish quickly. To check if any tests are hanging, you can add "-v -p 1" after
"go test" args in the Makefile to show individual test names and to run them
in serial.

### JS

The frontend is being migrated from Vue to React and is split into two
applications: the Vue app in /frontend and the React app in /frontend-v2.

#### Vue app

This project uses webpack to compile our frontend files. Frontend
files that need to be compiled are in `frontend/`, and the compiled
outputs are in `frontend/dist/`.

- package.json: file with all frontend dependencies

- webpack.config.js: configuration file for webpack, which builds the js

- `make watch`: watch the frontend folder for changes and
  automatically build the file if anything changes.

The most convenient workflow is to run `make watch` in one terminal
and `make run` in another one. Then your JS changes will automatically
be built as you edit them.

#### React app

See frontend-v2/README.md for more information on the React app. See above for
instructions on building and running the React app along with other components
of ADB.

## Environment variables

### Required environment variables

These variables are always required in prod:

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

### For signing people up to DxE's main mailing list & chapter-specific mailing lists

Please reach out to tech@dxe.io to get an API key to sign people up.

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

### Google Places API Key for finding city information on public-facing forms

- GOOGLE_PLACES_API_KEY

## Logging and Monitoring

In production, logs are collected in CloudWatch under a log group called `adb`.

The log messages are a mix of English strings often generated by our use of Go's `log` package, and JSON-encoded strings
generated by our use of the `github.com/rs/zerolog` package.

JSON log lines have a property called `level` which when set to `error` is captured by a "Metric filter" called
`error_logs` on the `adb` CloudWatch log group. When any log item matches this filter, our CloudWatch alarm is triggerd
and the tech team is notified.

## Authentication

Note: authentication is bypassed during development in the server's `model.authADBUser` function.

Authentication is provided by Google Identity's 'Sign in with Google' for Web:
https://developers.google.com/identity/gsi/web/guides/overview

This makes use of our OAuth 2.0 client ID in Google Cloud Platform. However many OAuth 2.0 details are taken care of
for us by Google.

Pages protected by authentication check the cookies to see if a valid session exists and redirect to `/v2/login` which
serves up `frontend-v2/src/app/login/page.tsx`. This page handles authentication with Google client-side to get a token
and then makes a POST request to ADB's `/tokensignin` endpoint with the token. The backend endpoint verifies the token,
obtains the user indicated by the token and creates a session cookie. If the backend returns a successful response,
the login page then redirects to the homepage.

For local testing, you may need to add an ngrok authorized origin in GCP console. The origin `http://localhost:8080`
does not seem to work as of December 2025.
