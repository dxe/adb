package international_application_processor

import (
	"testing"

	"time"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testfixtures"
	"github.com/stretchr/testify/assert"
)

func TestBuildOnboardingEmailMessage(t *testing.T) {
	t.Run("ForSFBayChapter", func(t *testing.T) {
		t.Run("ContainsBasicInfo", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithFirstName("John").
				WithLastName("Doe").
				WithEmail("john.doe@example.com").
				Build()

			chapter := testfixtures.NewChapterBuilder().
				WithChapterID(model.SFBayChapterId).
				WithName("SF Bay").
				WithFbURL("https://facebook.com/test-chapter").
				WithInstaURL("https://instagram.com/test-chapter").
				WithTwitterURL("https://twitter.com/test-chapter").
				WithEmail("chapter-email@example.org").
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Equal(t, sfBayCoordinator.Name, msg.FromName)
			assert.Equal(t, sfBayCoordinator.Address, msg.FromAddress)
			assert.Contains(t, msg.BCC, sfBayCoordinator.Address)
			assert.Equal(t, "John Doe", msg.ToName)
			assert.Equal(t, "john.doe@example.com", msg.ToAddress)
			assert.Equal(t, "Join your local Direct Action Everywhere chapter!", msg.Subject)
			assert.Contains(t, msg.BodyHTML, "Hey John!")
			assert.Contains(t, msg.BodyHTML, "The DxE SF Bay Area chapter is within 100 miles of you.")
			assert.Contains(t, msg.BodyHTML, "https://dxe.io/events")
		})

		t.Run("ContainsNextEventLink", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().Build()
			chapter := testfixtures.NewChapterBuilder().Build()
			var nextEvent *model.ExternalEvent = &model.ExternalEvent{
				ID:        "555500",
				Name:      "Test Event",
				StartTime: time.Now().Add(48 * time.Hour),
			}

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nextEvent)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Contains(t, msg.BodyHTML, "Test Event")
			assert.Contains(t, msg.BodyHTML, "https://facebook.com/events/555500")
		})
	})

	t.Run("ForNonSFBayChapter", func(t *testing.T) {
		t.Run("ContainsBasicInfo", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithFirstName("John").
				WithLastName("Doe").
				WithEmail("john.doe@example.com").
				Build()

			chapter := testfixtures.NewChapterBuilder().
				WithName("Test Chapter").
				WithFbURL("https://facebook.com/test-chapter").
				WithInstaURL("https://instagram.com/test-chapter").
				WithTwitterURL("https://twitter.com/test-chapter").
				WithEmail("chapter-email@example.org").
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Equal(t, globalCoordinator.Name, msg.FromName)
			assert.Equal(t, globalCoordinator.Address, msg.FromAddress)
			assert.Contains(t, msg.BCC, globalCoordinator.Address)
			assert.Equal(t, "John Doe", msg.ToName)
			assert.Equal(t, "john.doe@example.com", msg.ToAddress)
			assert.Equal(t, "Join your local Direct Action Everywhere chapter!", msg.Subject)
			assert.Contains(t, msg.BodyHTML, "Hey John!")
			assert.Contains(t, msg.BodyHTML, "There is currently a DxE chapter near you")
			assert.Contains(t, msg.BodyHTML, "Test Chapter")
			assert.Contains(t, msg.BodyHTML, "https://facebook.com/test-chapter")
			assert.Contains(t, msg.BodyHTML, "https://instagram.com/test-chapter")
			assert.Contains(t, msg.BodyHTML, "https://twitter.com/test-chapter")
			assert.Contains(t, msg.BodyHTML, "mailto:chapter-email@example.org")
		})

		t.Run("ContainsNextEventLink", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().Build()
			chapter := testfixtures.NewChapterBuilder().Build()
			var nextEvent *model.ExternalEvent = &model.ExternalEvent{
				ID:        "555500",
				Name:      "Test Event",
				StartTime: time.Now().Add(48 * time.Hour),
			}

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nextEvent)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Contains(t, msg.BodyHTML, "Test Event")
			assert.Contains(t, msg.BodyHTML, "https://facebook.com/events/555500")
		})

		t.Run("CCsChapterEmailAndOrganizers", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().Build()

			chapter := testfixtures.NewChapterBuilder().
				WithEmail("test-email@example.org").
				WithOrganizers([]*model.Organizer{
					{Name: "Test Organizer", Email: "organizer1@example.org"},
					{Name: "Test Organizer 2", Email: "organizer2@example.org"},
				}).
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Contains(t, msg.CC, "test-email@example.org")
			assert.Contains(t, msg.CC, "organizer1@example.org")
			assert.Contains(t, msg.CC, "organizer2@example.org")
			assert.Equal(t, 3, len(msg.CC))
		})

		t.Run("CCsGlobalCoordinator", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithState("ZZ").
				Build()

			chapter := testfixtures.NewChapterBuilder().
				WithEmail("").
				WithOrganizers([]*model.Organizer{}).
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Contains(t, msg.CC, globalCoordinator.Address)
		})

		t.Run("CCsCACoordinator", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithState("CA").
				Build()

			chapter := testfixtures.NewChapterBuilder().
				WithEmail("").
				WithOrganizers([]*model.Organizer{}).
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, &chapter, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Contains(t, msg.CC, californiaCoordinator.Address)
		})
	})

	t.Run("ForCaOrganizerNotNearAnyChapter", func(t *testing.T) {
		t.Run("ContainsBasicInfo", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithFirstName("John").
				WithLastName("Doe").
				WithEmail("john.doe@example.com").
				WithInvolvement("organize").
				WithState("CA").
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, nil, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Equal(t, californiaCoordinator.Name, msg.FromName)
			assert.Equal(t, californiaCoordinator.Address, msg.FromAddress)
			assert.Contains(t, msg.BCC, californiaCoordinator.Address)
			assert.Equal(t, "John Doe", msg.ToName)
			assert.Equal(t, "john.doe@example.com", msg.ToAddress)
			assert.Equal(t, "Getting involved with Direct Action Everywhere", msg.Subject)
			assert.Contains(t, msg.BodyHTML, "Hi John,")
			assert.Contains(t, msg.BodyHTML, "There is no active chapter in your area")
		})
	})

	t.Run("ForCaParticipantNotNearAnyChapter", func(t *testing.T) {

	})

	t.Run("ForNonCaOrganizerNotNearAnyChapter", func(t *testing.T) {
		t.Run("ContainsBasicInfo", func(t *testing.T) {
			// Arrange
			formData := testfixtures.NewInternationalFormDataBuilder().
				WithFirstName("John").
				WithLastName("Doe").
				WithEmail("john.doe@example.com").
				WithInvolvement("organize").
				WithState("ZZ").
				Build()

			// Act
			msg, err := buildOnboardingEmailMessage(formData, nil, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, msg)
			assert.Equal(t, globalCoordinator.Name, msg.FromName)
			assert.Equal(t, globalCoordinator.Address, msg.FromAddress)
			assert.Contains(t, msg.BCC, globalCoordinator.Address)
			assert.Equal(t, "John Doe", msg.ToName)
			assert.Equal(t, "john.doe@example.com", msg.ToAddress)
			assert.Equal(t, "Getting involved with Direct Action Everywhere", msg.Subject)
			assert.Contains(t, msg.BodyHTML, "Hi John,")
			assert.Contains(t, msg.BodyHTML, "I am part of the international coordination (IC) team")
		})
	})

	t.Run("ForNonCaParticipantNotNearAnyChapter", func(t *testing.T) {

	})
}
