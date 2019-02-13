package model

import (
	"database/sql"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/** Constant and Variable Definitions */

const Duration60Days = 60 * 24 * time.Hour
const Duration90Days = 90 * 24 * time.Hour

const ACTIVIST_LEVEL_CHAPTER_MEMBER = "Chapter Member"

const selectActivistBaseQuery string = `
SELECT
  email,
  facebook,
  id,
  location,
  name,
  phone
FROM activists
`

// Query perf tips:
//  - Test changes to this query against production data to see
//    how they effect performance
//  - It seems like it's usually faster to use subqueries in the top
//    part of the SELECT expression vs joining on a table.
const selectActivistExtraBaseQuery string = `
SELECT

  email,
  facebook,
  a.id,
  location,
  a.name,
  phone,
  dob,

  activist_level,
  source,

  connector,
  training0,
  training1,
  training2,
  training3,
  training4,
  training5,
  training6,
  dev_application_date,
  dev_manager,
  dev_interest,
  dev_auth,
  dev_email_sent,
  dev_vetted,
  dev_interview,
  dev_onboarding,

  prospect_senior_organizer,
  so_auth,
  so_core,
  so_agreement,
  so_training,
  so_quiz,
  so_connector,
  so_onboarding,

  cm_first_email,
  cm_approval_email,
  cm_warning_email,
  cir_first_email,
  prospect_organizer,
  prospect_chapter_member,
  circle_agreement,
  referral_friends,
  referral_apply,
  referral_outlet,
  circle_interest,
  interest_date,

          @first_event := (
      SELECT min(e.date) AS min_date
      FROM event_attendance ea
      JOIN activists inner_a ON inner_a.id = ea.activist_id
      JOIN events e ON e.id = ea.event_id
      WHERE inner_a.id = a.id
  ) AS first_event,

  @last_event := (
    SELECT max(e.date) AS max_date
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id
  ) AS last_event,

  @last_circle := (
    SELECT max(e.date) AS max_date
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id and e.event_type = 'Circle'
  ) AS last_circle,

      IFNULL(
    concat(@first_event, ' ', (
      SELECT name
      FROM events e
      JOIN event_attendance ea ON ea.event_id = e.id
      WHERE
	e.date = first_event
        AND ea.activist_id = a.id
      LIMIT 1)),
    '') AS first_event_name,

  IFNULL(
    concat(@last_event, ' ', (
      SELECT name
      FROM events e
      JOIN event_attendance ea ON ea.event_id = e.id
      WHERE
        e.date = last_event
        AND ea.activist_id = a.id
      LIMIT 1)),
    '') AS last_event_name,

  (SELECT COUNT(DISTINCT ea.event_id)
    FROM event_attendance ea
    WHERE
      ea.activist_id = a.id) as total_events,

  IFNULL(totalPoints, 0) as total_points,
  IF(@last_event >= (now() - interval 30 day), 1, 0) as active,

  IFNULL(
    (SELECT
      GROUP_CONCAT(DISTINCT wg.name SEPARATOR ', ')
    FROM working_groups wg
    JOIN working_group_members wgm ON wg.id = wgm.working_group_id
    WHERE
      wgm.activist_id = a.id and wgm.non_member_on_mailing_list = 0),
    '') AS working_group_list,
    IFNULL(
    (SELECT
      GROUP_CONCAT(DISTINCT circles.name SEPARATOR ', ')
    FROM circles
    JOIN circle_members ON circles.id = circle_members.circle_id
    WHERE
      circle_members.activist_id = a.id),
    '') AS circles_list,

    IFNULL((
    SELECT max(e.date) AS max_date
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id and e.event_type = "Connection"
  	),"") AS last_connection,

  	IF((CONCAT( (IFNULL((SELECT GROUP_CONCAT(DISTINCT wg.name SEPARATOR ', ') FROM working_groups wg JOIN working_group_members wgm ON wg.id = wgm.working_group_id WHERE   wgm.activist_id = a.id and wgm.non_member_on_mailing_list = 0), '')), (IFNULL( (SELECT   GROUP_CONCAT(DISTINCT circles.name SEPARATOR ', ') FROM circles JOIN circle_members ON circles.id = circle_members.circle_id WHERE   circle_members.activist_id = a.id), '')))) <> "","1","0")
  	as wg_or_cir_member,

    mpi,
    notes

FROM activists a

LEFT JOIN (
  SELECT
      activist_id, count(ea.event_id) as totalPoints
      FROM event_attendance ea
      JOIN events e ON e.id = ea.event_id
      WHERE
        e.date BETWEEN (NOW() - INTERVAL 30 DAY) AND NOW()
      GROUP BY activist_id
) points
  ON points.activist_id = a.id
`

const updateActivistExtraBaseQuery string = `UPDATE activists
SET

  email = :email,
  facebook = :facebook,
  location = :location,
  name = :name,
  phone = :phone,
  dob = :dob,

  activist_level = :activist_level,
  source = :source,

  connector = :connector,
  training0 = :training0,
  training1 = :training1,
  training2 = :training2,
  training3 = :training3,
  training4 = :training4,
  training5 = :training5,
  training6 = :training6,
  dev_application_date = :dev_application_date,
  dev_manager = :dev_manager,
  dev_interest = :dev_interest,
  dev_auth = :dev_auth,
  dev_email_sent = :dev_email_sent,
  dev_vetted = :dev_vetted,
  dev_interview = :dev_interview,
  dev_onboarding = :dev_onboarding,
  prospect_senior_organizer = :prospect_senior_organizer,
  so_auth = :so_auth,
  so_core = :so_core,
  so_agreement = :so_agreement,
  so_training = :so_training,
  so_quiz = :so_quiz,
  so_connector = :so_connector,
  so_onboarding = :so_onboarding,
  cm_first_email = :cm_first_email,
  cm_approval_email = :cm_approval_email,
  cm_warning_email = :cm_warning_email,
  cir_first_email = :cir_first_email,
  prospect_organizer = :prospect_organizer,
  prospect_chapter_member = :prospect_chapter_member,
  circle_agreement = :circle_agreement,
  referral_friends = :referral_friends,
  referral_apply = :referral_apply,
  referral_outlet = :referral_outlet,
  circle_interest = :circle_interest,
  interest_date = :interest_date,
  mpi = :mpi,
  notes = :notes

WHERE
  id = :id`

const DescOrder int = 2
const AscOrder int = 1

/** Type Definitions */

type Activist struct {
	Email    string         `db:"email"`
	Facebook string         `db:"facebook"`
	Hidden   bool           `db:"hidden"`
	ID       int            `db:"id"`
	Location sql.NullString `db:"location"`
	Name     string         `db:"name"`
	Phone    string         `db:"phone"`
	Birthday sql.NullString `db:"dob"`
}

type ActivistEventData struct {
	FirstEvent     mysql.NullTime `db:"first_event"`
	LastEvent      mysql.NullTime `db:"last_event"`
	LastCircle     mysql.NullTime `db:"last_circle"`
	FirstEventName string         `db:"first_event_name"`
	LastEventName  string         `db:"last_event_name"`
	TotalEvents    int            `db:"total_events"`
	TotalPoints    int            `db:"total_points"`
	Active         bool           `db:"active"`
	Status         string
}

type ActivistMembershipData struct {
	ActivistLevel string `db:"activist_level"`
	Source        string `db:"source"`
	WorkingGroups string `db:"working_group_list"`
	Circles       string `db:"circles_list"`
	WgOrCirMember bool   `db:"wg_or_cir_member"`
}

type ActivistConnectionData struct {
	Connector       string         `db:"connector"`
	Training0       sql.NullString `db:"training0"`
	Training1       sql.NullString `db:"training1"`
	Training2       sql.NullString `db:"training2"`
	Training3       sql.NullString `db:"training3"`
	Training4       sql.NullString `db:"training4"`
	Training5       sql.NullString `db:"training5"`
	Training6       sql.NullString `db:"training6"`
	ApplicationDate mysql.NullTime `db:"dev_application_date"`
	DevManager      string         `db:"dev_manager"`
	DevInterest     string         `db:"dev_interest"`
	DevAuth         sql.NullString `db:"dev_auth"`
	DevEmailSent    sql.NullString `db:"dev_email_sent"`
	DevVetted       bool           `db:"dev_vetted"`
	DevInterview    sql.NullString `db:"dev_interview"`
	DevOnboarding   bool           `db:"dev_onboarding"`

	ProspectSeniorOrganizer bool           `db:"prospect_senior_organizer"`
	SOAuth                  sql.NullString `db:"so_auth"`
	SOCore                  sql.NullString `db:"so_core"`
	SOAgreement             bool           `db:"so_agreement"`
	SOTraining              sql.NullString `db:"so_training"`
	SOQuiz                  sql.NullString `db:"so_quiz"`
	SOConnector             string         `db:"so_connector"`
	SOOnboarding            bool           `db:"so_onboarding"`

	CMFirstEmail          sql.NullString `db:"cm_first_email"`
	CMApprovalEmail       sql.NullString `db:"cm_approval_email"`
	CMWarningEmail        sql.NullString `db:"cm_warning_email"`
	CirFirstEmail         sql.NullString `db:"cir_first_email"`
	ProspectOrganizer     bool           `db:"prospect_organizer"`
	ProspectChapterMember bool           `db:"prospect_chapter_member"`
	CircleAgreement       bool           `db:"circle_agreement"`
	LastConnection        sql.NullString `db:"last_connection"`
	ReferralFriends       string         `db:"referral_friends"`
	ReferralApply         string         `db:"referral_apply"`
	ReferralOutlet        string         `db:"referral_outlet"`
	CircleInterest        bool           `db:"circle_interest"`
	InterestDate          sql.NullString `db:"interest_date"`
	MPI                   bool           `db:"mpi"`
	Notes                 sql.NullString `db:"notes"`
}

type ActivistExtra struct {
	Activist
	ActivistEventData
	ActivistMembershipData
	ActivistConnectionData
}

type ActivistJSON struct {
	Email    string `json:"email"`
	Facebook string `json:"facebook"`
	ID       int    `json:"id"`
	Location string `json:"location"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Birthday string `json:"dob"`

	FirstEvent     string `json:"first_event"`
	LastEvent      string `json:"last_event"`
	LastCircle     string `json:"last_circle"`
	FirstEventName string `json:"first_event_name"`
	LastEventName  string `json:"last_event_name"`
	TotalEvents    int    `json:"total_events"`
	TotalPoints    int    `json:"total_points"`
	Active         bool   `json:"active"`
	Status         string `json:"status"`

	ActivistLevel string `json:"activist_level"`
	Source        string `json:"source"`
	WorkingGroups string `json:"working_group_list"`
	Circles       string `json:"circles_list"`
	WgOrCirMember bool   `json:"wg_or_cir_member"`

	Connector       string `json:"connector"`
	Training0       string `json:"training0"`
	Training1       string `json:"training1"`
	Training2       string `json:"training2"`
	Training3       string `json:"training3"`
	Training4       string `json:"training4"`
	Training5       string `json:"training5"`
	Training6       string `json:"training6"`
	ApplicationDate string `json:"dev_application_date"`
	DevManager      string `json:"dev_manager"`
	DevInterest     string `json:"dev_interest"`
	DevAuth         string `json:"dev_auth"`
	DevEmailSent    string `json:"dev_email_sent"`
	DevVetted       bool   `json:"dev_vetted"`
	DevInterview    string `json:"dev_interview"`
	DevOnboarding   bool   `json:"dev_onboarding"`

	ProspectSeniorOrganizer bool   `json:"prospect_senior_organizer"`
	SOAuth                  string `json:"so_auth"`
	SOCore                  string `json:"so_core"`
	SOAgreement             bool   `json:"so_agreement"`
	SOTraining              string `json:"so_training"`
	SOQuiz                  string `json:"so_quiz"`
	SOConnector             string `json:"so_connector"`
	SOOnboarding            bool   `json:"so_onboarding"`

	CMFirstEmail          string `json:"cm_first_email"`
	CMApprovalEmail       string `json:"cm_approval_email"`
	CMWarningEmail        string `json:"cm_warning_email"`
	CirFirstEmail         string `json:"cir_first_email"`
	ProspectOrganizer     bool   `json:"prospect_organizer"`
	ProspectChapterMember bool   `json:"prospect_chapter_member"`
	CircleAgreement       bool   `json:"circle_agreement"`
	LastConnection        string `json:"last_connection"`
	ReferralFriends       string `json:"referral_friends"`
	ReferralApply         string `json:"referral_apply"`
	ReferralOutlet        string `json:"referral_outlet"`
	CircleInterest        bool   `json:"circle_interest"`
	InterestDate          string `json:"interest_date"`
	MPI                   bool   `json:"mpi"`
	Notes                 string `json:"notes"`
}

type GetActivistOptions struct {
	ID                int    `json:"id"`
	Hidden            bool   `json:"hidden"`
	Order             int    `json:"order"`
	OrderField        string `json:"order_field"`
	LastEventDateFrom string `json:"last_event_date_from"`
	LastEventDateTo   string `json:"last_event_date_to"`
}

var validOrderFields = map[string]struct{}{
	"a.name":        struct{}{},
	"last_event":    struct{}{},
	"total_points":  struct{}{},
	"interest_date": struct{}{},
}

type ActivistRangeOptionsJSON struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Order int    `json:"order"`
}

/** Functions and Methods */

func GetActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	if options.ID != 0 {
		return nil, errors.New("GetActivistsJSON: Cannot include ID in options")
	}
	return getActivistsJSON(db, options)
}

func GetActivistJSON(db *sqlx.DB, options GetActivistOptions) (ActivistJSON, error) {
	if options.ID == 0 {
		return ActivistJSON{}, errors.New("GetActivistJSON: Must include ID in options")
	}

	activists, err := getActivistsJSON(db, options)
	if err != nil {
		return ActivistJSON{}, err
	} else if len(activists) == 0 {
		return ActivistJSON{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return ActivistJSON{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func getActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	activists, err := GetActivistsExtra(db, options)
	if err != nil {
		return nil, err
	}
	return buildActivistJSONArray(activists), nil
}

func GetActivistRangeJSON(db *sqlx.DB, options ActivistRangeOptionsJSON) ([]ActivistJSON, error) {
	activists, err := getActivistRange(db, options)
	if err != nil {
		return nil, err
	}
	return buildActivistJSONArray(activists), nil
}

func buildActivistJSONArray(activists []ActivistExtra) []ActivistJSON {
	var activistsJSON []ActivistJSON

	for _, a := range activists {
		firstEvent := ""
		if a.ActivistEventData.FirstEvent.Valid {
			firstEvent = a.ActivistEventData.FirstEvent.Time.Format(EventDateLayout)
		}
		lastEvent := ""
		if a.ActivistEventData.LastEvent.Valid {
			lastEvent = a.ActivistEventData.LastEvent.Time.Format(EventDateLayout)
		}
		lastCircle := ""
		if a.ActivistEventData.LastCircle.Valid {
			lastCircle = a.ActivistEventData.LastCircle.Time.Format(EventDateLayout)
		}
		applicationDate := ""
		if a.ActivistConnectionData.ApplicationDate.Valid {
			applicationDate = a.ActivistConnectionData.ApplicationDate.Time.Format(EventDateLayout)
		}
		location := ""
		if a.Activist.Location.Valid {
			location = a.Activist.Location.String
		}
		dob := ""
		if a.Activist.Birthday.Valid {
			dob = a.Activist.Birthday.String
		}
		training0 := ""
		if a.ActivistConnectionData.Training0.Valid {
			training0 = a.ActivistConnectionData.Training0.String
		}
		training1 := ""
		if a.ActivistConnectionData.Training1.Valid {
			training1 = a.ActivistConnectionData.Training1.String
		}
		training2 := ""
		if a.ActivistConnectionData.Training2.Valid {
			training2 = a.ActivistConnectionData.Training2.String
		}
		training3 := ""
		if a.ActivistConnectionData.Training3.Valid {
			training3 = a.ActivistConnectionData.Training3.String
		}
		training4 := ""
		if a.ActivistConnectionData.Training4.Valid {
			training4 = a.ActivistConnectionData.Training4.String
		}
		training5 := ""
		if a.ActivistConnectionData.Training5.Valid {
			training5 = a.ActivistConnectionData.Training5.String
		}
		training6 := ""
		if a.ActivistConnectionData.Training6.Valid {
			training6 = a.ActivistConnectionData.Training6.String
		}
		dev_auth := ""
		if a.ActivistConnectionData.DevAuth.Valid {
			dev_auth = a.ActivistConnectionData.DevAuth.String
		}
		dev_email_sent := ""
		if a.ActivistConnectionData.DevEmailSent.Valid {
			dev_email_sent = a.ActivistConnectionData.DevEmailSent.String
		}
		dev_interview := ""
		if a.ActivistConnectionData.DevInterview.Valid {
			dev_interview = a.ActivistConnectionData.DevInterview.String
		}
		last_connection := ""
		if a.ActivistConnectionData.LastConnection.Valid {
			last_connection = a.ActivistConnectionData.LastConnection.String
		}
		cm_first_email := ""
		if a.ActivistConnectionData.CMFirstEmail.Valid {
			cm_first_email = a.ActivistConnectionData.CMFirstEmail.String
		}
		cm_approval_email := ""
		if a.ActivistConnectionData.CMApprovalEmail.Valid {
			cm_approval_email = a.ActivistConnectionData.CMApprovalEmail.String
		}
		cm_warning_email := ""
		if a.ActivistConnectionData.CMWarningEmail.Valid {
			cm_warning_email = a.ActivistConnectionData.CMWarningEmail.String
		}
		cir_first_email := ""
		if a.ActivistConnectionData.CirFirstEmail.Valid {
			cir_first_email = a.ActivistConnectionData.CirFirstEmail.String
		}
		so_auth := ""
		if a.ActivistConnectionData.SOAuth.Valid {
			so_auth = a.ActivistConnectionData.SOAuth.String
		}
		so_core := ""
		if a.ActivistConnectionData.SOCore.Valid {
			so_core = a.ActivistConnectionData.SOCore.String
		}
		so_training := ""
		if a.ActivistConnectionData.SOTraining.Valid {
			so_training = a.ActivistConnectionData.SOTraining.String
		}
		so_quiz := ""
		if a.ActivistConnectionData.SOQuiz.Valid {
			so_quiz = a.ActivistConnectionData.SOQuiz.String
		}
		interest_date := ""
		if a.ActivistConnectionData.InterestDate.Valid {
			interest_date = a.ActivistConnectionData.InterestDate.String
		}
		notes := ""
		if a.ActivistConnectionData.Notes.Valid {
			notes = a.ActivistConnectionData.Notes.String
		}

		activistsJSON = append(activistsJSON, ActivistJSON{
			Email:    a.Email,
			Facebook: a.Facebook,
			ID:       a.ID,
			Location: location,
			Name:     a.Name,
			Phone:    a.Phone,
			Birthday: dob,

			FirstEvent:     firstEvent,
			LastEvent:      lastEvent,
			LastCircle:     lastCircle,
			FirstEventName: a.FirstEventName,
			LastEventName:  a.LastEventName,
			Status:         a.Status,
			TotalEvents:    a.TotalEvents,
			TotalPoints:    a.TotalPoints,
			Active:         a.Active,

			ActivistLevel: a.ActivistLevel,
			WorkingGroups: a.WorkingGroups,
			Circles:       a.Circles,
			WgOrCirMember: a.WgOrCirMember,
			Source:        a.Source,

			Connector:       a.Connector,
			Training0:       training0,
			Training1:       training1,
			Training2:       training2,
			Training3:       training3,
			Training4:       training4,
			Training5:       training5,
			Training6:       training6,
			ApplicationDate: applicationDate,
			DevManager:      a.DevManager,
			DevInterest:     a.DevInterest,
			DevAuth:         dev_auth,
			DevEmailSent:    dev_email_sent,
			DevVetted:       a.DevVetted,
			DevInterview:    dev_interview,
			DevOnboarding:   a.DevOnboarding,

			ProspectSeniorOrganizer: a.ProspectSeniorOrganizer,
			SOAuth:                  so_auth,
			SOCore:                  so_core,
			SOAgreement:             a.SOAgreement,
			SOTraining:              so_training,
			SOQuiz:                  so_quiz,
			SOConnector:             a.SOConnector,
			SOOnboarding:            a.SOOnboarding,

			CMFirstEmail:          cm_first_email,
			CMApprovalEmail:       cm_approval_email,
			CMWarningEmail:        cm_warning_email,
			CirFirstEmail:         cir_first_email,
			ProspectOrganizer:     a.ProspectOrganizer,
			ProspectChapterMember: a.ProspectChapterMember,
			CircleAgreement:       a.CircleAgreement,
			LastConnection:        last_connection,
			ReferralFriends:       a.ReferralFriends,
			ReferralApply:         a.ReferralApply,
			ReferralOutlet:        a.ReferralOutlet,
			CircleInterest:        a.CircleInterest,
			InterestDate:          interest_date,
			MPI:                   a.MPI,
			Notes:                 notes,
		})
	}

	return activistsJSON
}

func GetActivist(db *sqlx.DB, name string) (Activist, error) {
	activists, err := getActivists(db, name)
	if err != nil {
		return Activist{}, err
	} else if len(activists) == 0 {
		return Activist{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return Activist{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func GetActivists(db *sqlx.DB) ([]Activist, error) {
	return getActivists(db, "")
}

func getActivists(db *sqlx.DB, name string) ([]Activist, error) {
	var queryArgs []interface{}
	query := selectActivistBaseQuery

	if name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var activists []Activist
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists for %s", name)
	}

	return activists, nil
}

func GetChapterMembers(db *sqlx.DB) ([]Activist, error) {
	query := `
SELECT
  id,
  name,
  email
FROM activists
WHERE hidden = 0 AND activist_level IN('Organizer', 'Senior Organizer', 'Chapter Member')
`

	var activists []Activist
	err := db.Select(&activists, query)
	if err != nil {
		return []Activist{}, errors.Wrapf(err, "GetChapterMembers: Failed retrieving activists for levels Organizer, Senior Organizer, and Chapter Member")
	}

	return activists, nil
}

func GetActivistsExtra(db *sqlx.DB, options GetActivistOptions) ([]ActivistExtra, error) {
	// Redundant options validation
	var err error
	options, err = validateGetActivistOptions(options)
	if err != nil {
		return nil, err
	}

	query := selectActivistExtraBaseQuery

	var queryArgs []interface{}

	if options.ID != 0 {
		// retrieve specific activist rather than all activists
		query += " WHERE a.id = ? "
		queryArgs = append(queryArgs, options.ID)
	} else {
		// Only check filter by hidden if the activistID isn't
		// supplied.
		if options.Hidden == true {
			query += " WHERE a.hidden = true "
		} else {
			query += " WHERE a.hidden = false "
		}
	}

	havingClause := []string{}
	if options.LastEventDateFrom != "" {
		havingClause = append(havingClause, "last_event >= ?")
		queryArgs = append(queryArgs, options.LastEventDateFrom)
	}
	if options.LastEventDateTo != "" {
		havingClause = append(havingClause, "last_event <= ?")
		queryArgs = append(queryArgs, options.LastEventDateTo)
	}

	if len(havingClause) != 0 {
		query += " HAVING " + strings.Join(havingClause, " AND ")
	}

	orderField := options.OrderField
	// Default to a.name if orderField isn't specified
	if orderField == "" {
		orderField = "a.name"
	}
	// We already check that options.OrderField is valid in
	// CleanGetActivistOptions, but we check it again here again
	// to be paranoid b/c this is a sql injection if we don't
	// check it.
	if _, ok := validOrderFields[orderField]; !ok {
		return nil, errors.New("Invalid OrderField")
	}

	query += " ORDER BY " + options.OrderField
	if options.Order == DescOrder {
		query += " desc "
	}

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists extra for uid %d", options.ID)
	}

	for i := 0; i < len(activists); i++ {
		a := activists[i]
		activists[i].Status = getStatus(a.FirstEvent, a.LastEvent, a.TotalEvents)
	}

	return activists, nil
}

// TODO Make sure you only fetch non-hidden members
// THAT is not currently the case
func getActivistRange(db *sqlx.DB, options ActivistRangeOptionsJSON) ([]ActivistExtra, error) {
	// Redundant options validation
	var err error
	options, err = validateActivistRangeOptionsJSON(options)
	if err != nil {
		return nil, err
	}

	query := selectActivistExtraBaseQuery
	name := options.Name
	order := options.Order
	limit := options.Limit
	var queryArgs []interface{}

	query += " WHERE a.hidden = false "

	if name != "" {
		if order == DescOrder {
			query += " AND a.name < ? "
		} else {
			query += " AND a.name > ? "
		}
		queryArgs = append(queryArgs, name)
	}

	query += " GROUP BY a.name "

	query += " ORDER BY a.name "
	if order == DescOrder {
		query += "desc "
	}

	if limit > 0 {
		query += " LIMIT ? "
		queryArgs = append(queryArgs, limit)
	}

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve %d users before/after %s", limit, name)
	}

	for i := 0; i < len(activists); i++ {
		a := activists[i]
		activists[i].Status = getStatus(a.FirstEvent, a.LastEvent, a.TotalEvents)
	}

	return activists, nil
}

func (a Activist) GetActivistEventData(db *sqlx.DB) (ActivistEventData, error) {
	query := `
SELECT
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(*) as total_events
FROM events e
JOIN event_attendance
  ON event_attendance.event_id = e.id
WHERE
  event_attendance.activist_id = ?
`
	var data ActivistEventData
	if err := db.Get(&data, query, a.ID); err != nil {
		return ActivistEventData{}, errors.Wrap(err, "failed to get activist event data")
	}
	return data, nil
}

func GetOrCreateActivist(db *sqlx.DB, name string) (Activist, error) {
	activist, err := GetActivist(db, name)
	if err == nil {
		// We got a valid activist, return them.
		return activist, nil
	}

	// There was an error, so try inserting the activist first.
	// Wrap in transaction to avoid issue where a new activist
	// is inserted successfully, but we are unable to retrieve
	// the new activist, which will leave database in inconsistent state

	tx, err := db.Beginx()
	if err != nil {
		return Activist{}, errors.Wrap(err, "Failed to create transaction")
	}

	_, err = tx.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to insert activist %s", name)
	}

	query := selectActivistBaseQuery + " WHERE name = ? "

	var newActivist Activist
	err = tx.Get(&newActivist, query, name)

	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to get new activist %s", name)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to commit activist %s", name)
	}

	return newActivist, nil
}

func CreateActivist(db *sqlx.DB, activist ActivistExtra) (int, error) {
	if activist.ID != 0 {
		return 0, errors.New("Activist ID must be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	result, err := db.NamedExec(`
INSERT INTO activists (

  email,
  facebook,
  location,
  name,
  phone,
  dob,

  activist_level,
  source,

  connector,
  training0,
  training1,
  training2,
  training3,
  training4,
  training5,
  training6,
  dev_manager,
  dev_interest,
  dev_auth,
  dev_email_sent,
  dev_vetted,
  dev_interview,
  dev_onboarding,
  prospect_senior_organizer,
  so_auth,
  so_core,
  so_agreement,
  so_training,
  so_quiz,
  so_connector,
  so_onboarding,
  cm_first_email,
  cm_approval_email,
  cm_warning_email,
  cir_first_email,
  prospect_organizer,
  prospect_chapter_member,
  circle_agreement,
  last_connection,
  referral_friends,
  referral_apply,
  referral_outlet,
  circle_interest,
  interest_date,
  mpi,
  notes

) VALUES (

  :email,
  :facebook,
  :location,
  :name,
  :phone,
  :dob,

  :activist_level,
  :source,

  :connector,
  :training0,
  :training1,
  :training2,
  :training3,
  :training4,
  :training5,
  :training6,
  :dev_manager,
  :dev_interest,
  :dev_auth,
  :dev_email_sent,
  :dev_vetted,
  :dev_interview,
  :dev_onboarding,
  :prospect_senior_organizer,
  :so_auth,
  :so_core,
  :so_agreement,
  :so_training,
  :so_quiz,
  :so_connector,
  :so_onboarding,
  :cm_first_email,
  :cm_approval_email,
  :cm_warning_email,
  :cir_first_email,
  :prospect_organizer,
  :prospect_chapter_member,
  :circle_agreement,
  :last_connection,
  :referral_friends,
  :referral_apply,
  :referral_outlet,
  :circle_interest,
  :interest_date,
  :mpi,
  :notes

)`, activist)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not create activist: %s", activist.Name)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "Could not get LastInsertId for %s", activist.Name)
	}
	return int(id), nil
}

func UpdateActivistData(db *sqlx.DB, activist ActivistExtra) (int, error) {
	if activist.ID == 0 {
		return 0, errors.New("activist ID cannot be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	_, err := db.NamedExec(`UPDATE activists
SET

  email = :email,
  facebook = :facebook,
  location = :location,
  name = :name,
  phone = :phone,
  dob = :dob,

  activist_level = :activist_level,
  source = :source,

  connector = :connector,
  training0 = :training0,
  training1 = :training1,
  training2 = :training2,
  training3 = :training3,
  training4 = :training4,
  training5 = :training5,
  training6 = :training6,
  dev_manager = :dev_manager,
  dev_interest = :dev_interest,
  dev_auth = :dev_auth,
  dev_email_sent = :dev_email_sent,
  dev_vetted = :dev_vetted,
  dev_interview = :dev_interview,
  dev_onboarding = :dev_onboarding,
  prospect_senior_organizer = :prospect_senior_organizer,
  so_auth = :so_auth,
  so_core = :so_core,
  so_agreement = :so_agreement,
  so_training = :so_training,
  so_quiz = :so_quiz,
  so_connector = :so_connector,
  so_onboarding = :so_onboarding,
  cm_first_email = :cm_first_email,
  cm_approval_email = :cm_approval_email,
  cm_warning_email = :cm_warning_email,
  cir_first_email = :cir_first_email,
  prospect_organizer = :prospect_organizer,
  prospect_chapter_member = :prospect_chapter_member,
  circle_agreement = :circle_agreement,
  referral_friends = :referral_friends,
  referral_apply = :referral_apply,
  referral_outlet = :referral_outlet,
  circle_interest = :circle_interest,
  interest_date = :interest_date,
  mpi = :mpi,
  notes = :notes

WHERE
  id = :id`, activist)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update activist data")
	}
	return activist.ID, nil
}

func HideActivist(db *sqlx.DB, activistID int) error {
	if activistID == 0 {
		return errors.New("HideActivist: activistID cannot be 0")
	}
	var activistCount int
	err := db.Get(&activistCount, `SELECT count(*) FROM activists WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrap(err, "failed to get activist count")
	}
	if activistCount == 0 {
		return errors.Errorf("Activist with id %d does not exist", activistID)
	}

	_, err = db.Exec(`UPDATE activists SET hidden = true WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update activist %d", activistID)
	}
	return nil
}

// Merge activistID into targetActivistID.
//  - The original activist is hidden
//  - All of the original activist's event attendance is updated to be the target activist.
func MergeActivist(db *sqlx.DB, originalActivistID, targetActivistID int) error {
	if originalActivistID == 0 {
		return errors.New("originalActivistID cannot be 0")
	}
	if targetActivistID == 0 {
		return errors.New("targetActivistID cannot be 0")
	}
	if originalActivistID == targetActivistID {
		return errors.New("originalActivist and targetActivist cannot be the same")
	}

	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "could not create transaction")
	}

	_, err = tx.Exec(`UPDATE activists SET hidden = true, name = concat(name,' ', id) WHERE id = ?`, originalActivistID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to hide original activist %d", originalActivistID)
	}

	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Merge Activist data details
	err = updateMergedActivistDataDetails(tx, originalActivistID, targetActivistID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrapf(err,
			"failed to commit merge activist transaction. original activist id: %d, target activist id: %d",
			originalActivistID, targetActivistID)
	}

	return nil
}

func updateMergedActivistData(tx *sqlx.Tx, originalActivistID int, targetActivistID int, originalActivistOnly bool) error {
	baseQuery := `
SELECT event_id
FROM event_attendance ea
WHERE
  activist_id = ?
  AND `

	subquery := `
EXISTS(
  SELECT ea2.event_id
  FROM event_attendance ea2
  WHERE ea2.activist_id = ?
    AND ea2.event_id = ea.event_id)`

	var eventQuery string
	if originalActivistOnly {
		eventQuery = baseQuery + " NOT " + subquery
	} else {
		eventQuery = baseQuery + subquery
	}

	var eventIDs []int
	err := tx.Select(&eventIDs, eventQuery, originalActivistID, targetActivistID)
	if err != nil {
		return errors.Wrapf(err,
			"failed to get original activist's events: %d, originalActivistOnly: %v",
			originalActivistID, originalActivistOnly)
	}

	// There's nothing to do if there are no events.
	if len(eventIDs) == 0 {
		return nil
	}

	var eaQuery string
	var eaArgs []interface{}
	if originalActivistOnly {
		eaQuery, eaArgs, err = sqlx.In(`
UPDATE event_attendance
SET activist_id = ?
WHERE
  activist_id = ?
  AND event_id IN (?)`,
			targetActivistID, originalActivistID, eventIDs)
	} else {
		eaQuery, eaArgs, err = sqlx.In(`
DELETE FROM event_attendance
WHERE
  activist_id = ?
  AND event_id IN (?)`, originalActivistID, eventIDs)
	}
	if err != nil {
		return errors.Wrapf(err, "could not create sqlx.IN query. originalActivistOnly: %v",
			originalActivistOnly)
	}
	eaQuery = tx.Rebind(eaQuery)
	_, err = tx.Exec(eaQuery, eaArgs...)
	if err != nil {
		return errors.Wrapf(err, "could not update event attendance for activist: %d",
			originalActivistID)
	}
	err = insertMergedActivistAttendance(tx, originalActivistID, targetActivistID, eventIDs, originalActivistOnly)
	if err != nil {
		return err
	}
	return nil
}

func insertMergedActivistAttendance(tx *sqlx.Tx, originalActivistID int, targetActivistID int, eventIDs []int, replacedWithTargetActivist bool) error {
	if len(eventIDs) == 0 {
		return nil
	}

	query := `INSERT INTO merged_activist_attendance (original_activist_id, target_activist_id, event_id, replaced_with_target_activist) VALUES`

	var queryValues []string
	var queryArgs []interface{}
	for _, eventID := range eventIDs {
		queryValues = append(queryValues, " (?, ?, ?, ?) ")
		queryArgs = append(queryArgs, originalActivistID, targetActivistID, eventID, replacedWithTargetActivist)
	}
	query += strings.Join(queryValues, ",")
	_, err := tx.Exec(query, queryArgs...)

	return errors.Wrapf(err, "could not insert merged_activist_attendance for originalActivistID: %d, targetActivistID: %d",
		originalActivistID, targetActivistID)
}

func getMergeActivistWinner(original ActivistExtra, target ActivistExtra) ActivistExtra {
	levels := map[string]int{
		"Supporter":        0,
		"Circle Member":    1,
		"Chapter Member":   2,
		"Organizer":        3,
		"Senior Organizer": 4,
	}

	// Check boolean values

	target.ProspectOrganizer = boolMerge(original.ProspectOrganizer, target.ProspectOrganizer)
	target.ProspectChapterMember = boolMerge(original.ProspectChapterMember, target.ProspectChapterMember)
	target.CircleAgreement = boolMerge(original.CircleAgreement, target.CircleAgreement)
	target.DevVetted = boolMerge(original.DevVetted, target.DevVetted)
	target.DevOnboarding = boolMerge(original.DevOnboarding, target.DevOnboarding)
	target.ProspectSeniorOrganizer = boolMerge(original.ProspectSeniorOrganizer, target.ProspectSeniorOrganizer)
	target.SOAgreement = boolMerge(original.SOAgreement, target.SOAgreement)
	target.SOOnboarding = boolMerge(original.SOOnboarding, target.SOOnboarding)
	target.CircleInterest = boolMerge(original.CircleInterest, target.CircleInterest)
	target.MPI = boolMerge(original.MPI, target.MPI)

	// Check string fields for empty values

	target.Email = stringMerge(original.Email, target.Email)
	target.Phone = stringMerge(original.Phone, target.Phone)
	target.Birthday = stringMergeSqlNullString(original.Birthday, target.Birthday)
	target.Location = stringMergeSqlNullString(original.Location, target.Location)
	target.Facebook = stringMerge(original.Facebook, target.Facebook)
	target.Connector = stringMerge(original.Connector, target.Connector)
	target.Source = stringMerge(original.Source, target.Source)
	target.Training0 = stringMergeSqlNullString(original.Training0, target.Training0)
	target.Training1 = stringMergeSqlNullString(original.Training1, target.Training1)
	target.Training2 = stringMergeSqlNullString(original.Training2, target.Training2)
	target.Training3 = stringMergeSqlNullString(original.Training3, target.Training3)
	target.Training4 = stringMergeSqlNullString(original.Training4, target.Training4)
	target.Training5 = stringMergeSqlNullString(original.Training5, target.Training5)
	target.Training6 = stringMergeSqlNullString(original.Training6, target.Training6)
	target.DevManager = stringMerge(original.DevManager, target.DevManager)
	target.DevInterest = stringMerge(original.DevInterest, target.DevInterest)
	target.DevAuth = stringMergeSqlNullString(original.DevAuth, target.DevAuth)
	target.DevEmailSent = stringMergeSqlNullString(original.DevEmailSent, target.DevEmailSent)
	target.DevInterview = stringMergeSqlNullString(original.DevInterview, target.DevInterview)
	target.ApplicationDate = stringMergeSqlNullTime(original.ApplicationDate, target.ApplicationDate)
	target.CMFirstEmail = stringMergeSqlNullString(original.CMFirstEmail, target.CMFirstEmail)
	target.CMApprovalEmail = stringMergeSqlNullString(original.CMApprovalEmail, target.CMApprovalEmail)
	target.CMWarningEmail = stringMergeSqlNullString(original.CMWarningEmail, target.CMWarningEmail)
	target.CirFirstEmail = stringMergeSqlNullString(original.CirFirstEmail, target.CirFirstEmail)
	target.SOAuth = stringMergeSqlNullString(original.SOAuth, target.SOAuth)
	target.SOCore = stringMergeSqlNullString(original.SOCore, target.SOCore)
	target.SOTraining = stringMergeSqlNullString(original.SOTraining, target.SOTraining)
	target.SOQuiz = stringMergeSqlNullString(original.SOQuiz, target.SOQuiz)
	target.SOConnector = stringMerge(original.SOConnector, target.SOConnector)
	target.ReferralFriends = stringMerge(original.ReferralFriends, target.ReferralFriends)
	target.ReferralApply = stringMerge(original.ReferralApply, target.ReferralApply)
	target.ReferralOutlet = stringMerge(original.ReferralOutlet, target.ReferralOutlet)
	target.InterestDate = stringMergeSqlNullString(original.InterestDate, target.InterestDate)
	target.Notes = stringMergeSqlNullString(original.Notes, target.Notes)

	// Check Activist Levels
	if len(original.ActivistLevel) != 0 && len(target.ActivistLevel) != 0 {
		if levels[original.ActivistLevel] > levels[target.ActivistLevel] {
			target.ActivistLevel = original.ActivistLevel
		}
	} else {
		if len(original.ActivistLevel) != 0 {
			target.ActivistLevel = original.ActivistLevel
		}
	}

	return target
}

func boolMerge(original bool, target bool) bool {
	return target || original
}

func stringMerge(original string, target string) string {
	if len(target) == 0 && len(original) != 0 {
		return original
	}

	return target
}

func stringMergeSqlNullString(original sql.NullString, target sql.NullString) sql.NullString {
	if !target.Valid && original.Valid {
		return original
	}

	return target
}

func stringMergeSqlNullTime(original mysql.NullTime, target mysql.NullTime) mysql.NullTime {
	if !target.Valid && original.Valid {
		return original
	}

	return target
}

func updateMergedActivistDataDetails(tx *sqlx.Tx, originalActivistID int, targetActivistID int) error {
	// Merge details of original activist into target activist
	// Favor booleans that are set to TRUE, and pull in missing data from original activist to target; when both
	// activists have data for the same field, we should use the target activist's data.

	query := selectActivistExtraBaseQuery + " WHERE id = ?"

	var originalActivist = new(ActivistExtra)
	err := tx.Get(originalActivist, query, originalActivistID)
	if err != nil || originalActivist == nil {
		return errors.Wrapf(err, "failed to get activist with id %d", originalActivistID)
	}

	var targetActivist = new(ActivistExtra)
	err = tx.Get(targetActivist, query, targetActivistID)
	if err != nil || (targetActivist == nil) {
		return errors.Wrapf(err, "failed to get activist with id %d", targetActivistID)
	}

	mergedActivist := getMergeActivistWinner(*originalActivist, *targetActivist)

	_, err = tx.NamedExec(updateActivistExtraBaseQuery, mergedActivist)

	if err != nil {
		return errors.Wrapf(err, "failed to update activist with id %d", targetActivistID)
	}

	return nil
}

func GetAutocompleteNames(db *sqlx.DB) []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	// Order the activists by the last even they've been to.
	err := db.Select(&names, `
SELECT a.name FROM activists a
LEFT OUTER JOIN event_attendance ea ON a.id = ea.activist_id
LEFT OUTER JOIN events e ON e.id = ea.event_id
WHERE a.hidden = 0
GROUP BY a.name
ORDER BY MAX(e.date) DESC`)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	ret := []string{}
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

type ActivistBasicInfoJSON struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type ActivistBasicInfo struct {
	Name  string `db:"name"`
	Email string `db:"email"`
	Phone string `db:"phone"`
}

func (activist *ActivistBasicInfo) ToJSON() ActivistBasicInfoJSON {
	return ActivistBasicInfoJSON{
		Name:  activist.Name,
		Email: activist.Email,
		Phone: activist.Phone,
	}
}

func GetActivistListBasicJSON(db *sqlx.DB) []ActivistBasicInfoJSON {
	activists := []ActivistBasicInfo{}

	// Order the activists by the last even they've been to.
	err := db.Select(&activists, `
SELECT a.name, a.email, a.phone FROM activists a
LEFT OUTER JOIN event_attendance ea ON a.id = ea.activist_id
LEFT OUTER JOIN events e ON e.id = ea.event_id
WHERE a.hidden = 0
GROUP BY a.name
ORDER BY MAX(e.date) DESC`)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	activistsJSON := make([]ActivistBasicInfoJSON, 0, len(activists))

	for _, activist := range activists {
		activistsJSON = append(activistsJSON, activist.ToJSON())
	}

	return activistsJSON
}

func CleanActivistData(body io.Reader) (ActivistExtra, error) {
	var activistJSON ActivistJSON
	err := json.NewDecoder(body).Decode(&activistJSON)
	if err != nil {
		return ActivistExtra{}, err
	}

	// Check if name field contains dangerous input
	if err := checkForDangerousChars(activistJSON.Name); err != nil {
		return ActivistExtra{}, err
	}

	validLoc := true
	if activistJSON.Location == "" {
		// No location specified so insert null value into database
		validLoc = false
	}
	validBirthday := true
	if activistJSON.Birthday == "" {
		// No location specified so insert null value into database
		validBirthday = false
	}
	validTraining0 := true
	if activistJSON.Training0 == "" {
		// Not specified so insert null value into database
		validTraining0 = false
	}
	validTraining1 := true
	if activistJSON.Training1 == "" {
		// Not specified so insert null value into database
		validTraining1 = false
	}
	validTraining2 := true
	if activistJSON.Training2 == "" {
		// Not specified so insert null value into database
		validTraining2 = false
	}
	validTraining3 := true
	if activistJSON.Training3 == "" {
		// Not specified so insert null value into database
		validTraining3 = false
	}
	validTraining4 := true
	if activistJSON.Training4 == "" {
		// Not specified so insert null value into database
		validTraining4 = false
	}
	validTraining5 := true
	if activistJSON.Training5 == "" {
		// Not specified so insert null value into database
		validTraining5 = false
	}
	validTraining6 := true
	if activistJSON.Training6 == "" {
		// Not specified so insert null value into database
		validTraining6 = false
	}
	validDevAuth := true
	if activistJSON.DevAuth == "" {
		// Not specified so insert null value into database
		validDevAuth = false
	}
	validDevEmailSent := true
	if activistJSON.DevEmailSent == "" {
		// Not specified so insert null value into database
		validDevEmailSent = false
	}
	validDevInterview := true
	if activistJSON.DevInterview == "" {
		// Not specified so insert null value into database
		validDevInterview = false
	}
	validCMFirstEmail := true
	if activistJSON.CMFirstEmail == "" {
		// Not specified so insert null value into database
		validCMFirstEmail = false
	}
	validCMApprovalEmail := true
	if activistJSON.CMApprovalEmail == "" {
		// Not specified so insert null value into database
		validCMApprovalEmail = false
	}
	validCMWarningEmail := true
	if activistJSON.CMWarningEmail == "" {
		// Not specified so insert null value into database
		validCMWarningEmail = false
	}
	validCirFirstEmail := true
	if activistJSON.CirFirstEmail == "" {
		// Not specified so insert null value into database
		validCirFirstEmail = false
	}
	validSOAuth := true
	if activistJSON.SOAuth == "" {
		// Not specified so insert null value into database
		validSOAuth = false
	}
	validSOCore := true
	if activistJSON.SOCore == "" {
		// Not specified so insert null value into database
		validSOCore = false
	}
	validSOTraining := true
	if activistJSON.SOTraining == "" {
		// Not specified so insert null value into database
		validSOTraining = false
	}
	validSOQuiz := true
	if activistJSON.SOQuiz == "" {
		// Not specified so insert null value into database
		validSOQuiz = false
	}
	validInterestDate := true
	if activistJSON.InterestDate == "" {
		// Not specified so insert null value into database
		validInterestDate = false
	}
	validNotes := true
	if activistJSON.Notes == "" {
		// Not specified so insert null value into database
		validNotes = false
	}

	activistExtra := ActivistExtra{
		Activist: Activist{
			Email:    strings.TrimSpace(activistJSON.Email),
			Facebook: strings.TrimSpace(activistJSON.Facebook),
			ID:       activistJSON.ID,
			Location: sql.NullString{String: strings.TrimSpace(activistJSON.Location), Valid: validLoc},
			Name:     strings.TrimSpace(activistJSON.Name),
			Phone:    strings.TrimSpace(activistJSON.Phone),
			Birthday: sql.NullString{String: strings.TrimSpace(activistJSON.Birthday), Valid: validBirthday},
		},
		ActivistMembershipData: ActivistMembershipData{
			ActivistLevel: strings.TrimSpace(activistJSON.ActivistLevel),
			Source:        strings.TrimSpace(activistJSON.Source),
		},
		ActivistConnectionData: ActivistConnectionData{
			Connector:     strings.TrimSpace(activistJSON.Connector),
			Training0:     sql.NullString{String: strings.TrimSpace(activistJSON.Training0), Valid: validTraining0},
			Training1:     sql.NullString{String: strings.TrimSpace(activistJSON.Training1), Valid: validTraining1},
			Training2:     sql.NullString{String: strings.TrimSpace(activistJSON.Training2), Valid: validTraining2},
			Training3:     sql.NullString{String: strings.TrimSpace(activistJSON.Training3), Valid: validTraining3},
			Training4:     sql.NullString{String: strings.TrimSpace(activistJSON.Training4), Valid: validTraining4},
			Training5:     sql.NullString{String: strings.TrimSpace(activistJSON.Training5), Valid: validTraining5},
			Training6:     sql.NullString{String: strings.TrimSpace(activistJSON.Training6), Valid: validTraining6},
			DevManager:    strings.TrimSpace(activistJSON.DevManager),
			DevInterest:   strings.TrimSpace(activistJSON.DevInterest),
			DevAuth:       sql.NullString{String: strings.TrimSpace(activistJSON.DevAuth), Valid: validDevAuth},
			DevEmailSent:  sql.NullString{String: strings.TrimSpace(activistJSON.DevEmailSent), Valid: validDevEmailSent},
			DevVetted:     activistJSON.DevVetted,
			DevInterview:  sql.NullString{String: strings.TrimSpace(activistJSON.DevInterview), Valid: validDevInterview},
			DevOnboarding: activistJSON.DevOnboarding,

			ProspectSeniorOrganizer: activistJSON.ProspectSeniorOrganizer,
			SOAuth:                  sql.NullString{String: strings.TrimSpace(activistJSON.SOAuth), Valid: validSOAuth},
			SOCore:                  sql.NullString{String: strings.TrimSpace(activistJSON.SOCore), Valid: validSOCore},
			SOAgreement:             activistJSON.SOAgreement,
			SOTraining:              sql.NullString{String: strings.TrimSpace(activistJSON.SOTraining), Valid: validSOTraining},
			SOQuiz:                  sql.NullString{String: strings.TrimSpace(activistJSON.SOQuiz), Valid: validSOQuiz},
			SOConnector:             strings.TrimSpace(activistJSON.SOConnector),
			SOOnboarding:            activistJSON.SOOnboarding,

			CMFirstEmail:          sql.NullString{String: strings.TrimSpace(activistJSON.CMFirstEmail), Valid: validCMFirstEmail},
			CMApprovalEmail:       sql.NullString{String: strings.TrimSpace(activistJSON.CMApprovalEmail), Valid: validCMApprovalEmail},
			CMWarningEmail:        sql.NullString{String: strings.TrimSpace(activistJSON.CMWarningEmail), Valid: validCMWarningEmail},
			CirFirstEmail:         sql.NullString{String: strings.TrimSpace(activistJSON.CirFirstEmail), Valid: validCirFirstEmail},
			ProspectOrganizer:     activistJSON.ProspectOrganizer,
			ProspectChapterMember: activistJSON.ProspectChapterMember,
			CircleAgreement:       activistJSON.CircleAgreement,
			ReferralFriends:       strings.TrimSpace(activistJSON.ReferralFriends),
			ReferralApply:         strings.TrimSpace(activistJSON.ReferralApply),
			ReferralOutlet:        strings.TrimSpace(activistJSON.ReferralOutlet),
			CircleInterest:        activistJSON.CircleInterest,
			InterestDate:          sql.NullString{String: strings.TrimSpace(activistJSON.InterestDate), Valid: validInterestDate},
			MPI:                   activistJSON.MPI,
			Notes:                 sql.NullString{String: strings.TrimSpace(activistJSON.Notes), Valid: validNotes},
		},
	}

	if err := validateActivist(activistExtra); err != nil {
		return ActivistExtra{}, err
	}

	return activistExtra, nil

}

var validActivistLevels = map[string]struct{}{
	"Supporter":        struct{}{},
	"Circle Member":    struct{}{},
	"Chapter Member":   struct{}{},
	"Organizer":        struct{}{},
	"Senior Organizer": struct{}{},
}

func validateActivist(a ActivistExtra) error {
	if _, ok := validActivistLevels[a.ActivistLevel]; !ok {
		return errors.New("ActivistLevel is invalid.")
	}
	return nil
}

func validateActivistRangeOptionsJSON(a ActivistRangeOptionsJSON) (ActivistRangeOptionsJSON, error) {
	// Set defaults
	if a.Order == 0 {
		a.Order = AscOrder
	}

	// Check that order matches one of the defined order constants
	if a.Order != DescOrder && a.Order != AscOrder {
		return ActivistRangeOptionsJSON{}, errors.New("User Range order must be ascending or descending")
	}
	return a, nil
}

func GetActivistRangeOptions(body io.Reader) (ActivistRangeOptionsJSON, error) {
	var options ActivistRangeOptionsJSON
	err := json.NewDecoder(body).Decode(&options)
	if err != nil {
		return ActivistRangeOptionsJSON{}, err
	}
	options, err = validateActivistRangeOptionsJSON(options)
	if err != nil {
		return ActivistRangeOptionsJSON{}, err
	}
	return options, nil
}

func validateGetActivistOptions(a GetActivistOptions) (GetActivistOptions, error) {
	// Set defaults
	if a.Order == 0 {
		a.Order = DescOrder
	}
	if a.OrderField == "" {
		a.OrderField = "a.name"
	}

	// Check that order matches one of the defined order constants
	if a.Order != DescOrder && a.Order != AscOrder {
		return GetActivistOptions{}, errors.New("User Range order must be ascending or descending")
	}
	if _, ok := validOrderFields[a.OrderField]; !ok {
		return GetActivistOptions{}, errors.New("OrderField is not valid")
	}
	return a, nil
}

func CleanGetActivistOptions(body io.Reader) (GetActivistOptions, error) {
	var getActivistOptions GetActivistOptions
	err := json.NewDecoder(body).Decode(&getActivistOptions)
	if err != nil {
		return GetActivistOptions{}, err
	}
	getActivistOptions, err = validateGetActivistOptions(getActivistOptions)
	if err != nil {
		return GetActivistOptions{}, err
	}
	return getActivistOptions, nil
}

// Returns one of the following statuses:
//  - Current
//  - New
//  - Former
//  - No attendance
// Must be kept in sync with the list in frontend/ActivistList.vue
func getStatus(firstEvent mysql.NullTime, lastEvent mysql.NullTime, totalEvents int) string {
	if !firstEvent.Valid || !lastEvent.Valid {
		return "No attendance"
	}

	if time.Since(lastEvent.Time) > Duration60Days {
		return "Former"
	}
	if time.Since(firstEvent.Time) < Duration90Days && totalEvents < 5 {
		return "New"
	}
	return "Current"
}
