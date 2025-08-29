package model

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func parseTimeOrPanic(s string) time.Time {
	time, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return time
}

type ActivistBuilder struct {
	activist ActivistExtra
}

var defaultModificationTime = parseTimeOrPanic("1970-01-01 00:00:01")

func NewActivistBuilder() *ActivistBuilder {
	return &ActivistBuilder{
		activist: ActivistExtra{
			Activist: Activist{
				ID:           0,
				Name:         "name" + fmt.Sprintf("%d", time.Now().UnixNano()),
				ChapterID:    SFBayChapterIdDevTest,
				EmailUpdated: defaultModificationTime,
				PhoneUpdated: defaultModificationTime,

				LocationUpdated: defaultModificationTime,
			},
			ActivistConnectionData: ActivistConnectionData{
				AddressUpdated: defaultModificationTime,
			},
		},
	}
}

func (b *ActivistBuilder) WithName(name string) *ActivistBuilder {
	b.activist.Name = name
	return b
}

func (b *ActivistBuilder) WithEmail(email string) *ActivistBuilder {
	b.activist.Email = email
	return b
}

func (b *ActivistBuilder) WithPhone(phone string) *ActivistBuilder {
	b.activist.Phone = phone
	return b
}

func (b *ActivistBuilder) WithChapterID(chapterID int) *ActivistBuilder {
	b.activist.ChapterID = chapterID
	return b
}

func (b *ActivistBuilder) WithAddress(street string, city string, state string) *ActivistBuilder {
	b.activist.StreetAddress = street
	b.activist.City = city
	b.activist.State = state
	return b
}

func (b *ActivistBuilder) WithLocation(location sql.NullString) *ActivistBuilder {
	b.activist.Location = location
	return b
}

func (b *ActivistBuilder) WithCoords(lat float64, lng float64) *ActivistBuilder {
	b.activist.Lat = lat
	b.activist.Lng = lng
	return b
}

func (b *ActivistBuilder) Build() *ActivistExtra {
	return &b.activist
}

func MustInsertActivist(t *testing.T, db *sqlx.DB, activist *ActivistExtra) {
	id, err := CreateActivist(db, *activist)
	if err != nil {
		t.Fatalf("MustInsertActivist failed: %v", err)
	}
	activist.ID = id
}

func MustInsertActivistWithTimestamps(t *testing.T, db *sqlx.DB, activist *ActivistExtra) {
	id, err := CreateActivistWithTimestamps(db, *activist)
	if err != nil {
		t.Fatalf("MustInsertActivistWithTimestamps failed: %v", err)
	}
	activist.ID = id
}

func MustGetActivist(t *testing.T, db *sqlx.DB, id int) *ActivistExtra {
	activist, err := GetActivistExtra(db, id)
	if err != nil {
		t.Fatalf("MustGetActivist failed: %v", err)
	}
	return activist
}
