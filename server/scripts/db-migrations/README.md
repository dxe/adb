# Database migrations

This project uses [`golang-migrate`](https://github.com/golang-migrate/migrate) for migrations.

The `make deps` command installs the CLI and makes the `migrate` command available.

## Executing migrations

ADB will automatically apply all migrations when it starts.

To apply migrations manually, run:

```bash
migrate -source "file://server/scripts/db-migrations/" -database "mysql://$DB_USER:$DB_PASSWORD@$DB_PROTOCOL/$DB_NAME?parseTime=true&charset=utf8mb4&multiStatements=true" up
```

If using the devcontainer, note that the environment variables should already be set.

The database flag accepts the value we pass to `migrate.New` in the codebase.

## Creating migrations

To create a new migration, create a pair of blank, timestamped "up"/"down" files, replacing `<name>` with a lower snake
case name for the migration such as `add_activist_middle_name_column.

```bash
migrate create -dir server/scripts/db-migrations/ -ext sql <name>
```

Then write the migration code in the SQL files.

The "up" file should contain the main migration code, and the "down" file should undo what the "up" file does, allowing
reverting to a previous version of the schema. This is useful, for example, in case there is a bug in production that
requires a rollback, or to test a previous version during development.
