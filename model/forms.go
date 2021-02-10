package model

import (
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
	Form            string `json:"form" db:"form"`
	Name            string `json:"name" db:"name"`
	Email           string `json:"email" db:"email"`
	Zip             string `json:"zip" db:"zip"`
	Phone           string `json:"phone" db:"phone"`
	ReferralFriends string `json:"referralFriends" db:"referral_friends"`
	ReferralApply   string `json:"referralApply" db:"referral_apply"`
	ReferralOutlet  string `json:"referralOutlet" db:"referral_outlet"`
	Interests       string `json:"interests" db:"interests"`
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

	return nil
}
