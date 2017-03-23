# adb
Activist Database Project

# Run

First, download all the needed dependencies. Then start the server by running `go run main.go` and going to localhost:8080.

# Deploy

NOTE: This project is under active development and shouldn't be deployed in its current state.

The activist database is deployed using Dokku. To deploy it, first add the dokku project "adb" as a remote repo:

```
git remote add dokku dokku@dxetech.org:adb
```

Then, push your changes with `git push dokku master`.
