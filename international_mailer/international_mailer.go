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
	nearestChapter := nearestChapters[0]

	cc := []string{"jake@dxe.io", "vanas@umich.edu"}
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
	subject := "Join your local Direct Action Everywhere chapter!"
	body := `<p>Hey ` + strings.Title(formData.FirstName) + `!</p>
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

		subject = "Getting involved with Direct Action Everywhere"
		body = `<p>Hey ` + formData.FirstName + `!</p>
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

	fmt.Printf("int'l mailer sending email to %v\n", formData.Email)

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
		fmt.Println("failed to send email for international form submission")
	}

	err = model.UpdateInternationalFormSubmissionEmailStatus(db, formData.ID)
	if err != nil {
		fmt.Println("error updating international for submission email status")
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
		panic("failed to get int'l form submissions to email " + err.Error())
	}
	fmt.Printf("int'l form mailer found %d records to process\n", len(records))

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
