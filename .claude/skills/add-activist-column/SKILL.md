---
name: add-activist-column
description: Add, rename, or remove a column on the activists table end to end.
---

# Add an activist column

A column has to be registered in several places.

The `server/src/model` package is mid-refactor into `pkg/activists`, so this
skill avoids naming specific files.

## 1. Migration

Follow [pkg/shared/db-migrations/README.md](pkg/shared/db-migrations/README.md).
Prefer `NOT NULL DEFAULT ''` / `0` over a nullable column unless the field
genuinely needs "unset" distinct from empty.

## 2. `pkg/activists/`

- `types.go` — `Col<Name> ActivistColumnName = "<api_name>"`, plus the field on
  the `Activist` struct with a `db:"<db_col>"` tag (`sql.NullString` if nullable).
- `columns.go` — entry in `simpleColumns`. A computed or joined column needs
  `alias` (so the output column name matches) and `joins` instead.
- `columns_meta.go` — entry in `ActivistColumns`: `Setter`, `DbCol`,
  `UserPatchable: true` if editable via PATCH, `Nullable: true` if NULL-able,
  `BumpTimestamps` if a companion `<x>_updated` column should be touched.

`ActivistColumnName` (API name) and `DbCol` (SQL name) are deliberately separate
even though they usually match.

## 3. `server/src/model/`

- Alias constant re-exporting `activists.Col<Name>` (`grep -rn 'ColPronouns' server/src/model/`).
- Raw `SELECT` list — add `a.<db_col>` (`grep -rn 'a\.pronouns' server/src/model/`).
- `ActivistUserEditableDataFieldAssignments` — add `<db_col> = :<db_col>` if editable.
- `ActivistJSON` struct field + json tag, and the mappings in `BuildActivistJSON`
  and `CleanActivistData` (`strings.TrimSpace` for strings).
- `getMergeActivistWinner` — `stringMerge` / equivalent, so the value survives an
  activist merge.
- `validateActivistUpdate` — only if the value is constrained. Note the existing
  checks run only when the value _changed_, so pre-existing bad data is grandfathered.
- Only if the column is also collected by the public application form: `forms.go`.

## 4. `server/src/transport/activists.go`

Patchable columns need a `*string` (or matching pointer) field on
`ActivistPatchInput` plus an `addString`/`addNullableString` line in `ToPatchData`.

## 5. `frontend-v2/`

- `src/lib/api/activists.ts` — add to the `ActivistJSON` zod object; the
  `ActivistColumnName` enum is derived from its keys. Add to `ActivistPatchInput`
  iff editable: that schema is the **only** thing that makes a field editable in
  the UI.
- `src/app/(authed)/activists/column-definitions.ts` — `COLUMN_DEFINITIONS` entry.
  See the `ColumnDefinition` interface for the options; the easy ones to forget are
  `editInputType` + `editOptions` (non-text inputs), `isDate`, `linkType`, and
  `blankValue` for `false`/`0` values that Go's `omitempty` drops.
- Shared enum value lists live in `filter-types.ts`. If the backend also validates
  the set, the two lists are synced by hand — cross-reference them in comments.
- Adding a column does **not** make it filterable. Filters are a separate system
  (`ActivistFilters` in `pkg/activists/filters.go`, `filter_impl.go`, and the
  frontend filter UI).

## 6. Test

Add the column to `TestPatchActivist_UpdatesAllPatchableFields`
(`server/src/persistence/activist_patch_test.go`) — it is the de facto check that
struct tags, `DbCol`, and the UPDATE statement agree. Nothing enforces
exhaustiveness, so an omission here is silent.

## 7. Verify

Build/lint as usual.

## If this skill is stale

If the greps above return nothing, or the structures described have moved or been
renamed, say so to the user, name the step that drifted, and offer to update this
file.
