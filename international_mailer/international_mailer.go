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
	nearestChapter := nearestChapters[0]

	var cc []string
	if nearestChapter.Email != "" {
		cc = append(cc, nearestChapter.Email)
	}
	nearestChapterDetails, err := model.GetChapterByID(db, nearestChapter.ChapterID)
	if err != nil {
		panic(err)
	}
	organizers := nearestChapterDetails.Organizers
	if len(organizers) > 0 {
		for _, o := range organizers {
			if o.Email != "" {
				cc = append(cc, o.Email)
			}
		}
	}

	// Send an email to the person who submitted the form.
	subject := "Getting involved with Direct Action Everywhere"
	body := `<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
<p>My name is Anastasia and I’m an organizer with Direct Action Everywhere. I wanted to reach out about your inquiry to get involved in our international network.</p>
<p>We don’t currently have a DxE chapter in your city, and at the moment, getting involved with a chapter is the main way we have for people around the world to get involved. However, we have some actions you could take to get started! First you can <a href="http://dxe.io/discord">join our Discord server</a>. Next you can <a href="http://nomorefactoryfarms.com">sign our petition to stop factory farms</a>. Most importantly you can <a href="http://dxe.io/workshop">attend our next Zoom workshop for new and aspiring activists</a>.</p>
<p>In the meantime, I wanted to reach out and see if you want to chat about the possibility of starting a chapter. Sometimes, the thought of "organizing" or starting a chapter can feel really intimidating, but we have a team here to support all our organizers and help you mobilize your community. If you’re open to it, I’d love to give you more information about what’s involved – let me know!</p> 
<p>Let me know if you have any questions!</p>
<p>In Solidarity,<br/>
Anastasia Rogers<br/>
Direct Action Everywhere Organizer</p>
`
	if nearestChapter.Distance < 150 {
		var contactInfo string
		if nearestChapter.FbURL != "" {
			contactInfo += fmt.Sprintf(`<a href="%v">%v Facebook page</a><br />`, nearestChapter.FbURL, nearestChapter.Name)
		}
		if nearestChapter.Email != "" {
			contactInfo += fmt.Sprintf(`Email address: <a href="mailto:%v">%v</a><br />`, nearestChapter.Email, nearestChapter.Email)
		}

		subject = "Join your local Direct Action Everywhere chapter!"
		body = `<p>Hey ` + strings.Title(strings.TrimSpace(formData.FirstName)) + `!</p>
<p>My name is Anastasia and I’m an organizer with Direct Action Everywhere. I wanted to reach out about your inquiry to get involved in our international network. There is a DxE chapter near you, so I’ve included their information below so you can reach out and get involved with them!</p> 
<p>` + contactInfo + `
I’ve also cc’ed the organizers in your local chapter on this email so that they can reach out as well.</p> 
<p>In the meantime there are a few actions you could take. First you can <a href="http://dxe.io/discord">join our Discord server</a>. Next you can <a href="http://nomorefactoryfarms.com">sign our petition to stop factory farms</a>. Most importantly you can <a href="http://dxe.io/workshop">attend our next Zoom workshop for new and aspiring activists</a>.</p>
<p>Let me know if you have any questions or if you still have trouble connecting with your local chapter after attending the workshop!</p>
<p>In Solidarity,<br/>
Anastasia Rogers<br/>
Direct Action Everywhere Organizer</p>
`
	}

	if nearestChapter.Name != "SF Bay Area" {
		log.Printf("Int'l mailer sending email to %v\n", formData.Email)

		err = mailer.Send(mailer.Message{
			FromName:       "Anastasia Rogers",
			FromAddress:    "arogers@directactioneverywhere.com",
			ToName:         formData.FirstName + " " + formData.LastName,
			ToAddress:      formData.Email,
			ReplyToAddress: "vanas@umich.edu",
			Subject:        subject,
			BodyHTML:       body,
			CC:             cc,
		})
		if err != nil {
			log.Println("Failed to send email for international form submission")
		}
	}

	err = model.UpdateInternationalFormSubmissionEmailStatus(db, formData.ID)
	if err != nil {
		log.Println("Error updating international for submission email status")
	}

}

