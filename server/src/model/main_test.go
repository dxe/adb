package model

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	initializeTestDBSchema()
	os.Exit(m.Run())
}
