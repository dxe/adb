package international_application_processor

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testfixtures"
	"github.com/stretchr/testify/require"
)

func TestBuildNotificationEmail(t *testing.T) {
	t.Run("IncludesResponderInfo", func(t *testing.T) {
		// Arrange
		formData := testfixtures.NewInternationalFormDataBuilder().
			WithEmail("test@example.com").
			WithFirstName("John").
			WithLastName("Doe").
			WithPhone("123-456-7890").
			WithCity("Test City").
			Build()

		chapter := testfixtures.NewChapterBuilder().Build()

		// Act
		msg, err := buildNotificationEmail(formData, chapter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, "John Doe signed up to join your chapter", msg.Subject)
		require.Contains(t, msg.BodyHTML, "Name: John Doe")
		require.Contains(t, msg.BodyHTML, "Email: test@example.com")
		require.Contains(t, msg.BodyHTML, "Phone: 123-456-7890")
		require.Contains(t, msg.BodyHTML, "City: Test City")
	})

	t.Run("EmailsChapterEmail", func(t *testing.T) {
		// Arrange
		formData := testfixtures.NewInternationalFormDataBuilder().Build()

		chapter := testfixtures.NewChapterBuilder().
			WithEmail("chapter@example.com").
			Build()

		// Act
		msg, err := buildNotificationEmail(formData, chapter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, "chapter@example.com", msg.ToAddress)
	})

	t.Run("EmailsChapterOrganizers", func(t *testing.T) {
		// Arrange
		formData := testfixtures.NewInternationalFormDataBuilder().Build()

		chapter := testfixtures.NewChapterBuilder().
			WithEmail("").
			WithOrganizers([]*model.Organizer{
				{Name: "Organizer 1", Email: "organizer1@example.com"},
				{Name: "Organizer 2", Email: "organizer2@example.com"},
			}).
			Build()

		// Act
		msg, err := buildNotificationEmail(formData, chapter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, "organizer1@example.com", msg.ToAddress)
		require.Equal(t, 1, len(msg.CC))
		require.Equal(t, "organizer2@example.com", msg.CC[0])
	})

	t.Run("EmailsChapterAndOrganizers", func(t *testing.T) {
		// Arrange
		formData := testfixtures.NewInternationalFormDataBuilder().Build()

		chapter := testfixtures.NewChapterBuilder().
			WithEmail("chapter@example.com").
			WithOrganizers([]*model.Organizer{
				{Name: "Organizer 1", Email: "organizer1@example.com"},
				{Name: "Organizer 2", Email: "organizer2@example.com"},
			}).
			Build()

		// Act
		msg, err := buildNotificationEmail(formData, chapter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, "chapter@example.com", msg.ToAddress)
		require.Equal(t, 2, len(msg.CC))
		require.Equal(t, "organizer1@example.com", msg.CC[0])
		require.Equal(t, "organizer2@example.com", msg.CC[1])
	})

	t.Run("EmailsSFBayCoordinator", func(t *testing.T) {
		// Arrange
		formData := testfixtures.NewInternationalFormDataBuilder().Build()

		chapter := testfixtures.NewChapterBuilder().
			WithChapterID(model.SFBayChapterId).
			Build()

		// Act
		msg, err := buildNotificationEmail(formData, chapter)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, sfBayCoordinator.Address, msg.ToAddress)
	})

	t.Run("EmailsCaliforniaCoordinator", func(t *testing.T) {
		formData := testfixtures.NewInternationalFormDataBuilder().
			WithState("CA").
			Build()

		msg, err := buildNotificationEmail(formData, nil)

		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, californiaCoordinator.Address, msg.ToAddress)
	})

	t.Run("EmailsGlobalCoordinator", func(t *testing.T) {
		formData := testfixtures.NewInternationalFormDataBuilder().
			WithState("ZZ").
			Build()

		msg, err := buildNotificationEmail(formData, nil)

		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, globalCoordinator.Address, msg.ToAddress)
	})
}