func sendInternationalActionEmail(db *sqlx.DB, chapter model.ChapterWithToken) {

	subject := "PLEASE READ - New database of DxE chapters and May Global Strategy Call"
	body := `
	<p>Hi all!</p>
	
	<p>Exciting news from our tech team! They have created an ADB (activist database) that will help automate some of the
	International Coordination Working Group’s messaging and tasks and will automatically track international actions that
	have facebook event pages if you are just able to make our tech team lead <a href="https://www.facebook.com/jhobbs91">Jake Hobbs</a> an admin of your chapter’s facebook
	page. He won’t read your messages or interact with your page at all except to create and maintain an automated portal
	that puts your upcoming event pages on <a href="http://dxe.io/events">DxE’s website</a> so they can be found by visitors who are looking for events in
	your area. If you haven’t already, please add <a href="https://www.facebook.com/jhobbs91">Jake</a> as an admin now.</p>
	
	<p>Also, with the new ADB, we’ll be better able to keep track of assuring that all chapters are organizing actions
	(online or in person) each quarter in order to remain as an active chapter in the DxE International Organizers Network.
	In place of the monthly report forms that we’ve used lately, your chapter will automatically receive this email on the
	first and seventh of each month with <a href="` + fmt.Sprintf("https://adb.dxe.io/international_actions/%d/%v", chapter.ChapterID, chapter.EmailToken) + `">a link to a short form to report your previous last actions or ask for any assistance</a>.
	Please keep an eye out for the email on the 1st of the month so we can be sure of any actions that you did the previous
	month. Actions need not be an elaborate protest, especially during the pandemic, and can simply just be a social media
	challenge or organizing your community members to email representatives or businesses with an ask.</p>
	
	<p>If you aren't able to do an action in a quarter, we will remove your chapter and invite you to return later if you
	like or we can adjust your chapter’s status to “Hiatus” temporarily, but ideally, we’d just love for you to take some
	form of action even if it’s online. Of course, In-person actions are great so if conditions permit, please consider
	organizing a <a href="https://www.facebook.com/events/276596890763352">Let Dairy Die protest</a> on or around Mothers Day
	this month or <a href="https://docs.google.com/spreadsheets/d/1-y_r8BgepiHpnOYyzoJISn30gNHuChUTBYxSUGWFSMg/edit?usp=sharing">list your address here</a> if you’d like to be
	sent materials to do postering or a banner drop promoting Netflix’s Seaspiracy in your city. As always, also
	please just consider organising whatever action works for your chapter’s goals. Thank you!</p>
	
	<p>We can discuss and share more during the <a href="https://www.facebook.com/events/126551249410206">May Global Strategy Call</a> tomorrow, Sunday, May 2nd at 10am Pacific.
	We hope to see you there!</p>
	
	<p>Thank you for all that you do,</p>
	
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

	// send to each person in the chapter
	for _, e := range toEmails {
		fmt.Printf("Sending int'l action email to %v\n", e)
		err := mailer.Send(mailer.Message{
			FromName:    "Paul Darwin Picklesimer",
			FromAddress: "paul@directactioneverywhere.com",
			ToName:      "DxE " + chapter.Name,
			ToAddress:   e,
			Subject:     subject,
			BodyHTML:    body,
			CC:          []string{"jake@directactioneverywhere.com"},
		})
		if err != nil {
			log.Println("Failed to send email for international form submission")
		}
	}

	chapter.LastCheckinEmailSent = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	_, err := model.UpdateChapter(db, chapter)
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

	// Only run on the 1st or 8th of the month b/w 4pm and midnight UTC (9am-5pm PT).
	now := time.Now()
	if now.Day() != 1 && now.Day() != 8 {
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
			dateLayout := "2006-01-02"
			lastContactDate, err := time.Parse(dateLayout, chap.LastContact)
			if err != nil {
				fmt.Printf("Error parsing last contact date for chapter %v", chap.Name)
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
		fmt.Println("Error getting new int'l action form responses to process", err.Error())
		return
	}
	for _, form := range newResponses {
		chap, err := model.GetChapterByID(db, form.ChapterID)
		if err != nil {
			fmt.Println("Error looking up chapter for int'l action form response", form.ID, err.Error())
			continue
		}
		chap.LastContact = form.SubmittedAt.Time.Format("2006-01-02")
		if form.LastAction != "" {
			chap.LastAction = form.LastAction
		}
		_, err = model.UpdateChapter(db, chap)
		if err != nil {
			fmt.Println("Failed to update chapter with int'l action form response data", form.ID, err.Error())
			continue
		}
		if form.Needs != "" {
			err = mailer.Send(mailer.Message{
				FromName:       "DxE International Action Form",
				FromAddress:    "noreply@directactioneverywhere.com",
				ToName:         "International Coordination",
				ToAddress:      "internationalcoordination@directactioneverywhere.com",
				ReplyToAddress: "tech@dxe.io",
				Subject:        fmt.Sprintf("Assistance needed for %v (%v)", chap.Name, form.OrganizerName),
				BodyHTML:       fmt.Sprintf("<p>%v</p>", form.Needs),
				CC:             []string{"jake@dxe.io"},
			})
			if err != nil {
				log.Println("Failed to send email to int'l coordination for int'l action form", form.ID)
			}
		}
		err = model.MarkInternationalActionFormProcessed(db, form.ID)
		if err != nil {
			fmt.Println("Failed to mark int'l action form as processed", form.ID, err.Error())
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
