package model

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

type ActivistBuilder struct {
	activist ActivistExtra
}

func NewActivistBuilder() *ActivistBuilder {
	return &ActivistBuilder{
		activist: ActivistExtra{
			Activist: Activist{
				ID:        0,
				Email:     "email1",
				Name:      "name1",
				ChapterID: SFBayChapterIdDevTest,
			},
		},
	}
}

func (b *ActivistBuilder) WithEmail(email string) *ActivistBuilder {
	b.activist.Email = email
	return b
}

func (b *ActivistBuilder) WithName(name string) *ActivistBuilder {
	b.activist.Name = name
	return b
}

func (b *ActivistBuilder) WithChapterID(chapterID int) *ActivistBuilder {
	b.activist.ChapterID = chapterID
	return b
}

func (b *ActivistBuilder) Build() *ActivistExtra {
	return &b.activist
}

func MustInsertActivist(t *testing.T, db *sqlx.DB, activist *ActivistExtra) {
	id, err := CreateActivist(db, *activist)
	if err != nil {
		t.Fatalf("MustInsertActivist failed: %s", err)
	}
	activist.ID = id
}
