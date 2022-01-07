package international_mailer

import (
	"database/sql"
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
				<a href="http://nomorefactoryfarms.com">sign our petition to stop factory farms</a>.
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

func sendInternationalActionEmail(db *sqlx.DB, chapter model.ChapterWithToken) {
	subject := "Please report your last action & let us know how we can assist you"
	body := `
	<p>Hi DxE ` + chapter.Name + `,</p>

	<p>We are doing monthly check-ins to keep track of which chapters are organizing actions each quarter (either online or in person).
    <a href="` + fmt.Sprintf("https://adb.dxe.io/international_actions/%d/%v", chapter.ChapterID, chapter.EmailToken) + `">
    <strong>Please click here to provide your chapter's update for last month.</strong></a> Actions can be community events,  protests,
	campaigns, investigatory work, or any other type of nonviolent direct action that your chapter has decided to do. Please feel free to fill
	out the information in Spanish, if that is easier for you.
</p>`

	if chapter.Token == "" {
		body += `<p>Also note that we now have the ability to showcase your chapter's events on
		<a href="http://dxe.io/events">DxE's main website</a>, so they can be found by visitors who are looking for
		events in your area. In order for this to happen, please make
		<a href="https://www.facebook.com/cassie.king.399">Cassie King</a> an admin on your Facebook page.
        Feel free to reach out if you have any questions or concerns.</p>`
	}

	body += `<p>Thank you for all your hard work to create a better world for animals!</p>
	
	<p>Michelle Del Cueto<br />
	International Coordinator<br />
	Direct Action Everywhere</p>
	`

	var toEmails []string
	if chapter.Email != "" {
		toEmails = append(toEmails, chapter.Email)
	}
	if len(chapter.Organizers) > 0 {
		for _, o := range chapter.Organizers {
			if o.Email != "" {
				toEmails = append(toEmails, o.Email)
			}
		}
	}

	if len(toEmails) == 0 {
		log.Printf("Can't send international actions email to %v due to missing email address\n", chapter.Name)
		return
	}

	log.Printf("Sending int'l action email to %v: %v\n", chapter.Name, strings.Join(toEmails, ","))
	err := mailer.Send(mailer.Message{
		FromName:    "Michelle Del Cueto",
		FromAddress: "internationalcoordination@directactioneverywhere.com",
		ToName:      "DxE " + chapter.Name,
		ToAddress:   toEmails[0],
		Subject:     subject,
		BodyHTML:    body,
		CC:          append(toEmails[1:], "jake@directactioneverywhere.com"),
	})
	if err != nil {
		log.Println("Failed to send email for international actions email")
		return
	}

	chapter.LastCheckinEmailSent = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	_, err = model.UpdateChapter(db, chapter)
	if err != nil {
		panic("Failed to update chapter last check-in email sent time " + err.Error())
	}
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

func internationalActionMailerWrapper(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in int'l action mailer", r)
		}
	}()

	// Only run on the 1st or 7th of the month b/w 4pm and midnight UTC (9am-5pm PT).
	now := time.Now()
	if now.Day() != 1 && now.Day() != 7 {
		return
	}
	if now.Hour() < 16 || now.Hour() > 23 {
		return
	}

	// Calculate first day of current month.
	y, m, _ := now.Date()
	startOfCurrentMonth := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())

	chapters, err := model.GetAllChapters(db)
	if err != nil {
		panic("Failed to get chapters for int'l action mailer " + err.Error())
	}
	for _, chap := range chapters {

		if chap.Region == "Online" {
			continue
		}

		neverSentEmail := !chap.LastCheckinEmailSent.Valid
		sentEmailBeforeCurrentMonth := chap.LastCheckinEmailSent.Valid && chap.LastCheckinEmailSent.Time.Before(startOfCurrentMonth)
		sentEmailToday := chap.LastCheckinEmailSent.Valid && chap.LastCheckinEmailSent.Time.Year() == now.Year() && chap.LastCheckinEmailSent.Time.YearDay() == now.YearDay()

		switch now.Day() {
		case 1:
			if neverSentEmail || sentEmailBeforeCurrentMonth {
				sendInternationalActionEmail(db, chap)
			}
		case 7:
			if neverSentEmail || sentEmailToday {
				continue
			}
			if chap.LastContact == "" {
				sendInternationalActionEmail(db, chap)
				continue
			}
			dateLayout := "2006-01-02"
			lastContactDate, err := time.Parse(dateLayout, chap.LastContact)
			if err != nil {
				log.Printf("Error parsing last contact date for chapter %v\n", chap.Name)
				continue
			}
			if lastContactDate.Before(startOfCurrentMonth) {
				// Chapter hasn't responded this month, so send the email again.
				sendInternationalActionEmail(db, chap)
			}
		default:
			continue
		}
	}
}

