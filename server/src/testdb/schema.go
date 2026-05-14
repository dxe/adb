package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/pkg/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var initializeSchemaOnce sync.Once

func InitializeSchema() {
	initializeSchemaOnce.Do(func() {
		shared.WipeDatabase(config.DBTestDataSource()+"&multiStatements=true", false)
	})
}

func listResettableTables(ctx context.Context, conn *sql.Conn) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_type = 'BASE TABLE'
  AND table_name <> 'schema_migrations'
ORDER BY table_name
`)
	if err != nil {
		return nil, fmt.Errorf("query resettable tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan resettable table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resettable tables: %w", err)
	}

	return tables, nil
}

func Reset(db *sqlx.DB) error {
	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get test database connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("disable foreign key checks: %w", err)
	}
	defer func() {
		if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
			fmt.Printf("warning: failed to re-enable foreign key checks: %v\n", err)
		}
	}()

	tables, err := listResettableTables(ctx, conn)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE `%s`", table)); err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}

	// Prevent tests that create chapters from unintentionally claiming the
	// special SF Bay chapter.
	if err := seedSFBayChapter(ctx, conn); err != nil {
		return err
	}

	if err := seedDevUser(ctx, conn); err != nil {
		return err
	}

	return nil
}

// seedSFBayChapter inserts the SF Bay chapter so the first auto-generated
// chapter_id reliably equals shared.SFBayChapterIdDevTest. Tests reference
// this constant; without the seed, the next InsertChapter call from a test
// would unintentionally claim chapter_id = 1.
func seedSFBayChapter(ctx context.Context, conn *sql.Conn) error {
	res, err := conn.ExecContext(ctx, "INSERT INTO fb_pages (name, fb_url) VALUES (?, '')", shared.SFBayChapterName)
	if err != nil {
		return fmt.Errorf("seed SF Bay chapter: %w", err)
	}
	chapterID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("read seeded chapter id: %w", err)
	}
	if chapterID != shared.SFBayChapterIdDevTest {
		return fmt.Errorf("seeded SF Bay chapter_id = %d, expected %d", chapterID, shared.SFBayChapterIdDevTest)
	}
	return nil
}

// seedDevUser inserts the dev/test user with organizer role in the SF Bay
// chapter so tests have a ready-made authed user to act as. Its id matches
// shared.DevTestUserId, mirroring the dev login behavior in main.go.
func seedDevUser(ctx context.Context, conn *sql.Conn) error {
	res, err := conn.ExecContext(ctx,
		"INSERT INTO adb_users (email, name, disabled, chapter_id) VALUES (?, ?, false, ?)",
		shared.DevTestUserEmail, "Dev User", shared.SFBayChapterIdDevTest,
	)
	if err != nil {
		return fmt.Errorf("seed dev user: %w", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("read seeded dev user id: %w", err)
	}
	if userID != shared.DevTestUserId {
		return fmt.Errorf("seeded dev user id = %d, expected %d", userID, shared.DevTestUserId)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO users_roles (user_id, role) VALUES (?, ?)",
		userID, shared.RoleOrganizer,
	); err != nil {
		return fmt.Errorf("seed dev user role: %w", err)
	}
	return nil
}

func NewDB() *sqlx.DB {
	InitializeSchema()

	db, err := sqlx.Open("mysql", config.DBTestDataSource())
	if err != nil {
		panic(err)
	}
	if err := Reset(db); err != nil {
		_ = db.Close()
		panic(fmt.Errorf("reset test database: %w", err))
	}

	return db
}
