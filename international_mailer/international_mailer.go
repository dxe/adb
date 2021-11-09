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
	msg.ToName = formData.FirstName + " " + formData.LastName
	msg.ToAddress = formData.Email
	msg.CC = append(msg.CC, "jake@directactioneverywhere.com") // TODO: remove after testing in prod
	msg.CC = append(msg.CC, "internationalcoordination@directactioneverywhere.com")

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

		// assemble the message
		msg.FromName = "Anastasia Rogers"
		msg.FromAddress = "arogers@directactioneverywhere.com"
		msg.ReplyToAddress = "vanas@umich.edu"
		msg.Subject = "Join your local Direct Action Everywhere chapter!"
		msg.BodyHTML = `
			<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
			<p>
				My name is Anastasia and I’m an organizer with Direct Action Everywhere. I wanted to reach out about your
				inquiry to get involved in our international network. There is a DxE chapter near you, so I’ve included
				their information below so you can reach out and get involved with them!
			</p> 
			<p>
				` + contactInfo + `
				I’ve also cc’ed the organizers in your local chapter on this email so that they can reach out as well.
			</p> 
			<p>
				In the meantime there are a few actions you could take. First you can
				<a href="http://dxe.io/discord">join our Discord server</a>. Next you can
				<a href="http://nomorefactoryfarms.com">sign our petition to stop factory farms</a>.
			</p>
			<p>Let me know if you have any questions or if you still have trouble connecting with your local chapter!</p>
			<p>
				In Solidarity,<br/>
				Anastasia Rogers<br/>
				Direct Action Everywhere Organizer
			</p>
		`

	default:
		msg.FromName = "Michelle Del Cueto"
		msg.FromAddress = "michelle@directactioneverywhere.com"
		msg.ReplyToAddress = "michelle@directactioneverywhere.com"
		msg.Subject = "Getting involved with Direct Action Everywhere"
		msg.BodyHTML = `
			<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
			<p>I saw that you showed interest in getting involved with our international network.</p>
			<p>
				Currently, there isn’t a DxE chapter in your city, but if you are interested in starting a chapter and
				organizing actions or events that would help mobilize your community for animal rights, the international coordination
				team can help you. Sometimes, the thought of "organizing" and starting a chapter from zero can feel really intimidating,
				but we have multiple resources and a mentorship program to support you.
			</p>
			<p>
				The next step is to watch
				<a href="https://www.dropbox.com/s/4dusc12v35u5lfb/How%20to%20Change%20the%20World%20Nov%202020.mp4?dl=0">this workshop</a>
				of our theory of change, so you can become more familiar with our mission and strategy.   
			</p>
			<p>
				Next, please join our onboarding calls that are hosted on Zoom on the first and third Wednesday of every
				month, at 11am Pacific Time (6pm GMT). This link will take you to the onboarding calls at that time:
				<a href="https://dxe.io/ioonboarding">dxe.io/ioonboarding</a>.
			</p>
			<p>Hope that you can join us! Let me know if you have any questions.</p>
			<p>
				Michelle Del Cueto<br/>
				Direct Action Everywhere International Coordinator
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
			<p>Interest: %v</p>
			<p>Skills: %v</p>
	`, formData.FirstName, formData.LastName, formData.Email, formData.Phone, formData.City, formData.Interest, formData.Skills)

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
	<p>Hi all!</p>

	<p>We are doing monthly check-ins to keep track of which chapters are organizing actions each quarter (either online
	or in person).
    <a href="` + fmt.Sprintf("https://adb.dxe.io/international_actions/%d/%v", chapter.ChapterID, chapter.EmailToken) + `">
    <strong>Please click here to provide your chapter's update for last month.</strong></a> Actions need not be an
    elaborate protest, especially during the pandemic, and can simply just be a social media
	challenge or organizing your community members to email representatives or businesses with an ask.</p>`

	if chapter.Token == "" {
		body += `<p>Also note that we now have the ability to showcase your chapter's events on
		<a href="http://dxe.io/events">DxE's main website</a>, so they can be found by visitors who are looking for
		events in your area. In order for this to happen, please make
		<a href="https://www.facebook.com/jhobbs91">Jake Hobbs</a> (our Tech team lead) an admin on your Facebook page.
        He won't read your messages or interact with your page at all, other than setting up an automated system that will 
		read the public events from your Facebook page. Feel free to reach out if you have any questions or concerns.</p>`
	}

	body += `<p>Thank you for all that you do,</p>
	
	<p>Paul Darwin Picklesimer<br />
	Direct Action Everywhere<br />
	International Coordination Working Group<br />
	(304) 479-3366</p>
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
		FromName:    "Paul Darwin Picklesimer",
		FromAddress: "paul@directactioneverywhere.com",
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
