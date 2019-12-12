package survey_mailer

import (
	"log"
	"strings"
	"time"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/sourcegraph/go-ses"
)

func sendMissingEmail(eventName string, attendees []string) {
	if len(attendees) > 0 {
		to := config.SurveyMissingEmail
		subject := "Missing emails for survey: " + eventName
		bodyText := "The following people did not receive a survey for this protest due to not having a valid email address: "
		bodyText += strings.Join(attendees, ", ")
		bodyHtml := bodyText
		sendEmail(to, subject, bodyText, bodyHtml)
	}
}

func sendEmail(to string, subject string, bodyText string, bodyHtml string) {
	from := config.SurveyFromEmail
	bodyHtml += `<br /><img src="https://adb.dxe.io/static/img/logo1.png" height="46" width="50">`
	// EnvConfig uses the AWS credentials in the environment
	// variables $AWS_ACCESS_KEY_ID nd $AWS_SECRET_KEY.
	_, err := ses.EnvConfig.SendEmailHTML(from, to, subject, bodyText, bodyHtml)
	if err != nil {
		log.Printf("Error sending email: %s\n", err)
	}
}

func bulkSendEmails(event model.Event, subject string, bodyText string, bodyHtml string) {
	for _, recipient := range event.AttendeeEmails {
		log.Println("Sending email to:", recipient)
		// Send email
		sendEmail(recipient, subject, bodyText, bodyHtml)
	}
	sendMissingEmail(event.EventName, event.AttendeeMissingEmails)
}

func updateSurveyStatus(db *sqlx.DB, eventId int) {
	// Update "survey_sent" to true (1)
	_, err := model.UpdateEventSurveyStatus(db, model.Event{
		ID:         eventId,
		SurveySent: 1,
	})
	if err != nil {
		log.Println("ERROR:", err)
	}
}

func surveyMeetup(db *sqlx.DB, queryDate string) {

	log.Println("Looking for meetup events on", queryDate)

	// Get today's meetup events that haven't had surveys sent yet
	events, err := model.GetEvents(db, model.GetEventOptions{
		DateFrom:       queryDate,
		DateTo:         queryDate,
		EventType:      "Community",
		EventNameQuery: "%meetup%",
		SurveySent:     "0",
	})
	if err != nil {
		log.Printf("Failed to get today's meetup events: %v", err)
		return
	}

	// Iterate through events
	for _, event := range events {

		// Set email subject & body
		subject := "Survey: " + event.EventName
		eventDateParam := event.EventDate.Format("2006-01-02")
		bodyText := `Thank you for attending the meetup! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSfV0smO8sQo1ch-rlX7g9Oz4t_2d3fjGytwrE_yJ8Ez9uLSZQ/viewform?usp=pp_url&entry.1369832182=` + eventDateParam
		bodyHtml := `<p>Thank you for attending the meetup! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfV0smO8sQo1ch-rlX7g9Oz4t_2d3fjGytwrE_yJ8Ez9uLSZQ/viewform?usp=pp_url&entry.1369832182=` + eventDateParam + `">click here</a> to provide feedback which will help us in planning future events.</p>`

		log.Println("Sending meetup survey for event:", event.EventName)

		// send all emails, including "missing" email
		bulkSendEmails(event, subject, bodyText, bodyHtml)

		// updateSurveyStatus
		updateSurveyStatus(db, event.ID)
	}
}

func surveyChapterMeeting(db *sqlx.DB, queryDate string) {

	log.Println("Looking for chapter meeting events on", queryDate)

	// Get today's chapter meeting events that haven't had surveys sent yet
	events, err := model.GetEvents(db, model.GetEventOptions{
		DateFrom:       queryDate,
		DateTo:         queryDate,
		EventNameQuery: "%chapter meeting%",
		SurveySent:     "0",
	})
	if err != nil {
		log.Printf("Failed to get today's chapter meeting events: %v", err)
		return
	}

	// Iterate through events
	for _, event := range events {

		// Set email subject & body
		subject := "Survey: " + event.EventName
		eventDateParam := event.EventDate.Format("2006-01-02")
		bodyText := `Thank you for attending the chapter meeting! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSfc_mgwH_zYYEQ5MTJwgyvCy5klsY_xrVBXgTDHM8sSxLIJrQ/viewform?usp=pp_url&entry.502269384=` + eventDateParam
		bodyHtml := `<p>Thank you for attending the chapter meeting! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfc_mgwH_zYYEQ5MTJwgyvCy5klsY_xrVBXgTDHM8sSxLIJrQ/viewform?usp=pp_url&entry.502269384=` + eventDateParam + `">click here</a> to take a quick survey.</p>`

		log.Println("Sending chapter meeting survey for event:", event.EventName)

		// send all emails, including "missing" email
		bulkSendEmails(event, subject, bodyText, bodyHtml)

		// updateSurveyStatus
		updateSurveyStatus(db, event.ID)
	}
}

func surveyProtest(db *sqlx.DB, queryDate string) {

	log.Println("Looking for protest events on", queryDate)

	// Get yesterday's protest events that haven't had surveys sent yet
	events, err := model.GetEvents(db, model.GetEventOptions{
		DateFrom:   queryDate,
		DateTo:     queryDate,
		EventType:  "%action",
		SurveySent: "0",
	})

	if err != nil {
		log.Printf("Failed to get today's protest events: %v", err)
		return
	}

	// Iterate through events
	for _, event := range events {

		// Set email subject & body
		subject := "Survey: " + event.EventName
		eventNameParam := strings.Replace(event.EventName, " ", "+", -1)
		bodyText := `Thank you for taking part in direct action! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLScfrPtPxmYAroODhBkwUGq753JPykYKNdosg4gUR_SRng8BRQ/viewform?usp=pp_url&entry.466557185=` + eventNameParam + ". If you captured any photos or videos, please upload them here: dxe.io/upload."
		bodyHtml := `<p>Thank you for taking part in direct action! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLScfrPtPxmYAroODhBkwUGq753JPykYKNdosg4gUR_SRng8BRQ/viewform?usp=pp_url&entry.466557185=` + eventNameParam + `">click here</a> to take a quick survey.</p><p>If you captured any photos or videos, please upload them <a href="http://dxe.io/upload">here</a>.</p>`

		log.Println("Sending protest survey for event:", event.EventName)

		// send all emails, including "missing" email
		bulkSendEmails(event, subject, bodyText, bodyHtml)

		// updateSurveyStatus
		updateSurveyStatus(db, event.ID)
	}
}

func surveyMailerWrapper(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in survey mailer", r)
		}
	}()

	// Get current hour of day & current day of week
	weekday := time.Now().Weekday()
	hour := time.Now().Hour()
	// Calculate date of yesterday & today
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// send protest surveys from previous day daily at 9am
	if hour == 9 {
		surveyProtest(db, yesterday)
	}

	// send meetup surveys on saturdays at 2pm
	if weekday == 6 && hour == 14 {
		surveyMeetup(db, today)
	}

	// send chapter mtg surveys on sun at 4pm
	if weekday == 0 && hour == 16 {
		surveyChapterMeeting(db, today)
	}

}

// Sends surveys based on event attendance every 60 minutes.
// Should be run in a goroutine.
func StartSurveyMailer(db *sqlx.DB) {
	for {
		log.Println("Starting survey mailer")
		surveyMailerWrapper(db)
		log.Println("Finished survey mailer")
		// TODO(jake): change this to 60 min before deploying
		time.Sleep(60 * time.Minute)
	}
}
