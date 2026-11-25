package model

import (
	"testing"
	"time"

	"github.com/dxe/adb/pkg/shared"
	"github.com/stretchr/testify/require"
)

type fakeExternalEventRepo struct {
	eventChapters map[string]int
	featured      map[string]bool
	canceled      map[string]bool
	listed        bool
}

var _ ExternalEventRepository = (*fakeExternalEventRepo)(nil)

func newFakeExternalEventRepo() *fakeExternalEventRepo {
	return &fakeExternalEventRepo{
		eventChapters: map[string]int{"1": 10, "2": 20},
		featured:      map[string]bool{},
		canceled:      map[string]bool{},
	}
}

func (f *fakeExternalEventRepo) ListUpcomingForChapter(int, time.Time) ([]ExternalEvent, error) {
	f.listed = true
	return nil, nil
}
func (f *fakeExternalEventRepo) GetChapterID(id string) (int, error) {
	c, ok := f.eventChapters[id]
	if !ok {
		return 0, ErrNotFound
	}
	return c, nil
}
func (f *fakeExternalEventRepo) SetFeatured(_ int, id string, v bool) error {
	f.featured[id] = v
	return nil
}
func (f *fakeExternalEventRepo) Cancel(_ int, id string) error {
	f.canceled[id] = true
	return nil
}

func TestExternalEventAuthorization(t *testing.T) {
	organizer := ADBUser{Roles: []string{shared.RoleOrganizer}, ChapterID: 10}
	admin := ADBUser{Roles: []string{shared.RoleAdmin}, ChapterID: 10}
	noChapter := ADBUser{Roles: []string{shared.RoleOrganizer}}
	noRole := ADBUser{Roles: []string{shared.RoleAttendance}, ChapterID: 10}

	cases := []struct {
		name      string
		user      ADBUser
		chapterID int
		eventID   string
		wantErr   error // nil = allowed
	}{
		{"organizer own chapter", organizer, 10, "1", nil},
		{"organizer other chapter", organizer, 20, "2", ErrValidation},
		{"admin other chapter", admin, 20, "2", nil},
		{"organizer without chapter", noChapter, 10, "1", ErrValidation},
		{"non-organizer role", noRole, 10, "1", ErrValidation},
		{"unknown event", organizer, 10, "999", ErrNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeExternalEventRepo()
			allowed := tc.wantErr == nil

			// Listing doesn't look at event IDs, so "unknown event" is allowed here.
			_, err := ListExternalEventsForChapter(repo, tc.user, tc.chapterID, time.Now())
			if allowed || tc.wantErr == ErrNotFound {
				require.NoError(t, err)
				require.True(t, repo.listed)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
				require.False(t, repo.listed, "repo must not be queried when access is denied")
			}

			err = FeatureExternalEvent(repo, tc.user, tc.eventID, true)
			if allowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
			require.Equal(t, allowed, repo.featured[tc.eventID])

			err = CancelExternalEvent(repo, tc.user, tc.eventID)
			if allowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
			require.Equal(t, allowed, repo.canceled[tc.eventID])
		})
	}
}
