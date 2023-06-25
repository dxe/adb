package international_mailer

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"

	"github.com/jmoiron/sqlx"
)

func processFormSubmission(db *sqlx.DB, formData model.InternationalFormData) {
	nearestChapters, err := model.FindNearestChapters(db, formData.Lat, formData.Lng)
	if err != nil {
		panic(err)
	}

	var nearestChapter *model.ChapterWithToken
	for _, chapter := range nearestChapters {
		if chapter.Distance > 150 {
			break
		}
		if chapter.Country == formData.Country {
			nearestChapter = &chapter
			break
		}
	}

	err = sendInternationalOnboardingEmail(db, formData, nearestChapter)
	if err != nil {
		panic(err)
	}

	err = model.UpdateInternationalFormSubmissionEmailStatus(db, formData.ID)
	if err != nil {
		log.Println("Error updating international form submission email status")
	}
}

func sendInternationalOnboardingEmail(db *sqlx.DB, formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	var msg mailer.Message
	msg.FromName = "Michelle Del Cueto"
	msg.FromAddress = "internationalcoordination@directactioneverywhere.com"
	msg.ToName = formData.FirstName + " " + formData.LastName
	msg.ToAddress = formData.Email

	switch chapter != nil {
	case true:
		if chapter.Name == "SF Bay Area" {
			return nil
		}

		// append CC list
		if chapter.Email != "" {
			msg.CC = append(msg.CC, chapter.Email)
		}
		chapterDetails, err := model.GetChapterByID(db, chapter.ChapterID)
		if err != nil {
			panic(err)
		}
		organizers := chapterDetails.Organizers
		if len(organizers) > 0 {
			for _, o := range organizers {
				if o.Email != "" {
					msg.CC = append(msg.CC, o.Email)
				}
			}
		}

		err = sendInternationalAlertEmail(formData, msg.CC)
		if err != nil {
			log.Printf("Error sending int'l alert email: %v\n", err.Error())
		}

		// build contact info links
		var contactInfo string
		socialLinks := map[string]string{
			"Facebook page": chapter.FbURL,
			"Instagram":     chapter.InstaURL,
			"Twitter":       chapter.TwitterURL,
			"Email":         chapter.Email,
		}
		for k, v := range socialLinks {
			if v != "" {
				contactInfo += fmt.Sprintf(`<a href="%v">%v %v</a><br />`, v, chapter.Name, k)
			}
		}

		// check if chapter has an upcoming event
		var nextEvent model.ExternalEvent
		if chapter.ID != 0 {
			startTime := time.Now()
			endTime := time.Now().Add(60 * 24 * time.Hour)
			events, _ := model.GetExternalEvents(db, chapter.ID, startTime, endTime, false)
			if len(events) > 0 {
				nextEvent = events[0]
			}
		}

		// assemble the message
		msg.Subject = "Join your local Direct Action Everywhere chapter!"
		msg.BodyHTML = `
			<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
			<p>
				I wanted to reach out about your inquiry of getting involved with DxE’s international network. There is
				currently a DxE chapter near you, so I’ve included their information and contact below so you can reach out,
				get involved, and start taking action with them!

			</p> 
			<p>
				` + contactInfo + `
			</p>
			<p>I’ve also cc'd the organizers in your local chapter on this email, so you can both be in contact.</p>
		`

		if nextEvent.ID != 0 {
			msg.BodyHTML += fmt.Sprintf(`
				<p>You can also find details of their next event here: <a href="https://facebook.com/%v">%v</a>.</p>
			`, nextEvent.ID, nextEvent.Name)
		}

		msg.BodyHTML += `
			<p>
				In the meantime you can
				<a href="https://righttorescue.com/">sign a letter to support the right to rescue</a>.
			</p>
			<p>Let me know if you have any questions or if you still haven't been able to connect with your local chapter.</p>
			<p>Hope that you can join us!</p>
			<p>
				<strong>Michelle Del Cueto</strong><br/>
				International Coordinator<br/>
				Direct Action Everywhere
			</p>
		`

	default:
		msg.Subject = "Getting involved with Direct Action Everywhere"
		msg.BodyHTML = `
			<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
			<p>I saw that you showed interest in getting involved with our international network.</p>
			<p>
				Currently, there isn’t a DxE chapter in your city, but if you are interested in starting a chapter and
				organizing actions or events that would help mobilize your community for animal rights, the international
				coordination team is here to help you!
			</p>
            <p>
 				We just launched a Workshop on How to Organize a DxE Chapter. I highly encourage you to attend, so you
 				can learn about our mission, strategy and everything else you need to know to be involved with DxE. Here
 				is the event link, please RSVP:
 				<a href="https://dxe.io/organizedxechapter">dxe.io/organizedxechapter</a>.
 			</p>
 			<p>
 				I really hope to see you there, and if you have any questions please let me know.
 			</p>
			<p>
				<strong>Michelle Del Cueto</strong><br/>
				International Coordinator<br/>
				Direct Action Everywhere
			</p>
		`
	}

	log.Printf("Int'l mailer sending email to %v\n", formData.Email)

	err := mailer.Send(msg)
	if err != nil {
		log.Println("Failed to send email for international form submission")
	}

	return nil
}

func sendInternationalAlertEmail(formData model.InternationalFormData, to []string) error {
	if len(to) == 0 {
		return nil
	}

	msg := mailer.Message{
		FromName:    "DxE Join Form",
		FromAddress: "noreply@directactioneverywhere.com",
		ToAddress:   to[0],
		Subject:     fmt.Sprintf("%v %v signed up to join your chapter", formData.FirstName, formData.LastName),
	}

	if len(to) > 1 {
		msg.CC = to[1:]
	}

	msg.BodyHTML = fmt.Sprintf(`
			<p>Name: %v %v</p>
			<p>Email: %v</p>
			<p>Phone: %v</p>
			<p>City: %v</p>
			<p>Involvement: %v</p>
			<p>Skills: %v</p>
	`, formData.FirstName, formData.LastName, formData.Email, formData.Phone, formData.City, formData.Involvement, formData.Skills)

	log.Println("Int'l mailer sending alert email")
	err := mailer.Send(msg)
	if err != nil {
		log.Println("Failed to send int'l alert email")
	}
	return nil
}

func internationalMailerWrapper(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in international mailer", r)
		}
	}()

	records, err := model.GetInternationalFormSubmissionsToEmail(db)
	if err != nil {
		panic("Failed to get int'l form submissions to email " + err.Error())
	}
	log.Printf("Int'l form mailer found %d records to process\n", len(records))

	for _, rec := range records {
		processFormSubmission(db, rec)
	}
}

// Sends emails every 60 minutes.
// Should be run in a goroutine.
func StartInternationalMailer(db *sqlx.DB) {
	for {
		log.Println("Starting international mailer")
		internationalMailerWrapper(db)
		log.Println("Finished international mailer")

		time.Sleep(60 * time.Minute)
	}
}
