package config

import (
	"strings"
	"testing"

	"github.com/dxe/adb/pkg/shared"
)

func clearDBEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"DB_USER", "DB_PASSWORD", "DB_PROTOCOL", "DB_NAME", "PROD"} {
		t.Setenv(k, "")
	}
}

func TestDBDataSource_DevDefaults(t *testing.T) {
	clearDBEnv(t)

	got := DBDataSource()
	want := shared.BuildDBDataSource(
		shared.DevDBUser, shared.DevDBPassword, shared.DevDBProtocol, shared.DevDBName, false)
	if got != want {
		t.Errorf("DBDataSource() = %q, want %q", got, want)
	}
}

func TestDBDataSource_EnvOverridesDefaults(t *testing.T) {
	clearDBEnv(t)
	t.Setenv("DB_PROTOCOL", "tcp(mysql:3306)")

	if got := DBDataSource(); !strings.Contains(got, "@tcp(mysql:3306)/") {
		t.Errorf("DBDataSource() = %q, want it to use DB_PROTOCOL", got)
	}
}

func TestDBDataSource_NoDevDefaultsInProd(t *testing.T) {
	clearDBEnv(t)
	t.Setenv("PROD", "true")

	got := DBDataSource()
	for _, dev := range []string{shared.DevDBUser, shared.DevDBPassword, shared.DevDBProtocol, shared.DevDBName} {
		if strings.Contains(got, dev) {
			t.Errorf("DBDataSource() = %q, must not contain dev default %q in prod", got, dev)
		}
	}
}