func internationalActionFormProcessor(db *sqlx.DB) {
	newResponses, err := model.GetUnprocessedInternationalActionFormResponses(db)
	if err != nil {
		log.Println("Error getting new int'l action form responses to process", err.Error())
		return
	}
	for _, form := range newResponses {
		chap, err := model.GetChapterByID(db, form.ChapterID)
		if err != nil {
			log.Println("Error looking up chapter for int'l action form response", form.ID, err.Error())
			continue
		}
		chap.LastContact = form.SubmittedAt.Time.Format("2006-01-02")
		if form.LastAction != "" {
			chap.LastAction = form.LastAction
		}
		_, err = model.UpdateChapter(db, chap)
		if err != nil {
			log.Println("Failed to update chapter with int'l action form response data", form.ID, err.Error())
			continue
		}
		if form.Needs != "" {
			var chapEmails []string
			if chap.Email != "" {
				chapEmails = append(chapEmails, chap.Email)
			}
			if len(chap.Organizers) > 0 {
				for _, o := range chap.Organizers {
					if o.Email != "" {
						chapEmails = append(chapEmails, o.Email)
					}
				}
			}
			emailLink := fmt.Sprintf(`https://mail.google.com/mail/?view=cm&fs=1&su=%v&to=%v`, chap.Name, strings.Join(chapEmails, ","))

			body := fmt.Sprintf("<p>%v</p>", form.Needs)
			body += fmt.Sprintf(`<p><a href="%v">Click here to reply to %v</a></p>`, emailLink, chap.Name)

			err = mailer.Send(mailer.Message{
				FromName:         "DxE International Action Form",
				FromAddress:      "noreply@directactioneverywhere.com",
				ToName:           "International Coordination",
				ToAddress:        "internationalcoordination@directactioneverywhere.com",
				ReplyToAddresses: chapEmails,
				Subject:          fmt.Sprintf("Assistance needed for %v (%v)", chap.Name, form.OrganizerName),
				BodyHTML:         body,
				CC:               []string{"jake@dxe.io"},
			})
			if err != nil {
				log.Println("Failed to send email to int'l coordination for int'l action form", form.ID)
			}
		}
		err = model.MarkInternationalActionFormProcessed(db, form.ID)
		if err != nil {
			log.Println("Failed to mark int'l action form as processed", form.ID, err.Error())
			continue
		}
	}
}

// Sends emails every 60 minutes.
// Should be run in a goroutine.
func StartInternationalMailer(db *sqlx.DB) {
	for {
		log.Println("Starting international mailer")
		internationalMailerWrapper(db)
		log.Println("Finished international mailer")

		log.Println("Starting international action mailer")
		internationalActionMailerWrapper(db)
		log.Println("Finished international action mailer")

		time.Sleep(60 * time.Minute)
	}
}

// Process International Action form responses every 5 minutes.
// Should be run in a goroutine.
func StartInternationalActionFormProcessor(db *sqlx.DB) {
	for {
		log.Println("Starting international action form processor")
		internationalActionFormProcessor(db)
		log.Println("Finished international action form processor")
		time.Sleep(5 * time.Minute)
	}
}
