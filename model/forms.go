package model

import (
	"fmt"

	"github.com/dxe/adb/mailer"

	"github.com/dxe/adb/mailing_list_signup"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ApplicationFormData struct {
	Name            string `json:"name" db:"name"`
	Email           string `json:"email" db:"email"`
	Address         string `json:"address" db:"address"`
	City            string `json:"city" db:"city"`
	Zip             string `json:"zip" db:"zip"`
	Phone           string `json:"phone" db:"phone"`
	Birthday        string `json:"birthday" db:"birthday"`
	Referral        string `json:"referral" db:"referral_apply"`
	ApplicationType string `json:"applicationType" db:"application_type"`
}

type InterestFormData struct {
	Form                      string `json:"form" db:"form"`
	Name                      string `json:"name" db:"name"`
	Email                     string `json:"email" db:"email"`
	Zip                       string `json:"zip" db:"zip"`
	Phone                     string `json:"phone" db:"phone"`
	ReferralFriends           string `json:"referralFriends" db:"referral_friends"`
	ReferralApply             string `json:"referralApply" db:"referral_apply"`
	ReferralOutlet            string `json:"referralOutlet" db:"referral_outlet"`
	Interests                 string `json:"interests" db:"interests"`
	SubmittedViaSignupService bool   `json:"submitted_via_signup_service"`
	DiscordID                 string `json:"discord_id" db:"discord_id"`
}

type InternationalFormData struct {
	FirstName   string  `json:"firstName" db:"first_name"`
	LastName    string  `json:"lastName" db:"last_name"`
	Email       string  `json:"email" db:"email"`
	Phone       string  `json:"phone" db:"phone"`
	Interest    string  `json:"interest" db:"interest"`
	Skills      string  `json:"skills" db:"skills"`
	Involvement string  `json:"involvement" db:"involvement"`
	City        string  `json:"city" db:"city"`
	State       string  `json:"state" db:"state"`
	Country     string  `json:"country" db:"country"`
	Lat         float64 `json:"lat" db:"lat"`
	Lng         float64 `json:"lng" db:"lng"`
}

type DiscordFormData struct {
	ID        string  `json:"id" db:"discord_id"`
	Token     string  `json:"token" db:"token"`
	FirstName string  `json:"firstName" db:"first_name"`
	LastName  string  `json:"lastName" db:"last_name"`
	Email     string  `json:"email" db:"email"`
	City      string  `json:"city" db:"city"`
	State     string  `json:"state" db:"state"`
	Country   string  `json:"country" db:"country"`
	Lat       float64 `json:"lat" db:"lat"`
	Lng       float64 `json:"lng" db:"lng"`
}

func SubmitApplicationForm(db *sqlx.DB, formData ApplicationFormData) error {
	_, err := db.NamedExec(`INSERT INTO form_application
		(email, name, phone, address, city, zip, birthday, application_type, referral_apply)
		VALUES
		(:email, :name, :phone, :address, :city, :zip, :birthday, :application_type, :referral_apply)
		`, formData)

	if err != nil {
		return errors.Wrap(err, "failed to insert application data")
	}

	signup := mailing_list_signup.Signup{
		Source: "adb-application-form",
		Name:   formData.Name,
		Email:  formData.Email,
		Phone:  formData.Phone,
		City:   formData.City,
		Zip:    formData.Zip,
	}
	err = mailing_list_signup.Enqueue(signup)
	if err != nil {
		// Don't return this error because we still want to successfully update the database.
		fmt.Println("ERROR adding application form submission to mailing list:", err.Error())
	}

	return nil
}

func SubmitInterestForm(db *sqlx.DB, formData InterestFormData) error {
	_, err := db.NamedExec(`INSERT INTO form_interest
		(form, email, name, phone, zip, referral_friends, referral_apply, referral_outlet, interests, discord_id)
		VALUES
		(:form, :email, :name, :phone, :zip, :referral_friends, :referral_apply, :referral_outlet, :interests, :discord_id)
		`, formData)

	if err != nil {
		return errors.Wrap(err, "failed to insert interest data")
	}

	if !formData.SubmittedViaSignupService {
		signup := mailing_list_signup.Signup{
			Source: "adb-interest-form",
			Name:   formData.Name,
			Email:  formData.Email,
			Phone:  formData.Phone,
			Zip:    formData.Zip,
		}
		err = mailing_list_signup.Enqueue(signup)
		if err != nil {
			// Don't return this error because we still want to successfully update the database.
			fmt.Println("ERROR adding application form submission to mailing list:", err.Error())
		}
	}

	return nil
}

func SubmitInternationalForm(db *sqlx.DB, formData InternationalFormData) error {
	_, err := db.NamedExec(`INSERT INTO form_international
		(first_name, last_name, email, phone, interest, skills, involvement, city, state, country, lat, lng)
		VALUES
		(:first_name, :last_name, :email, :phone, :interest, :skills, :involvement, :city, :state, :country, :lat, :lng)
		`, formData)

	if err != nil {
		return errors.Wrap(err, "failed to insert international form data")
	}

	nearestChapters, err := FindNearestChapters(db, formData.Lat, formData.Lng)
	if err != nil {
		panic(err)
	}
	nearestChapter := nearestChapters[0]
	fmt.Println("found nearest chapter!")
	fmt.Println(nearestChapter)

	/* TODO: Make an internal FindNearestChapters function that includes organizer info so we don't need to make two
	calls. (We don't want to include organizer info on the public endpoint.) */
	cc := []string{"jake@dxe.io", "vanas@umich.edu"}
	if nearestChapter.Email != "" {
		cc = append(cc, nearestChapter.Email)
	}
	nearestChapterDetails, err := GetChapterByID(db, nearestChapter.ChapterID)
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
	body := `<p>Hey ` + formData.FirstName + `!</p>
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

	signup := mailing_list_signup.Signup{
		Source:  "international-form",
		Name:    formData.FirstName + " " + formData.LastName,
		Email:   formData.Email,
		City:    formData.City,
		State:   formData.State,
		Country: formData.Country,
		Coords:  fmt.Sprintf("%.6f", formData.Lat) + "," + fmt.Sprintf("%.6f", formData.Lng),
	}
	err = mailing_list_signup.Enqueue(signup)
	if err != nil {
		// Don't return this error because we still want to successfully update the database.
		fmt.Println("ERROR adding international form submission to mailing list:", err.Error())
	}

	return nil
}

func SubmitDiscordForm(db *sqlx.DB, formData DiscordFormData) error {
	_, err := db.NamedExec(`INSERT INTO form_discord
		(first_name, last_name, email, city, state, country, lat, lng, discord_id)
		VALUES
		(:first_name, :last_name, :email, :city, :state, :country, :lat, :lng, :discord_id)
		`, formData)

	if err != nil {
		return errors.Wrap(err, "failed to insert discord form data")
	}

	signup := mailing_list_signup.Signup{
		Source:    "discord-form",
		Name:      formData.FirstName + " " + formData.LastName,
		Email:     formData.Email,
		City:      formData.City,
		State:     formData.State,
		Country:   formData.Country,
		Coords:    fmt.Sprintf("%.6f", formData.Lat) + "," + fmt.Sprintf("%.6f", formData.Lng),
		DiscordID: formData.ID,
	}
	err = mailing_list_signup.Enqueue(signup)
	if err != nil {
		// Don't return this error because we still want to successfully update the database.
		fmt.Println("ERROR adding discord form submission to mailing list:", err.Error())
	}

	return nil
}
