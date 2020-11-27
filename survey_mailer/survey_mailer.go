package survey_mailer

import (
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/sourcegraph/go-ses"
)

type SurveyOptions struct {
	SurveyType     string
	QueryDate      string
	QueryEventType string
	QueryEventName string
	BodyText       string
	BodyHtml       string
	LinkParam      string
}

func sendMissingEmail(eventName string, attendees []string, sendingErrors []string) {
	subject := "Missing emails and errors for survey: " + eventName
	to := config.SurveyMissingEmail
	bodyText := ""
	bodyHtml := ""

	if len(attendees) > 0 {
		bodyText += "The following people did not receive a survey for this event due to not having a valid email address: "
		bodyText += strings.Join(attendees, ", ")
		bodyText += ". "
		bodyHtml += "<p><strong>The following people did not receive a survey for this event due to not having a valid email address:</strong><br />"
		bodyHtml += strings.Join(attendees, "<br />")
		bodyHtml += "</p>"
	}
	if len(sendingErrors) > 0 {
		bodyText += "The following addresses did not receive the email due to sending errors: "
		bodyText += strings.Join(sendingErrors, ", ")
		bodyText += ". "
		bodyHtml += "<p><strong>The following addresses did not receive the email due to sending errors:</strong><br />"
		bodyHtml += strings.Join(sendingErrors, "<br />")
		bodyHtml += "</p>"
	}

	if bodyText != "" {
		sendEmail(to, subject, bodyText, bodyHtml)
		log.Println("Sending email of missing emails and errors.")
	}
}

func sendEmail(to string, subject string, bodyText string, bodyHtml string) error {
	from := config.SurveyFromEmail
	bodyHtml += `<br /><img src="https://adb.dxe.io/static/img/logo1.png" height="46" width="50">`
	// EnvConfig uses the AWS credentials in the environment
	// variables $AWS_ACCESS_KEY_ID and $AWS_SECRET_KEY.
	_, err := ses.EnvConfig.SendEmailHTML(from, to, subject, bodyText, bodyHtml)
	if err != nil {
		return err
	}
	return nil
}

