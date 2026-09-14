# ADB Server instructions

- Avoid passing `sqlx.DB` directly into handler functions. We are migrating to
repository pattern e.g. `ActivistRepository`.
