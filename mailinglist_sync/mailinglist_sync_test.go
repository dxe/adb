package mailinglist_sync

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/stretchr/testify/require"
)

func TestGetInsertAndRemoveEmails(t *testing.T) {
	// Should not add/remove activists if the list is empty.
	i0, r0 := getInsertAndRemoveEmails([]model.WorkingGroupMember{{
		ActivistEmail: "hello@hello.com",
	}}, []string{"hello@hello.com"})

	require.Equal(t, len(i0), 0)
	require.Equal(t, len(r0), 0)

	// Test that emails are added/removed correctly.
	i1, r1 := getInsertAndRemoveEmails([]model.WorkingGroupMember{{
		ActivistEmail: "hello@hello.com",
	}, {
		ActivistEmail: "goodbye@goodbye.com",
	}, {
		ActivistEmail: "heyo@hey.com",
	}}, []string{
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