func bulkSendEmails(event model.Event, subject string, bodyText string, bodyHtml string) {
	var missingEmails []string
	var sendingErrors []string
	for i, recipient := range event.Attendees {
		receipientEmail := event.AttendeeEmails[i]
		if receipientEmail == "" {
			missingEmails = append(missingEmails, recipient)
			continue
		}

		// add stanford survey link to email (DISABLED 2020.10.23 as per Eva's request)
		// newBodyText := bodyText
		// newBodyHtml := bodyHtml
		// receipientID := strconv.Itoa(event.AttendeeIDs[i])
		// stanfordLink := "http://ec2.dxe.io/adb-forms/survey.php?activist-id=" + receipientID
		// newBodyText += "P.S. You can greatly help us improve our work by clicking the following link to take one additional survey. This link is unique to you, so please DO NOT share it with others: " + stanfordLink
		// newBodyHtml += "<p>P.S. You can greatly help us improve our work by <a href=\"" + stanfordLink + "\">clicking here</a> to take one additional survey. This link is unique to you, so please DO NOT share it with others.</p>"

		// Send email
		log.Println("Sending email to:", recipient)
		err := sendEmail(receipientEmail, subject, bodyText, bodyHtml)
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
	})
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		return
	}

	// Iterate through events
	for _, event := range events {
		subject := "Survey: " + event.EventName
		// set linkParam based on LinkParam option
		linkParam := ""
		if surveyOptions.LinkParam == "name" {
			linkParam = strings.Replace(event.EventName, " ", "+", -1)
		}
		if surveyOptions.LinkParam == "date" {
			linkParam = event.EventDate.Format("2006-01-02")
		}
		// build body by replacing LINK_PARAM with the actual link param
		bodyText := strings.Replace(surveyOptions.BodyText, "LINK_PARAM", linkParam, -1)
		// TODO: Look into better ways for escaping this to prevent XSS attacks
		bodyHtml := strings.Replace(surveyOptions.BodyHtml, "LINK_PARAM", html.EscapeString(linkParam), -1)

		log.Println("Sending", surveyOptions.SurveyType, "survey for event:", event.EventName)

		// send all emails, including "missing" email
		bulkSendEmails(event, subject, bodyText, bodyHtml)

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
	loc, _ := time.LoadLocation("America/Los_Angeles")
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

	// send protest, sanctuary, & training surveys daily
	survey(db, SurveyOptions{
		SurveyType:     "protest",
		QueryDate:      yesterday,
		QueryEventType: "%Action",
		QueryEventName: "",
		BodyText:       `Thank you for taking part in direct action! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLScfrPtPxmYAroODhBkwUGq753JPykYKNdosg4gUR_SRng8BRQ/viewform?usp=pp_url&entry.466557185=LINK_PARAM. If you captured any photos or videos, please upload them here: dxe.io/upload.`,
		BodyHtml:       `<p>Thank you for taking part in direct action! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLScfrPtPxmYAroODhBkwUGq753JPykYKNdosg4gUR_SRng8BRQ/viewform?usp=pp_url&entry.466557185=LINK_PARAM">click here</a> to take a quick survey.</p><p>If you captured any photos or videos, please upload them <a href="http://dxe.io/upload">here</a>.</p>`,
		LinkParam:      "name",
	})
	survey(db, SurveyOptions{
		SurveyType:     "sanctuary",
		QueryDate:      yesterday,
		QueryEventType: "Sanctuary",
		QueryEventName: "",
		BodyText:       `Thank you for attending a sanctuary event! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSdxn514dpwXduMeaGr8xCszoAUYDS0_95faskbFCzVNcAJ_fw/viewform?usp=pp_url&entry.466557185=LINK_PARAM. If you captured any photos or videos, please upload them here: dxe.io/upload.`,
		BodyHtml:       `<p>Thank you for attending a sanctuary event! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSdxn514dpwXduMeaGr8xCszoAUYDS0_95faskbFCzVNcAJ_fw/viewform?usp=pp_url&entry.466557185=LINK_PARAM">click here</a> to take a quick survey.</p><p>If you captured any photos or videos, please upload them <a href="http://dxe.io/upload">here</a>.</p>`,
		LinkParam:      "name",
	})
	// trainings survey disabled as per ana's request on 2020.10.22
	// survey(db, SurveyOptions{
	// 	SurveyType:     "training",
	// 	QueryDate:      yesterday,
	// 	QueryEventType: "Training",
	// 	QueryEventName: "",
	// 	BodyText:       `Thank you for attending the training yesterday! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSdIUm7nBy83mTIgZ8eDcSjCHZTZ62jcTghmx_q25eA3umo2ug/viewform?usp=pp_url&entry.1179483307=LINK_PARAM.`,
	// 	BodyHtml:       `<p>Thank you for attending the training yesterday! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSdIUm7nBy83mTIgZ8eDcSjCHZTZ62jcTghmx_q25eA3umo2ug/viewform?usp=pp_url&entry.1179483307=LINK_PARAM">click here</a> to take a quick survey.</p>`,
	// 	LinkParam:      "name",
	// })

	// only send meetup & popup surveys on sunday
	if weekday == 0 {
		survey(db, SurveyOptions{
			SurveyType:     "meetup",
			QueryDate:      yesterday,
			QueryEventType: "Community",
			QueryEventName: "Meetup",
			BodyText:       `Thank you for attending the meetup! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSfV0smO8sQo1ch-rlX7g9Oz4t_2d3fjGytwrE_yJ8Ez9uLSZQ/viewform?usp=pp_url&entry.1369832182=LINK_PARAM`,
			BodyHtml:       `<p>Thank you for attending the meetup! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfV0smO8sQo1ch-rlX7g9Oz4t_2d3fjGytwrE_yJ8Ez9uLSZQ/viewform?usp=pp_url&entry.1369832182=LINK_PARAM">click here</a> to provide feedback which will help us in planning future events.</p>`,
			LinkParam:      "date",
		})
		survey(db, SurveyOptions{
			SurveyType:     "popup",
			QueryDate:      yesterday,
			QueryEventType: "Community",
			QueryEventName: "Popup",
			BodyText:       `Thank you for attending the popup! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLScwpVIvHItvJeUPkKk_UsRjsrDxj29vK8zElS19nnEZmaEy9Q/viewform?usp=pp_url&entry.610934849=LINK_PARAM`,
			BodyHtml:       `<p>Thank you for attending the meetup! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLScwpVIvHItvJeUPkKk_UsRjsrDxj29vK8zElS19nnEZmaEy9Q/viewform?usp=pp_url&entry.610934849=LINK_PARAM">click here</a> to provide feedback which will help us in planning future events.</p>`,
			LinkParam:      "date",
		})
	}

	// only send chapter mtg surveys on monday
	if weekday == 1 {
		survey(db, SurveyOptions{
			SurveyType:     "chapter meeting",
			QueryDate:      yesterday,
			QueryEventType: "",
			QueryEventName: `"Chapter Meeting"`,
			BodyText:       `Thank you for attending the chapter meeting! Please take this quick survey: https://docs.google.com/forms/d/e/1FAIpQLSfc_mgwH_zYYEQ5MTJwgyvCy5klsY_xrVBXgTDHM8sSxLIJrQ/viewform?usp=pp_url&entry.502269384=LINK_PARAM`,
			BodyHtml:       `<p>Thank you for attending the chapter meeting! Please <a href="https://docs.google.com/forms/d/e/1FAIpQLSfc_mgwH_zYYEQ5MTJwgyvCy5klsY_xrVBXgTDHM8sSxLIJrQ/viewform?usp=pp_url&entry.502269384=LINK_PARAM">click here</a> to take a quick survey.</p>`,
			LinkParam:      "date",
		})
	}
}

// Sends surveys based on event attendance every 60 minutes.
// Should be run in a goroutine.
func StartSurveyMailer(db *sqlx.DB) {
	for {
		log.Println("Starting survey mailer")
		surveyMailerWrapper(db)
		log.Println("Finished survey mailer")
		time.Sleep(60 * time.Minute)
	}
}
