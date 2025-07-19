package testfixtures

import "github.com/dxe/adb/model"

type ActivistBuilder struct {
	activist model.ActivistExtra
}

func NewActivistBuilder() *ActivistBuilder {
	return &ActivistBuilder{
		activist: model.ActivistExtra{
			Activist: model.Activist{
				ID:        0,
				Email:     "email1",
				Name:      "name1",
				ChapterID: model.SFBayChapterIdDevTest,
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

func (b *ActivistBuilder) Build() *model.ActivistExtra {
	return &b.activist
}
