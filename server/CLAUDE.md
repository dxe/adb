# ADB Server instructions

- Avoid passing `sqlx.DB` directly into handler functions. We are migrating to
  repository pattern e.g. `ActivistRepository`. New handlers go in
  `transport/` (HTTP) → `persistence/` (repository), not `main.go`/`model/`.
- Activist-related logic in `model/` is moving to `pkg/activists`.
- `src/main.go` registers all routes and auth middleware, and starts the
  background jobs (`event_sync`, `form_processor`, etc.).
- `members/` is a separate members-portal server on its own port.

## Layering and authorization

- `transport/` only parses the request and maps errors to HTTP statuses:
  `model.ErrValidation` → 400, `model.ErrNotFound` → 404, anything else → 500.
  Don't do authorization there.
- Authorization lives in the business layer (`model/`). Admins can do anything
  in any chapter, and organizers' access should be limited to their chapter. 
  Check `UserHasOrganizerAccess` and `CheckChapterAccess(user, chapterID)` 
  (admin, or member of that chapter) and return `ValidationErrorf(...)`.
- `persistence/` functions should have redundant auth checks in the SQL to
  avoid security issues. Keep the code simple because this redundant code path
  should never be called.
- New handlers take the authed user as a parameter and a repository interface.

## API conventions

- JSON endpoints should be registered under `/api/...`
