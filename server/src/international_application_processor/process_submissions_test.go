package international_application_processor

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/stretchr/testify/assert"
)

func TestPickNearestChapterOrNil(t *testing.T) {
	// Define test cases
	tests := []struct {
		name            string
		nearestChapters []model.ChapterWithToken
		country         string
		expectedChapter *model.ChapterWithToken
	}{
		{
			name: "Select nearest chapter",
			nearestChapters: []model.ChapterWithToken{
				{ChapterID: 1, Distance: 50, Country: "US"},
				{ChapterID: 2, Distance: 80, Country: "US"},
			},
			country: "US",
			expectedChapter: &model.ChapterWithToken{
				ChapterID: 1, Distance: 50, Country: "US",
			},
		},
		{
			name: "No chapter within 100 miles",
			nearestChapters: []model.ChapterWithToken{
				{ChapterID: 1, Distance: 150, Country: "US"},
				{ChapterID: 2, Distance: 200, Country: "US"},
			},
			country:         "US",
			expectedChapter: nil,
		},
		{
			name: "No matching country",
			nearestChapters: []model.ChapterWithToken{
				{ChapterID: 1, Distance: 50, Country: "CA"},
				{ChapterID: 2, Distance: 80, Country: "CA"},
			},
			country:         "US",
			expectedChapter: nil,
		},
		{
			name: "Skip different country",
			nearestChapters: []model.ChapterWithToken{
				{ChapterID: 1, Distance: 10, Country: "CA"},
				{ChapterID: 3, Distance: 20, Country: "US"},
				{ChapterID: 2, Distance: 30, Country: "US"},
			},
			country: "US",
			expectedChapter: &model.ChapterWithToken{
				ChapterID: 3, Distance: 20, Country: "US",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pickNearestChapterOrNil(tt.nearestChapters, tt.country)
			assert.Equal(t, tt.expectedChapter, result)
		})
	}
}
