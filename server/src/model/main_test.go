package model

import (
	"os"
	"testing"

	"github.com/dxe/adb/testdb"
)

func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithMySQLContainer(m))
}
