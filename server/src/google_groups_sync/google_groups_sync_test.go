package google_groups_sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dxe/adb/config"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountKey(t *testing.T) {
	origJSON, origFile := config.SyncMailingListsConfigJSON, config.SyncMailingListsConfigFile
	t.Cleanup(func() {
		config.SyncMailingListsConfigJSON, config.SyncMailingListsConfigFile = origJSON, origFile
	})

	path := filepath.Join(t.TempDir(), "client_secrets.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"from":"file"}`), 0600))

	// The file is read when only it is set.
	config.SyncMailingListsConfigJSON, config.SyncMailingListsConfigFile = "", path
	key, err := serviceAccountKey()
	require.NoError(t, err)
	require.Equal(t, `{"from":"file"}`, string(key))

	// The environment wins when both are set, so a deployment that injects the
	// key as a variable does not also need the file to exist.
	config.SyncMailingListsConfigJSON = `{"from":"env"}`
	key, err = serviceAccountKey()
	require.NoError(t, err)
	require.Equal(t, `{"from":"env"}`, string(key))

	// A missing file is still an error rather than an empty key.
	config.SyncMailingListsConfigJSON, config.SyncMailingListsConfigFile = "", path+".nope"
	_, err = serviceAccountKey()
	require.Error(t, err)
}

func TestGetInsertAndRemoveEmails(t *testing.T) {
	// Should not add/remove activists if the list is empty.
	i0, r0 := getInsertAndRemoveEmails([]string{
		"hello@hello.com",
	}, []string{
		"hello@hello.com",
	})

	require.Equal(t, len(i0), 0)
	require.Equal(t, len(r0), 0)

	// Test that emails are added/removed correctly.
	i1, r1 := getInsertAndRemoveEmails([]string{
		"hello@hello.com",
		"goodbye@goodbye.com",
		"heyo@hey.com",
	}, []string{
		"hello@hello.com",
		"anotherone@yo.com",
	})

	require.Equal(t, stringArrayToMap(i1), map[string]struct{}{
		"heyo@hey.com":        struct{}{},
		"goodbye@goodbye.com": struct{}{},
	})
	require.Equal(t, r1, []string{"anotherone@yo.com"})
}

func stringArrayToMap(a []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, item := range a {
		m[item] = struct{}{}
	}
	return m
}
