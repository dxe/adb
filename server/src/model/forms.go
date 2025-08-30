package model

import (
	"fmt"

	"github.com/dxe/adb/mailing_list_signup"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ApplicationFormData struct {
	Name            string `json:"name" db:"name"`
	Email           string `json:"email" db:"email"`
	Address         string `json:"address" db:"address"`
	City            string `json:"city" db:"city"`
	Zip             string `json:"zip" db:"zip"`
	Phone           string `json:"phone" db:"phone"`
	Birthday        string `json:"birthday" db:"birthday"`
	Pronouns        string `json:"pronouns" db:"pronouns"`
	Referral        string `json:"referral" db:"referral_apply"`
	Language        string `json:"language" db:"language"`
	Accessibility   string `json:"accessibility" db:"accessibility"`
	ApplicationType string `json:"applicationType" db:"application_type"`
}

type InterestFormData struct {
	ChapterId                 int    `json:"chapterId" db:"chapter_id"`
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
	ID          int     `json:"id" db:"id"`
	FirstName   string  `json:"firstName" db:"first_name"`
	LastName    string  `json:"lastName" db:"last_name"`
	Email       string  `json:"email" db:"email"`
	Phone       string  `json:"phone" db:"phone"`
	Interest    string  `json:"interest" db:"interest"`       // "organize" or "participate"
	Skills      string  `json:"skills" db:"skills"`           // Deprecated
	Involvement string  `json:"involvement" db:"involvement"` // skills and interests
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
		(email, name, pronouns, phone, address, city, zip, birthday, application_type, referral_apply, language, accessibility)
		VALUES
		(:email, :name, :pronouns, :phone, :address, :city, :zip, :birthday, :application_type, :referral_apply, :language, :accessibility)
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
		// Don't fail the HTTP request since at least the user's response was added to the database.
		log.Printf("ERROR adding application form submission to mailing list: %v", err)
	}
	log.Printf("Enqueued email for sign-up: %v", formData.Email)

	return nil
}

func SubmitInterestForm(db *sqlx.DB, formData InterestFormData) error {
	_, err := db.NamedExec(`INSERT INTO form_interest
		(chapter_id, form, email, name, phone, zip, referral_friends, referral_apply, referral_outlet, interests, discord_id)
		VALUES
		(:chapter_id, :form, :email, :name, :phone, :zip, :referral_friends, :referral_apply, :referral_outlet, :interests, :discord_id)
		`, formData)

	if err != nil {
		return errors.Wrap(err, "failed to insert interest data")
	}
	log.Printf("Saved interest form response for %v", formData.Email)

	if !formData.SubmittedViaSignupService {
		signup := mailing_list_signup.Signup{
			Source: "adb-interest-form",
			// Subscribe responder to the chapter that owns the form in addition to any chapter
			// near the responder's zip code.
			SourceChapterId: formData.ChapterId,
			Name:            formData.Name,
			Email:           formData.Email,
			Phone:           formData.Phone,
			Zip:             formData.Zip,
		}
		err = mailing_list_signup.Enqueue(signup)
		if err != nil {
			// Don't fail the HTTP request since at least the user's response was added to the database.
			log.Error().Msgf("ERROR adding interest form submission to mailing list: %v", err)
		}
		log.Printf("Enqueued email for sign-up: %v; source chapter id: %v; target chapter zip code: %v", formData.Email, formData.ChapterId, formData.Zip)
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
		// Don't return this error because we still want to indicate successfully updating the database.
		log.Printf("ERROR adding international application form submission to mailing list: %v", err.Error())
	}
	log.Printf("Enqueued email for sign-up: %v", formData.Email)

	return nil
}

func GetInternationalFormSubmissionsToEmail(db *sqlx.DB) ([]InternationalFormData, error) {
	query := `SELECT id, first_name, last_name, email, phone, interest, involvement, city, state, country, lat, lng
from form_international WHERE form_submitted is not null AND email_sent is null`

	var submissions []InternationalFormData
	err := db.Select(&submissions, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select int'l form submissions")
	}

	return submissions, nil
}

func UpdateInternationalFormSubmissionEmailStatus(db *sqlx.DB, id int) error {
	_, err := db.Exec(`UPDATE form_international
		SET email_sent = CURRENT_TIMESTAMP
		WHERE id = ?
		`, id)

	if err != nil {
		return errors.Wrap(err, "failed to update international form submission email status")
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
		log.Printf("ERROR adding discord form submission to mailing list: %v", err)
	}
	log.Printf("Enqueued email for sign-up: %v", formData.Email)

	return nil
}
