package members

import (
	"os"
	"testing"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/testdb"
)

func TestMain(m *testing.M) {
	// Tests run like local development (server/debug.env sets this).
	config.MembersDevLogin = true
	os.Exit(testdb.RunWithMySQLContainer(m))
}
