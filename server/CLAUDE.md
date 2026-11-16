# ADB Server instructions

- Avoid passing `sqlx.DB` directly into handler functions. We are migrating to
  repository pattern e.g. `ActivistRepository`. New handlers go in
  `transport/` (HTTP) → `persistence/` (repository), not `main.go`/`model/`.
- Activist-related logic in `model/` is moving to `pkg/activists`.
- `src/main.go` registers all routes and auth middleware, and starts the
  background jobs (`event_sync`, `form_processor`, etc.).
- `members/` is a separate members-portal server on its own port.
