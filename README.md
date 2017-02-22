# adb
Activist Database Project

# Deploy

The activist database is deployed using Dokku. To deploy it, first add the dokku project "adb" as a remote repo:

```
git remote add dokku dokku@dxetech.org:adb
```

Then, push your changes with `git push dokku master`.

# Run locally

Run the following from your terminal:

```
python -m SimpleHTTPServer 9934
```

And navigate to localhost:9934/event-new.html.
