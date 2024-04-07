package survey_mailer

import (
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/dxe/adb/mailer"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

type SurveyOptions struct {
	SurveyType     string
	QueryDate      string
	QueryEventType string
	QueryEventName string
	BodyHtml       string
}

func sendMissingEmail(eventName string, attendees []string, sendingErrors []string) {
	subject := "Missing emails and errors for survey: " + eventName
	bodyHtml := ""

	if len(attendees) > 0 {
		bodyHtml += "<p><strong>The following people did not receive a survey for this event due to not having a valid email address:</strong><br />"
		bodyHtml += strings.Join(attendees, "<br />")
		bodyHtml += "</p>"
	}
	if len(sendingErrors) > 0 {
		bodyHtml += "<p><strong>The following addresses did not receive the email due to sending errors:</strong><br />"
		bodyHtml += strings.Join(sendingErrors, "<br />")
		bodyHtml += "</p>"
	}

	if bodyHtml == "" {
		return
	}

	err := mailer.Send(mailer.Message{
		FromName:    "DxE Surveys",
		FromAddress: config.SurveyFromEmail,
		ToAddress:   config.SurveyMissingEmail,
		Subject:     subject,
		BodyHTML:    bodyHtml,
	})
	if err != nil {
		log.Println("ERROR sending email of missing emails and errors.")
		return
	}
	log.Println("Sent email of missing emails and errors.")
}

func bulkSendEmails(event model.Event, subject string, bodyHtml string) {
	var missingEmails []string
	var sendingErrors []string
	for i, recipient := range event.Attendees {
		recipientEmail := event.AttendeeEmails[i]
		if recipientEmail == "" {
			missingEmails = append(missingEmails, recipient)
			continue
		}
		recipientName := event.Attendees[i]

		// add stanford survey link to email (DISABLED 2020.10.23 as per Eva's request)
		// newBodyHtml := bodyHtml
		// recipientID := strconv.Itoa(event.AttendeeIDs[i])
		// stanfordLink := "http://ec2.dxe.io/adb-forms/survey.php?activist-id=" + recipientID
		// newBodyHtml += "<p>P.S. You can greatly help us improve our work by <a href=\"" + stanfordLink + "\">clicking here</a> to take one additional survey. This link is unique to you, so please DO NOT share it with others.</p>"

		// Send email
		log.Println("Sending email to:", recipient)
		err := mailer.Send(mailer.Message{
			FromName:    "DxE Surveys",
			FromAddress: config.SurveyFromEmail,
			ToName:      recipientName,
			ToAddress:   recipientEmail,
			Subject:     subject,
			BodyHTML:    bodyHtml,
		})
		if err != nil {
			sendingErrors = append(sendingErrors, fmt.Sprintf("%v [%v] %v", recipient, event.AttendeeEmails[i], err))
		}
	}
	sendMissingEmail(event.EventName, missingEmails, sendingErrors)
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

func survey(db *sqlx.DB, surveyOptions SurveyOptions) {
	log.Println("Looking for", surveyOptions.SurveyType, "events on", surveyOptions.QueryDate)

	// Get events matching query that that haven't had surveys sent yet
	events, err := model.GetEvents(db, model.GetEventOptions{
		DateFrom:       surveyOptions.QueryDate,
		DateTo:         surveyOptions.QueryDate,
		EventType:      surveyOptions.QueryEventType,
		EventNameQuery: surveyOptions.QueryEventName,
		SurveySent:     "0",
		SuppressSurvey: "0",
		// TODO: consider not hard-coding this
		ChapterID: 47, // SF Bay Area
	})
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		return
	}
	log.Printf("Survey mailer found %d events\n", len(events))

	// Iterate through events
	for _, event := range events {
		subject := "Survey: " + event.EventName
		linkParamName := strings.Replace(event.EventName, " ", "+", -1)
		linkParamDate := event.EventDate.Format("2006-01-02")
		// TODO: Look into better ways for escaping this to prevent XSS attacks
		bodyHtml := strings.Replace(surveyOptions.BodyHtml, "LINK_PARAM_NAME", html.EscapeString(linkParamName), -1)
		bodyHtml = strings.Replace(bodyHtml, "LINK_PARAM_DATE", html.EscapeString(linkParamDate), -1)

		log.Println("Sending", surveyOptions.SurveyType, "survey for event:", event.EventName)

		// send all emails, including "missing" email
		bulkSendEmails(event, subject, bodyHtml)

		// update survey sent status to 1 (true)
		updateSurveyStatus(db, event.ID)
	}
}

func surveyMailerWrapper(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in survey mailer", r)
		}
	}()

	// Get current time in US Pacific time zone
	loc, _ := time.LoadLocation("America/Los_Angeles") // TODO: move to config
	now := time.Now().In(loc)
	// Calculate current hour of day & current day of week
	weekday := now.Weekday()
	hour := now.Hour()
	// Calculate date of yesterday
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// don't send surveys before 8am or after 5pm since ppl may
	// be less likely to see the email notification
	if hour < 8 || hour > 17 {
		return
	}

	// send meetup survey on sunday
	if weekday == 0 {
		survey(db, SurveyOptions{
			SurveyType:     "meetup",
			QueryDate:      yesterday,
			QueryEventType: "Community",
			QueryEventName: "Meetup",
			BodyHtml:       `<p>Thank you for attending the meetup! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfJNn8vjrLWXqFkgzKsnObUZSrbIs_9k3_KFTy7M8glt4ZNug/viewform?usp=pp_url&entry.783443419=LINK_PARAM_NAME&entry.1369832182=LINK_PARAM_DATE">click here</a> to provide feedback which will help us in planning future events.</p>`,
		})
	}

	// send chapter mtg surveys on monday
	if weekday == 1 {
		survey(db, SurveyOptions{
			SurveyType:     "chapter meeting",
			QueryDate:      yesterday,
			QueryEventType: "",
			QueryEventName: `"Chapter Meeting"`,
			BodyHtml:       `<p>Thank you for attending the chapter meeting! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfc_mgwH_zYYEQ5MTJwgyvCy5klsY_xrVBXgTDHM8sSxLIJrQ/viewform?usp=pp_url&entry.502269384=LINK_PARAM_DATE">click here</a> to take a quick survey.</p>`,
		})
	}

	// send protest, animal care, & community surveys daily
	survey(db, SurveyOptions{
		SurveyType:     "protest",
		QueryDate:      yesterday,
		QueryEventType: "%Action",
		QueryEventName: "",
		BodyHtml:       `<p>Thank you for taking part in direct action! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLScfrPtPxmYAroODhBkwUGq753JPykYKNdosg4gUR_SRng8BRQ/viewform?usp=pp_url&entry.466557185=LINK_PARAM_NAME">click here</a> to take a quick survey.</p><p>If you captured any photos or videos, please upload them <a href="http://dxe.io/upload">here</a>.</p>`,
	})
	survey(db, SurveyOptions{
		SurveyType:     "animal care",
		QueryDate:      yesterday,
		QueryEventType: "Animal Care",
		QueryEventName: "",
		BodyHtml:       `<p>Thank you for attending an animal care event! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSdxn514dpwXduMeaGr8xCszoAUYDS0_95faskbFCzVNcAJ_fw/viewform?usp=pp_url&entry.466557185=LINK_PARAM_NAME">click here</a> to take a quick survey.</p><p>If you captured any photos or videos, please upload them <a href="http://dxe.io/upload">here</a>.</p>`,
	})
	survey(db, SurveyOptions{
		SurveyType:     "community",
		QueryDate:      yesterday,
		QueryEventType: "Community",
		QueryEventName: "",
		BodyHtml:       `<p>Thank you for attending our community event! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfJNn8vjrLWXqFkgzKsnObUZSrbIs_9k3_KFTy7M8glt4ZNug/viewform?usp=pp_url&entry.783443419=LINK_PARAM_NAME&entry.1369832182=LINK_PARAM_DATE">click here</a> to provide feedback which will help us in planning future events.</p>`,
	})

}

func validSurveyConfig() bool {
	if config.SurveyMissingEmail == "" {
		return false
	}
	if config.SurveyFromEmail == "" {
		return false
	}
	return true
}

// Sends surveys based on event attendance every 60 minutes.
// Should be run in a goroutine.
func StartSurveyMailer(db *sqlx.DB) {

	if !validSurveyConfig() {
		log.Println("WARNING: Survey config invalid.")
		return
	}

	for {
		log.Println("Starting survey mailer")
		surveyMailerWrapper(db)
		log.Println("Finished survey mailer")
		time.Sleep(60 * time.Minute)
	}
}
