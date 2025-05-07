package testfixtures

import "github.com/dxe/adb/model"

type ChapterBuilder struct {
	chapter model.ChapterWithToken
}

func NewChapterBuilder() *ChapterBuilder {
	return &ChapterBuilder{
		chapter: model.ChapterWithToken{
			ID:                98240,
			ChapterID:         28992,
			Name:              "Esperantoland",
			Flag:              "🟩",
			FbURL:             "https://facebook.com/dxe-esperantoland",
			TwitterURL:        "https://twitter.com/dxe-esperantoland",
			InstaURL:          "https://instagram.com/dxe-esperantoland",
			Email:             "dxe-esperantoland@example.com",
			Region:            "Default Region",
			Lat:               37.7749,
			Lng:               -122.4194,
			Distance:          10.0,
			MailingListType:   "sendgrid",
			MailingListRadius: 50,
			MailingListID:     "999999389248",
			Token:             "default_token",
			LastFBSync:        "2025-01-01",
			LastFBEvent:       "2025-01-01",
			EventbriteID:      "default_eventbrite_id",
			EventbriteToken:   "default_eventbrite_token",
			Mentor:            "Mentor Anne",
			Country:           "US",
			Notes:             "some nice notes",
			LastContact:       "2025-01-01",
			LastAction:        "2025-01-01",
			Organizers:        model.Organizers{},
			EmailToken:        "abcd00244",
		},
	}
}

func (b *ChapterBuilder) WithFbPagesTableID(id int) *ChapterBuilder {
	// Not to be confused with ChapterID.
	b.chapter.ID = id
	return b
}

func (b *ChapterBuilder) WithChapterID(chapterID int) *ChapterBuilder {
	b.chapter.ChapterID = chapterID
	return b
}

func (b *ChapterBuilder) WithName(name string) *ChapterBuilder {
	b.chapter.Name = name
	return b
}

func (b *ChapterBuilder) WithFlag(flag string) *ChapterBuilder {
	b.chapter.Flag = flag
	return b
}

func (b *ChapterBuilder) WithFbURL(fbURL string) *ChapterBuilder {
	b.chapter.FbURL = fbURL
	return b
}

func (b *ChapterBuilder) WithTwitterURL(twitterURL string) *ChapterBuilder {
	b.chapter.TwitterURL = twitterURL
	return b
}

func (b *ChapterBuilder) WithInstaURL(instaURL string) *ChapterBuilder {
	b.chapter.InstaURL = instaURL
	return b
}

func (b *ChapterBuilder) WithEmail(email string) *ChapterBuilder {
	b.chapter.Email = email
	return b
}

func (b *ChapterBuilder) WithRegion(region string) *ChapterBuilder {
	b.chapter.Region = region
	return b
}

func (b *ChapterBuilder) WithLat(lat float64) *ChapterBuilder {
	b.chapter.Lat = lat
	return b
}

func (b *ChapterBuilder) WithLng(lng float64) *ChapterBuilder {
	b.chapter.Lng = lng
	return b
}

func (b *ChapterBuilder) WithDistance(distance float32) *ChapterBuilder {
	b.chapter.Distance = distance
	return b
}

func (b *ChapterBuilder) WithMailingListType(mlType string) *ChapterBuilder {
	b.chapter.MailingListType = mlType
	return b
}

func (b *ChapterBuilder) WithMailingListRadius(mlRadius int) *ChapterBuilder {
	b.chapter.MailingListRadius = mlRadius
	return b
}

func (b *ChapterBuilder) WithMailingListID(mlID string) *ChapterBuilder {
	b.chapter.MailingListID = mlID
	return b
}

func (b *ChapterBuilder) WithToken(token string) *ChapterBuilder {
	b.chapter.Token = token
	return b
}

func (b *ChapterBuilder) WithLastFBSync(lastFBSync string) *ChapterBuilder {
	b.chapter.LastFBSync = lastFBSync
	return b
}

func (b *ChapterBuilder) WithLastFBEvent(lastFBEvent string) *ChapterBuilder {
	b.chapter.LastFBEvent = lastFBEvent
	return b
}

func (b *ChapterBuilder) WithEventbriteID(eventbriteID string) *ChapterBuilder {
	b.chapter.EventbriteID = eventbriteID
	return b
}

func (b *ChapterBuilder) WithEventbriteToken(eventbriteToken string) *ChapterBuilder {
	b.chapter.EventbriteToken = eventbriteToken
	return b
}

func (b *ChapterBuilder) WithMentor(mentor string) *ChapterBuilder {
	b.chapter.Mentor = mentor
	return b
}

func (b *ChapterBuilder) WithCountry(country string) *ChapterBuilder {
	b.chapter.Country = country
	return b
}

func (b *ChapterBuilder) WithNotes(notes string) *ChapterBuilder {
	b.chapter.Notes = notes
	return b
}

func (b *ChapterBuilder) WithLastContact(lastContact string) *ChapterBuilder {
	b.chapter.LastContact = lastContact
	return b
}

func (b *ChapterBuilder) WithLastAction(lastAction string) *ChapterBuilder {
	b.chapter.LastAction = lastAction
	return b
}

func (b *ChapterBuilder) WithOrganizers(organizers model.Organizers) *ChapterBuilder {
	b.chapter.Organizers = organizers
	return b
}

func (b *ChapterBuilder) WithEmailToken(emailToken string) *ChapterBuilder {
	b.chapter.EmailToken = emailToken
	return b
}

func (b *ChapterBuilder) Build() *model.ChapterWithToken {
	return &b.chapter
}
