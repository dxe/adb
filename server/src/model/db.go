package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/pkg/shared"
)

func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

func newTestDB() *sqlx.DB {
	shared.WipeDatabase(config.DBTestDataSource()+"&multiStatements=true", false)
	return NewDB(config.DBTestDataSource())
}
