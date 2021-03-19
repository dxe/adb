package model

import (
	"fmt"

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
		(form, email, name, phone, zip, referral_friends, referral_apply, referral_outlet, interests)
		VALUES
		(:form, :email, :name, :phone, :zip, :referral_friends, :referral_apply, :referral_outlet, :interests)
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

	// TODO: sign up to signup service w/ proper fields
	//signup := mailing_list_signup.Signup{
	//	Source: "adb-interest-form",
	//	Name:   formData.Name,
	//	Email:  formData.Email,
	//	Phone:  formData.Phone,
	//	Zip:    formData.Zip,
	//}
	//err = mailing_list_signup.Enqueue(signup)
	//if err != nil {
	//	// Don't return this error because we still want to successfully update the database.
	//	fmt.Println("ERROR adding application form submission to mailing list:", err.Error())
	//}

	return nil
}
