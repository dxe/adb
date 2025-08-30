package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dxe/adb/mailing_list_signup"

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
  phone,
  pronouns,
  lat,
  lng,
  chapter_id
FROM activists
`

// Query perf tips:
//   - Test changes to this query against production data to see
//     how they effect performance
//   - It seems like it's usually faster to use subqueries in the top
//     part of the SELECT expression vs joining on a table.
const selectActivistExtraBaseQuery string = `
SELECT

  lower(email) as email,
  email_updated,
  facebook,
  a.id,
  a.chapter_id,
  mpi,
  a.notes,
  vision_wall,
  mpp_requirements,
  voting_agreement,
  street_address,
  city,
  state,
  location,
  address_updated,
  location_updated,
  lat,
  lng,
  a.name,
  preferred_name,
  phone,
  phone_updated,
  pronouns,
  language,
  accessibility,
  dob,

  activist_level,
  source,
  hiatus,

  connector,
  training0,
  training1,
  training4,
  training5,
  training6,
  consent_quiz,
  training_protest,
  dev_application_date,
  dev_application_type,
  dev_quiz,
  dev_interest,

  cm_first_email,
  cm_approval_email,
  prospect_organizer,
  prospect_chapter_member,
  referral_friends,
  referral_apply,
  referral_outlet,
  interest_date,

  discord_id,

  @first_event := (
      SELECT min(e.date) AS min_date
      FROM event_attendance ea
      JOIN activists inner_a ON inner_a.id = ea.activist_id
      JOIN events e ON e.id = ea.event_id
      WHERE inner_a.id = a.id
  ) AS first_event,

  @last_event := (
    SELECT max(e.date)
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id
  ) AS last_event,

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
    
  @last_action := (
    SELECT max(e.date) AS max_date
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id and event_type in ("Action", "Outreach", "Frontline Surveillance", "Campaign Action", "Animal Care")
  ) AS last_action,
    
  IFNULL(
    TIMESTAMPDIFF(MONTH, DATE_FORMAT(@last_action, '%Y-%m-01'), NOW()),
    9999
  ) AS months_since_last_action,

  (SELECT COUNT(DISTINCT ea.event_id)
    FROM event_attendance ea
    WHERE
      ea.activist_id = a.id) as total_events,

  IFNULL(totalPoints, 0) as total_points,
  IF(@last_event >= (now() - interval 30 day), 1, 0) as active,

    IFNULL((
    SELECT max(e.date) AS max_date
    FROM event_attendance ea
    JOIN activists inner_a ON inner_a.id = ea.activist_id
    JOIN events e ON e.id = ea.event_id
    WHERE inner_a.id = a.id and e.event_type = "Connection"
  	),"") AS last_connection,

	IFNULL((
      SELECT GROUP_CONCAT(c.name)
      FROM circle_members cm
      JOIN circles c ON cm.circle_id = c.id
      WHERE cm.activist_id = a.id and c.type = 2
  	),"") AS geo_circles,

	IFNULL(@assigned_to := (
		SELECT adb_users.name
		FROM adb_users
		WHERE adb_users.id = a.assigned_to
	), "") AS assigned_to_name,

	DATE_FORMAT(followup_date, "%Y-%m-%d") as followup_date,

	@total_interactions := (
		SELECT count(id)
		FROM interactions
		WHERE interactions.activist_id = a.id
  	) AS total_interactions,

	IFNULL(@last_interaction_date := (
		SELECT DATE_FORMAT(max(interactions.timestamp), "%Y-%m-%d")
		FROM interactions
		WHERE interactions.activist_id = a.id
	), "") AS last_interaction_date

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

LEFT JOIN (

  select
    id as activist_id_mpp,
      IF(is_protest = 1 and is_community = 1, 'Fulfilling requirements',IF(is_protest = 1 and is_community = 0, 'Missing Community event',IF(is_protest = 0 and is_community = 1, 'Missing DA event','Missing Community & DA events'))) as MPP_Requirements
  from activists

  left join (
    select
        activist_id,
        max((CASE WHEN ((e.event_type = 'action') OR (e.event_type = 'outreach') OR (e.event_type = 'frontline surveillance') OR (e.event_type = 'animal care') OR (e.event_type = 'campaign action')) THEN '1' ELSE '0' END)) AS is_protest,
        max((CASE WHEN ((e.event_type = 'community') OR (e.event_type = 'training') OR (e.event_type = 'circle')) THEN '1' ELSE '0' END)) AS is_community  
    from
        event_attendance ea
    join events e on e.id = ea.event_id
    where
        YEAR(e.date) = YEAR(now()) AND (MONTH(e.date) = MONTH(now()))
    group by ea.activist_id   
  ) currentMonth on currentMonth.activist_id = activists.id
) mpp_requirements on mpp_requirements.activist_id_mpp = a.id
`

const insertActivistQuery string = `INSERT INTO activists SET ` +
	`chapter_id = :chapter_id,` +
	ActivistDataFieldAssignments

const insertActivistWithTimestampsQuery string = insertActivistQuery + `,
  email_updated = :email_updated,
  phone_updated = :phone_updated,
  address_updated = :address_updated,
  location_updated = :location_updated`

const updateActivistQuery string = `UPDATE activists
SET ` +
	// Modified timestamp field assignments
	//
	// These fields must be set before (above) the fields they track, otherwise the assignments here will see the new
	// data value instead of the old and it will appear as if the values of the data fields did not change, e.g.
	// `email` would always be equal to `:email`.
	`
  email_updated = IF(email <> :email, NOW(), email_updated),
  phone_updated = IF(phone <> :phone, NOW(), phone_updated),
  address_updated = IF(street_address <> :street_address OR city <> :city OR state <> :state, NOW(), address_updated),
  location_updated = IF(street_address <> :street_address OR city <> :city OR state <> :state OR NOT location <=> :location, NOW(), location_updated),
` + ActivistDataFieldAssignments + `
WHERE
  id = :id`

const updateActivistWithTimestampsQuery string = `UPDATE activists
SET
  email_updated =    :email_updated,
  phone_updated =    :phone_updated,
  address_updated =  :address_updated,
  location_updated = :location_updated,
` + ActivistDataFieldAssignments + `
WHERE
  id = :id`

// Warning: when adding fields, test that values aren't overwritten with blank values due to unpopulated
// fields in the model object. In particular, make sure these queries / functions are updated:
//   - selectActivistExtraBaseQuery
//   - buildActivistJSONArray
//   - CleanActivistData
//   - getMergeActivistWinner
//
// Note: Does not include chapter ID to avoid accidental updates.
const ActivistDataFieldAssignments = `
  email = :email,
  facebook = :facebook,
  name = :name,
  preferred_name = :preferred_name,
  phone = :phone,
  pronouns = :pronouns,
  language = :language,
  accessibility = :accessibility,
  dob = :dob,

  activist_level = :activist_level,
  source = :source,
  hiatus = :hiatus,

  connector = :connector,
  training0 = :training0,
  training1 = :training1,
  training4 = :training4,
  training5 = :training5,
  training6 = :training6,
  consent_quiz = :consent_quiz,
  training_protest = :training_protest,
  dev_application_date = :dev_application_date,
  dev_application_type = :dev_application_type,
  dev_quiz = :dev_quiz,
  dev_interest = :dev_interest,
  cm_first_email = :cm_first_email,
  cm_approval_email = :cm_approval_email,
  prospect_organizer = :prospect_organizer,
  prospect_chapter_member = :prospect_chapter_member,
  referral_friends = :referral_friends,
  referral_apply = :referral_apply,
  referral_outlet = :referral_outlet,
  interest_date = :interest_date,
  mpi = :mpi,
  notes = :notes,
  vision_wall = :vision_wall,
  voting_agreement = :voting_agreement,
  street_address = :street_address,
  city = :city,
  state = :state,
  location = :location,
  lat = :lat,
  lng = :lng,
  discord_id = :discord_id,
  assigned_to = :assigned_to,
  followup_date = :followup_date
`

const DescOrder int = 2
const AscOrder int = 1

/** Type Definitions */

type Activist struct {
	Email           string         `db:"email"`
	EmailUpdated    time.Time      `db:"email_updated"`
	Facebook        string         `db:"facebook"`
	Hidden          bool           `db:"hidden"`
	ID              int            `db:"id"`
	Location        sql.NullString `db:"location"`
	LocationUpdated time.Time      `db:"location_updated"`
	Name            string         `db:"name"`
	PreferredName   string         `db:"preferred_name"`
	Phone           string         `db:"phone"`
	PhoneUpdated    time.Time      `db:"phone_updated"`
	Pronouns        string         `db:"pronouns"`
	Language        string         `db:"language"`
	Accessibility   string         `db:"accessibility"`
	Birthday        sql.NullString `db:"dob"`
	Coords
	ChapterID int `db:"chapter_id"`
}

type Coords struct {
	Lat float64 `db:"lat"`
	Lng float64 `db:"lng"`
}

type ActivistEventData struct {
	FirstEvent            mysql.NullTime `db:"first_event"`
	LastEvent             mysql.NullTime `db:"last_event"`
	FirstEventName        string         `db:"first_event_name"`
	LastEventName         string         `db:"last_event_name"`
	LastAction            mysql.NullTime `db:"last_action"`
	MonthsSinceLastAction int            `db:"months_since_last_action"`
	TotalEvents           int            `db:"total_events"`
	TotalPoints           int            `db:"total_points"`
	Active                bool           `db:"active"`
	Status                string
}

type ActivistMembershipData struct {
	ActivistLevel string `db:"activist_level"`
	Source        string `db:"source"`
	Hiatus        bool   `db:"hiatus"`
}

type ActivistConnectionData struct {
	Connector       string         `db:"connector"`
	Training0       sql.NullString `db:"training0"`
	Training1       sql.NullString `db:"training1"`
	Training4       sql.NullString `db:"training4"`
	Training5       sql.NullString `db:"training5"`
	Training6       sql.NullString `db:"training6"`
	ConsentQuiz     sql.NullString `db:"consent_quiz"`
	TrainingProtest sql.NullString `db:"training_protest"`
	ApplicationDate mysql.NullTime `db:"dev_application_date"`
	ApplicationType string         `db:"dev_application_type"`
	Quiz            sql.NullString `db:"dev_quiz"`
	DevInterest     string         `db:"dev_interest"`

	CMFirstEmail          sql.NullString `db:"cm_first_email"`
	CMApprovalEmail       sql.NullString `db:"cm_approval_email"`
	ProspectOrganizer     bool           `db:"prospect_organizer"`
	ProspectChapterMember bool           `db:"prospect_chapter_member"`
	LastConnection        sql.NullString `db:"last_connection"`
	ReferralFriends       string         `db:"referral_friends"`
	ReferralApply         string         `db:"referral_apply"`
	ReferralOutlet        string         `db:"referral_outlet"`
	InterestDate          sql.NullString `db:"interest_date"`
	MPI                   bool           `db:"mpi"`
	Notes                 sql.NullString `db:"notes"`
	VisionWall            string         `db:"vision_wall"`
	MPPRequirements       string         `db:"mpp_requirements"`
	VotingAgreement       bool           `db:"voting_agreement"`
	ActivistAddress
	AddressUpdated      time.Time      `db:"address_updated"`
	DiscordID           sql.NullString `db:"discord_id"`
	GeoCircles          string         `db:"geo_circles"`
	AssignedTo          int            `db:"assigned_to"`
	AssignedToName      string         `db:"assigned_to_name"`
	FollowupDate        sql.NullString `db:"followup_date"`
	TotalInteractions   int            `db:"total_interactions"`
	LastInteractionDate string         `db:"last_interaction_date"`
}

type ActivistAddress struct {
	StreetAddress string `db:"street_address"`
	City          string `db:"city"`
	State         string `db:"state"`
}

type ActivistExtra struct {
	Activist
	ActivistEventData
	ActivistMembershipData
	ActivistConnectionData
}

type ActivistDiscord struct {
	Name      string `db:"name"`
	DiscordID string `db:"discord_id"`
}

type ActivistJSON struct {
	Email         string `json:"email"`
	Facebook      string `json:"facebook"`
	ID            int    `json:"id"`
	Location      string `json:"location"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_name"`
	Phone         string `json:"phone"`
	Pronouns      string `json:"pronouns"`
	Language      string `json:"language"`
	Accessibility string `json:"accessibility"`
	Birthday      string `json:"dob"`
	ChapterID     int    `json:"chapter_id"`

	FirstEvent            string `json:"first_event"`
	LastEvent             string `json:"last_event"`
	FirstEventName        string `json:"first_event_name"`
	LastEventName         string `json:"last_event_name"`
	LastAction            string `json:"last_action"`
	MonthsSinceLastAction int    `json:"months_since_last_action"`
	TotalEvents           int    `json:"total_events"`
	TotalPoints           int    `json:"total_points"`
	Active                bool   `json:"active"`
	Status                string `json:"status"`

	ActivistLevel string `json:"activist_level"`
	Source        string `json:"source"`
	Hiatus        bool   `json:"hiatus"`

	Connector       string `json:"connector"`
	Training0       string `json:"training0"`
	Training1       string `json:"training1"`
	Training4       string `json:"training4"`
	Training5       string `json:"training5"`
	Training6       string `json:"training6"`
	ConsentQuiz     string `json:"consent_quiz"`
	TrainingProtest string `json:"training_protest"`
	ApplicationDate string `json:"dev_application_date"`
	ApplicationType string `json:"dev_application_type"`
	Quiz            string `json:"dev_quiz"`
	DevInterest     string `json:"dev_interest"`

	CMFirstEmail          string  `json:"cm_first_email"`
	CMApprovalEmail       string  `json:"cm_approval_email"`
	ProspectOrganizer     bool    `json:"prospect_organizer"`
	ProspectChapterMember bool    `json:"prospect_chapter_member"`
	LastConnection        string  `json:"last_connection"`
	ReferralFriends       string  `json:"referral_friends"`
	ReferralApply         string  `json:"referral_apply"`
	ReferralOutlet        string  `json:"referral_outlet"`
	InterestDate          string  `json:"interest_date"`
	MPI                   bool    `json:"mpi"`
	Notes                 string  `json:"notes"`
	VisionWall            string  `json:"vision_wall"`
	MPPRequirements       string  `json:"mpp_requirements"`
	VotingAgreement       bool    `json:"voting_agreement"`
	StreetAddress         string  `json:"street_address"`
	City                  string  `json:"city"`
	State                 string  `json:"state"`
	DiscordID             string  `json:"discord_id"`
	GeoCircles            string  `json:"geo_circles"`
	Lat                   float64 `json:"lat"`
	Lng                   float64 `json:"lng"`
	AssignedTo            int     `json:"assigned_to"`
	AssignedToName        string  `json:"assigned_to_name"`
	FollowupDate          string  `json:"followup_date"`
	TotalInteractions     int     `json:"total_interactions"`
	LastInteractionDate   string  `json:"last_interaction_date"`
}

type GetActivistOptions struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	Hidden                bool   `json:"hidden"`
	Order                 int    `json:"order"`
	OrderField            string `json:"order_field"`
	LastEventDateFrom     string `json:"last_event_date_from"`
	LastEventDateTo       string `json:"last_event_date_to"`
	InterestDateFrom      string `json:"interest_date_from"`
	InterestDateTo        string `json:"interest_date_to"`
	Filter                string `json:"filter"`
	ChapterID             int    `json:"chapter_id"`
	AssignedToCurrentUser bool   `json:"assigned_to_current_user"`
	AssignedTo            int    `json:"assigned_to"`
	UpcomingFollowupsOnly bool   `json:"upcoming_followups_only"`
}

var validOrderFields = map[string]struct{}{
	"a.name":        struct{}{},
	"last_event":    struct{}{},
	"total_points":  struct{}{},
	"interest_date": struct{}{},
	"followup_date": struct{}{},
}

type ActivistRangeOptionsJSON struct {
	Name      string `json:"name"`
	Limit     int    `json:"limit"`
	Order     int    `json:"order"`
	ChapterID int    `json:"chapter_id"`
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
		lastAction := ""
		if a.ActivistEventData.LastAction.Valid {
			lastAction = a.ActivistEventData.LastAction.Time.Format(EventDateLayout)
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
		consent_quiz := ""
		if a.ActivistConnectionData.ConsentQuiz.Valid {
			consent_quiz = a.ActivistConnectionData.ConsentQuiz.String
		}
		training_protest := ""
		if a.ActivistConnectionData.TrainingProtest.Valid {
			training_protest = a.ActivistConnectionData.TrainingProtest.String
		}
		quiz := ""
		if a.ActivistConnectionData.Quiz.Valid {
			quiz = a.ActivistConnectionData.Quiz.String
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
		interest_date := ""
		if a.ActivistConnectionData.InterestDate.Valid {
			interest_date = a.ActivistConnectionData.InterestDate.String
		}
		notes := ""
		if a.ActivistConnectionData.Notes.Valid {
			notes = a.ActivistConnectionData.Notes.String
		}
		discord_id := ""
		if a.ActivistConnectionData.DiscordID.Valid {
			discord_id = a.ActivistConnectionData.DiscordID.String
		}
		followup_date := ""
		if a.ActivistConnectionData.FollowupDate.Valid {
			followup_date = a.ActivistConnectionData.FollowupDate.String
		}

		activistsJSON = append(activistsJSON, ActivistJSON{
			Email:         a.Email,
			Facebook:      a.Facebook,
			ID:            a.ID,
			ChapterID:     a.ChapterID,
			Location:      location,
			Name:          a.Name,
			PreferredName: a.PreferredName,
			Phone:         a.Phone,
			Pronouns:      a.Pronouns,
			Language:      a.Language,
			Accessibility: a.Accessibility,
			Birthday:      dob,

			FirstEvent:            firstEvent,
			LastEvent:             lastEvent,
			FirstEventName:        a.FirstEventName,
			LastEventName:         a.LastEventName,
			LastAction:            lastAction,
			MonthsSinceLastAction: a.MonthsSinceLastAction,
			Status:                a.Status,
			TotalEvents:           a.TotalEvents,
			TotalPoints:           a.TotalPoints,
			Active:                a.Active,

			ActivistLevel: a.ActivistLevel,
			Source:        a.Source,
			Hiatus:        a.Hiatus,

			Connector:       a.Connector,
			Training0:       training0,
			Training1:       training1,
			Training4:       training4,
			Training5:       training5,
			Training6:       training6,
			ConsentQuiz:     consent_quiz,
			TrainingProtest: training_protest,
			ApplicationDate: applicationDate,
			ApplicationType: a.ApplicationType,
			Quiz:            quiz,
			DevInterest:     a.DevInterest,

			CMFirstEmail:          cm_first_email,
			CMApprovalEmail:       cm_approval_email,
			ProspectOrganizer:     a.ProspectOrganizer,
			ProspectChapterMember: a.ProspectChapterMember,
			LastConnection:        last_connection,
			ReferralFriends:       a.ReferralFriends,
			ReferralApply:         a.ReferralApply,
			ReferralOutlet:        a.ReferralOutlet,
			InterestDate:          interest_date,
			MPI:                   a.MPI,
			Notes:                 notes,
			VisionWall:            a.VisionWall,
			MPPRequirements:       a.MPPRequirements,
			VotingAgreement:       a.VotingAgreement,
			StreetAddress:         a.StreetAddress,
			City:                  a.City,
			State:                 a.State,
			Lat:                   a.Lat,
			Lng:                   a.Lng,
			DiscordID:             discord_id,
			GeoCircles:            a.GeoCircles,
			AssignedToName:        a.AssignedToName,
			FollowupDate:          followup_date,
			TotalInteractions:     a.TotalInteractions,
			LastInteractionDate:   a.LastInteractionDate,
		})
	}

	return activistsJSON
}

func GetActivist(db *sqlx.DB, name string, chapterID int) (Activist, error) {
	activists, err := getActivists(db, name, chapterID)
	if err != nil {
		return Activist{}, err
	} else if len(activists) == 0 {
		return Activist{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return Activist{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func GetActivists(db *sqlx.DB, chapterID int) ([]Activist, error) {
	return getActivists(db, "", chapterID)
}

func getActivists(db *sqlx.DB, name string, chapterID int) ([]Activist, error) {
	var queryArgs []interface{}
	query := selectActivistBaseQuery

	query += " WHERE chapter_id = ? "
	queryArgs = append(queryArgs, chapterID)

	if name != "" {
		query += " AND name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var activists []Activist
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists for %s", name)
	}

	return activists, nil
}

func GetActivistsByEmail(db *sqlx.DB, email string) ([]ActivistExtra, error) {
	var queryArgs []interface{}
	query := selectActivistExtraBaseQuery

	if email != "" {
		query += " WHERE email = ? AND hidden = 0"
		queryArgs = append(queryArgs, email)
	}

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists for %s", email)
	}

	return activists, nil
}

func GetActivistsWithDiscordID(db *sqlx.DB) ([]ActivistDiscord, error) {
	query := "SELECT name, discord_id from activists WHERE discord_id is not null AND hidden = 0"

	var activists []ActivistDiscord
	if err := db.Select(&activists, query); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists with Discord ID")
	}

	return activists, nil
}

func GetSFBayChapterMembers(db *sqlx.DB) ([]Activist, error) {
	query := `
SELECT
  id,
  name,
  email
FROM activists
WHERE hidden = 0 AND activist_level IN('Organizer', 'Chapter Member') AND chapter_id = ` + SFBayChapterIdStr + `
`

	var activists []Activist
	err := db.Select(&activists, query)
	if err != nil {
		return []Activist{}, errors.Wrapf(err, "GetChapterMembers: Failed retrieving activists for levels Organizer and Chapter Member")
	}

	return activists, nil
}

func GetSFBayOrganizers(db *sqlx.DB) ([]Activist, error) {
	query := `
SELECT
  id,
  name,
  email
FROM activists
WHERE hidden = 0 AND activist_level = 'Organizer' AND chapter_id = ` + SFBayChapterIdStr + `
`

	var activists []Activist
	err := db.Select(&activists, query)
	if err != nil {
		return []Activist{}, errors.Wrapf(err, "GetOrganizers: Failed retrieving activists for levels Organizer")
	}

	return activists, nil
}

func GetActivistExtra(db *sqlx.DB, id int) (*ActivistExtra, error) {
	var activist ActivistExtra
	query := selectActivistExtraBaseQuery + " where a.id = ?"
	if err := db.Get(&activist, query, id); err != nil {
		return nil, errors.Wrapf(err, "failed to get activist with id %d", id)
	}
	return &activist, nil
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
		var whereClause []string

		if options.Name != "" {
			whereClause = append(whereClause, "a.name like '%"+options.Name+"%'")
		}

		if options.ChapterID != 0 {
			whereClause = append(whereClause, "a.chapter_id = "+strconv.Itoa(options.ChapterID))
		}

		if options.Hidden {
			whereClause = append(whereClause, "a.hidden = true")
		} else {
			whereClause = append(whereClause, "a.hidden = false")
		}

		if options.AssignedToCurrentUser {
			whereClause = append(whereClause, "assigned_to = ?")
			queryArgs = append(queryArgs, options.AssignedTo)
		}

		switch options.Filter {
		case "development":
			whereClause = append(whereClause, "a.activist_level like '%organizer'")
		case "chapter_member_prospects":
			whereClause = append(whereClause, "prospect_chapter_member = true AND a.activist_level <> 'chapter member' and a.activist_level not like '%organizer'")
		case "organizer_prospects":
			whereClause = append(whereClause, "prospect_organizer = true AND a.activist_level not like '%organizer'")
		case "chapter_member_development":
			whereClause = append(whereClause, "(a.activist_level like '%organizer' OR a.activist_level = 'chapter member')")
		case "community_prospects":
			whereClause = append(whereClause, "(source like '%form%' or source like 'petition%' or source like 'eventbrite%' or source='dxe-signup' or source='arc-signup') and source not like '%application%'")
			whereClause = append(whereClause, "activist_level = 'supporter'")
			//whereClause = append(whereClause, "interest_date >= DATE_SUB(now(), INTERVAL 3 MONTH)")
			//whereClause = append(whereClause, "a.id not in (select distinct activist_id from event_attendance)")
			// TODO: consider hiding people if they have attended 3+ events?
		case "community_prospects_followup":
			if options.UpcomingFollowupsOnly {
				whereClause = append(whereClause, "date(followup_date) > CURRENT_DATE")
			} else {
				whereClause = append(whereClause, "date(followup_date) <= CURRENT_DATE")
			}
			whereClause = append(whereClause, "assigned_to <> 0")
		case "leaderboard":
			whereClause = append(whereClause, "a.id in (select distinct activist_id  from event_attendance ea  where ea.event_id in (select id from events e where e.date >= (now() - interval 30 day)))")

		case "new_activists":
			whereClause = append(whereClause, `
				(
					SELECT COUNT(*) 
					FROM event_attendance ea
					JOIN events e ON ea.event_id = e.id
					WHERE ea.activist_id = a.id 
				) <= 3
				AND
				(
					SELECT MAX(e.date)
					FROM event_attendance ea
					JOIN events e ON ea.event_id = e.id
					WHERE ea.activist_id = a.id
				) BETWEEN ? AND ?`)
			queryArgs = append(queryArgs, options.LastEventDateFrom, options.LastEventDateTo)

		case "new_activists_pending_workshop":
			whereClause = append(whereClause, `
				training0 is NULL AND
				activist_level = 'supporter' AND
				(
					SELECT MIN(e.date)
					FROM event_attendance ea
					JOIN events e ON ea.event_id = e.id
					WHERE ea.activist_id = a.id
				) BETWEEN ? AND ?`)
			queryArgs = append(queryArgs, options.LastEventDateFrom, options.LastEventDateTo)

		}

		if len(whereClause) != 0 {
			query += " WHERE " + strings.Join(whereClause, " AND ")
		}
	}

	var havingClause []string
	if options.LastEventDateFrom != "" {
		havingClause = append(havingClause, "last_event >= ?")
		queryArgs = append(queryArgs, options.LastEventDateFrom)
	}
	if options.LastEventDateTo != "" {
		havingClause = append(havingClause, "last_event <= ?")
		queryArgs = append(queryArgs, options.LastEventDateTo)
	}
	if options.InterestDateFrom != "" {
		havingClause = append(havingClause, "interest_date >= ?")
		queryArgs = append(queryArgs, options.InterestDateFrom)
	}
	if options.InterestDateTo != "" {
		havingClause = append(havingClause, "interest_date < DATE_FORMAT(?, '%Y-%m-%d 23:59:59')")
		queryArgs = append(queryArgs, options.InterestDateTo)
	}
	if options.Filter == "community_prospects" {
		havingClause = append(havingClause, "total_interactions = 0")
		havingClause = append(havingClause, "(last_event < DATE_SUB(NOW(), INTERVAL 12 MONTH) or last_event is NULL)")
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

	if options.ChapterID != 0 {
		query += " AND a.chapter_id = ? "
		queryArgs = append(queryArgs, options.ChapterID)
	}

	if name != "" {
		if order == DescOrder {
			query += " AND a.name < ? "
		} else {
			query += " AND a.name > ? "
		}
		queryArgs = append(queryArgs, name)
	}

	query += " GROUP BY a.id "

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

func GetOrCreateActivist(db *sqlx.DB, name string, chapterID int) (Activist, error) {
	activist, err := GetActivist(db, name, chapterID)
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

	_, err = tx.Exec("INSERT INTO activists (name, chapter_id) VALUES (?, ?)", name, chapterID)
	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to insert activist %s", name)
	}

	query := selectActivistBaseQuery + " WHERE name = ? AND chapter_id = ?"

	var newActivist Activist
	err = tx.Get(&newActivist, query, name, chapterID)

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

func createActivist(db *sqlx.DB, activist ActivistExtra, query string) (int, error) {
	if activist.ID != 0 {
		return 0, errors.New("Activist ID must be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	result, err := db.NamedExec(query, activist)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not create activist: %s", activist.Name)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "Could not get LastInsertId for %s", activist.Name)
	}
	return int(id), nil
}

func CreateActivist(db *sqlx.DB, activist ActivistExtra) (int, error) {
	return createActivist(db, activist, insertActivistQuery)
}

func CreateActivistWithTimestamps(db *sqlx.DB, activist ActivistExtra) (int, error) {
	return createActivist(db, activist, insertActivistWithTimestampsQuery)
}

func UpdateActivistData(db *sqlx.DB, activist ActivistExtra, userEmail string) (int, error) {
	if activist.ID == 0 {
		return 0, errors.New("activist ID cannot be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	// get original activist to see if we need to update the mailing list
	orig, err := GetActivistsExtra(db, GetActivistOptions{
		ID: activist.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("error fetching existing activist data: %v", err)
	}
	origActivist := orig[0]

	mailingListInfoChanged := activist.Name != origActivist.Name ||
		activist.Email != origActivist.Email ||
		activist.Phone != origActivist.Phone ||
		activist.Location != origActivist.Location ||
		activist.City != origActivist.City ||
		activist.State != origActivist.State ||
		activist.ActivistLevel != origActivist.ActivistLevel
	if mailingListInfoChanged && activist.Email != "" {
		signup := mailing_list_signup.Signup{
			Source: "adb",
			Name:   activist.Name,
			Email:  activist.Email,
			Phone:  activist.Phone,
			City:   activist.City,
			State:  activist.State,
			// Zip will be used to find a chapter mailing list (often this chapter's mailing list).
			// Activist may be added to ADB chapter near this zip, if different from this chapter.
			Zip: activist.Location.String,
			// Let signup service know this chapter's ID so it doesn't try to sync back here causing an infinite loop.
			SourceChapterId: activist.ChapterID,
			ActivistLevel:   activist.ActivistLevel,
		}
		err := mailing_list_signup.Enqueue(signup)
		if err != nil {
			// Don't return this error because we still want to successfully update the activist in the database.
			log.Println("ERROR updating activist on mailing list:", err.Error())
		} else {
			log.Printf("Pushed updated activist record to sign-up service: name: %v, email: %v, chapter: %v",
				activist.Name, activist.Email, activist.ChapterID)
		}
	}
	geoInfoChanged := activist.City != origActivist.City ||
		activist.State != origActivist.State ||
		activist.StreetAddress != origActivist.StreetAddress
	if geoInfoChanged && activist.StreetAddress != "" && activist.City != "" && activist.State != "" {
		location := geoCodeAddress(activist.StreetAddress, activist.City, activist.State)
		if location != nil {
			activist.Lng = location.Lng
			activist.Lat = location.Lat
		}
	}

	_, err = db.NamedExec(updateActivistQuery, activist)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update activist data")
	}
	log.Printf("Updated data for activist %v", activist.Name)

	// LOGGING (work in progress)
	_, err = db.NamedExec(`INSERT INTO activists_history (activist_id, action, user_email, name, email, facebook, activist_level)
	VALUES (
		:id,
		'UPDATE',
		:user_email,
		:name,
		:email,
		:facebook,
		:activist_level
	)`, struct {
		UserEmail string `db:"user_email"`
		*ActivistExtra
	}{
		userEmail,
		&activist,
	})

	if err != nil {
		log.Println("Error logging activist update: " + err.Error())
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
//   - The original activist is hidden
//   - All of the original activist's event attendance is updated to be the target activist.
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

	_, err = updateMergedActivistDataDetails(tx, originalActivistID, targetActivistID)
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
		"Supporter":             0,
		"Non-Local":             1,
		"Global Network Member": 2,
		"Chapter Member":        3,
		"Organizer":             4,
	}

	// Check boolean values

	target.ProspectOrganizer = boolMerge(original.ProspectOrganizer, target.ProspectOrganizer)
	target.ProspectChapterMember = boolMerge(original.ProspectChapterMember, target.ProspectChapterMember)
	target.MPI = boolMerge(original.MPI, target.MPI)
	target.Hiatus = boolMerge(original.Hiatus, target.Hiatus)
	target.VotingAgreement = boolMerge(original.VotingAgreement, target.VotingAgreement)

	// Check string fields for empty values

	target.Email, target.EmailUpdated = stringMergeWithTimestamps(original.Email, original.EmailUpdated, target.Email, target.EmailUpdated)
	target.Phone, target.PhoneUpdated = stringMergeWithTimestamps(original.Phone, original.PhoneUpdated, target.Phone, target.PhoneUpdated)
	target.Pronouns = stringMerge(original.Pronouns, target.Pronouns)
	target.Language = stringMerge(original.Language, target.Language)
	target.Accessibility = stringMerge(original.Accessibility, target.Accessibility)
	target.Birthday = stringMergeSqlNullString(original.Birthday, target.Birthday)
	target.Location, target.LocationUpdated = stringMergeSqlNullStringWithTimestamps(original.Location, original.LocationUpdated, target.Location, target.LocationUpdated)
	target.Facebook = stringMerge(original.Facebook, target.Facebook)
	target.Connector = stringMerge(original.Connector, target.Connector)
	target.Source = stringMerge(original.Source, target.Source)
	target.Training0 = stringMergeSqlNullString(original.Training0, target.Training0)
	target.Training1 = stringMergeSqlNullString(original.Training1, target.Training1)
	target.Training4 = stringMergeSqlNullString(original.Training4, target.Training4)
	target.Training5 = stringMergeSqlNullString(original.Training5, target.Training5)
	target.Training6 = stringMergeSqlNullString(original.Training6, target.Training6)
	target.ConsentQuiz = stringMergeSqlNullString(original.ConsentQuiz, target.ConsentQuiz)
	target.TrainingProtest = stringMergeSqlNullString(original.TrainingProtest, target.TrainingProtest)
	target.DevInterest = stringMerge(original.DevInterest, target.DevInterest)
	target.ApplicationDate = stringMergeSqlNullTime(original.ApplicationDate, target.ApplicationDate)
	target.Quiz = stringMergeSqlNullString(original.Quiz, target.Quiz)
	target.CMFirstEmail = stringMergeSqlNullString(original.CMFirstEmail, target.CMFirstEmail)
	target.CMApprovalEmail = stringMergeSqlNullString(original.CMApprovalEmail, target.CMApprovalEmail)
	target.ReferralFriends = stringMerge(original.ReferralFriends, target.ReferralFriends)
	target.ReferralApply = stringMerge(original.ReferralApply, target.ReferralApply)
	target.ReferralOutlet = stringMerge(original.ReferralOutlet, target.ReferralOutlet)
	target.InterestDate = stringMergeSqlNullString(original.InterestDate, target.InterestDate)
	target.Notes = stringMergeSqlNullString(original.Notes, target.Notes)
	target.VisionWall = stringMerge(original.VisionWall, target.VisionWall)
	target.ApplicationType = stringMerge(original.ApplicationType, target.ApplicationType)
	target.ActivistAddress, target.Coords, target.AddressUpdated = mergeAddress(original.ActivistAddress, original.Coords, original.AddressUpdated, target.ActivistAddress, target.Coords, target.AddressUpdated)
	target.DiscordID = stringMergeSqlNullString(original.DiscordID, target.DiscordID)

	// The location field is considered to be at least as up-to-date as the address fields.
	// See comments on location_updated SQL column for details.
	if target.LocationUpdated.Before(target.AddressUpdated) {
		target.LocationUpdated = target.AddressUpdated
	}

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

func stringMergeWithTimestamps(original string, originalTimestamp time.Time, target string, targetTimestamp time.Time) (string, time.Time) {
	if targetTimestamp.After(originalTimestamp) && len(target) > 0 {
		return target, targetTimestamp
	}
	if originalTimestamp.After(targetTimestamp) && len(original) > 0 {
		return original, originalTimestamp
	}
	return stringMerge(original, target), targetTimestamp
}

func stringMergeSqlNullString(original sql.NullString, target sql.NullString) sql.NullString {
	if !target.Valid && original.Valid {
		return original
	}

	return target
}

func stringMergeSqlNullStringWithTimestamps(original sql.NullString, originalTimestamp time.Time, target sql.NullString, targetTimestamp time.Time) (sql.NullString, time.Time) {
	if targetTimestamp.After(originalTimestamp) && target.Valid && len(target.String) > 0 {
		return target, targetTimestamp
	}
	if originalTimestamp.After(targetTimestamp) && original.Valid && len(original.String) > 0 {
		return original, originalTimestamp
	}
	return stringMergeSqlNullString(original, target), targetTimestamp
}

func stringMergeSqlNullTime(original mysql.NullTime, target mysql.NullTime) mysql.NullTime {
	if !target.Valid && original.Valid {
		return original
	}

	return target
}

// Merges address and coords. In ADB, coords are computed from address, so they are merged atomically with the adddress here.
func mergeAddress(originalAddr ActivistAddress, originalCoords Coords, originalUpdated time.Time, target ActivistAddress, targetCoords Coords, targetUpdated time.Time) (ActivistAddress, Coords, time.Time) {
	// Determine which address is newerAddr
	newerAddr, newerCoords, newerUpdated := target, targetCoords, targetUpdated
	olderAddr, olderCoords, olderUpdated := originalAddr, originalCoords, originalUpdated
	if originalUpdated.After(targetUpdated) {
		newerAddr, newerCoords, newerUpdated = originalAddr, originalCoords, originalUpdated
		olderAddr, olderCoords, olderUpdated = target, targetCoords, targetUpdated
	}
	// Return older if newer is empty
	if newerAddr.StreetAddress == "" && newerAddr.City == "" && newerAddr.State == "" &&
		(olderAddr.StreetAddress != "" || olderAddr.City != "" || olderAddr.State != "") {
		return olderAddr, olderCoords, olderUpdated
	}
	addr := newerAddr
	// If newer is missing city, use from older if both have the same state
	if addr.City == "" && olderAddr.City != "" && addr.State == olderAddr.State {
		addr.City = olderAddr.City
	}
	// If newer is missing street address, use from older if both have the same city
	if addr.StreetAddress == "" && olderAddr.StreetAddress != "" && addr.City == olderAddr.City && addr.State == olderAddr.State {
		addr.StreetAddress = olderAddr.StreetAddress
	}
	return addr, newerCoords, newerUpdated
}

func updateMergedActivistDataDetails(tx *sqlx.Tx, originalActivistID int, targetActivistID int) (*ActivistExtra, error) {
	// Merge details of original activist into target activist
	// Favor booleans that are set to TRUE, and pull in missing data from original activist to target; when both
	// activists have data for the same field, we should use the target activist's data.

	query := selectActivistExtraBaseQuery + " WHERE a.id = ?"

	var originalActivist = new(ActivistExtra)
	err := tx.Get(originalActivist, query, originalActivistID)
	if err != nil || originalActivist == nil {
		return nil, errors.Wrapf(err, "failed to get original activist with id %d", originalActivistID)
	}

	var targetActivist = new(ActivistExtra)
	err = tx.Get(targetActivist, query, targetActivistID)
	if err != nil || targetActivist == nil {
		return nil, errors.Wrapf(err, "failed to get target activist with id %d", targetActivistID)
	}

	mergedActivist := getMergeActivistWinner(*originalActivist, *targetActivist)

	_, err = tx.NamedExec(updateActivistWithTimestampsQuery, mergedActivist)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to update activist with id %d", targetActivistID)
	}

	return &mergedActivist, nil
}

func GetAutocompleteNames(db *sqlx.DB, chapterID int) []string {
	type Name struct {
		Name string `db:"name"`
	}
	var names []Name
	// Order the activists by the last even they've been to.
	err := db.Select(&names, `
SELECT a.name FROM activists a
LEFT OUTER JOIN event_attendance ea ON a.id = ea.activist_id
LEFT OUTER JOIN events e ON e.id = ea.event_id
WHERE a.hidden = 0 AND a.chapter_id = ?
GROUP BY a.name
ORDER BY MAX(e.date) DESC`, chapterID)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	var ret []string
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

func GetAutocompleteOrganizerNames(db *sqlx.DB, chapterID int) []string {
	// includes non-local activist level for ppl to be added to working groups
	type Name struct {
		Name string `db:"name"`
	}
	var names []Name
	err := db.Select(&names, `
SELECT a.name FROM activists a
WHERE a.hidden = 0 and a.activist_level in ('non-local', 'organizer') AND a.chapter_id = ?
GROUP BY a.name`, chapterID)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	var ret []string
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

func GetAutocompleteChapterMembersNames(db *sqlx.DB, chapterID int) []string {
	type Name struct {
		Name string `db:"name"`
	}
	var names []Name
	err := db.Select(&names, `
SELECT a.name FROM activists a
WHERE a.hidden = 0 and a.activist_level in ('chapter member', 'organizer') AND a.chapter_id = ?
GROUP BY a.name`, chapterID)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	var ret []string
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

type ActivistBasicInfoJSON struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Pronouns      string `json:"pronouns"`
	PreferredName string `json:"preferred_name"`
}

type ActivistBasicInfo struct {
	Name          string `db:"name"`
	Email         string `db:"email"`
	Phone         string `db:"phone"`
	Pronouns      string `db:"pronouns"`
	PreferredName string `db:"preferred_name"`
}

func (activist *ActivistBasicInfo) ToJSON() ActivistBasicInfoJSON {
	return ActivistBasicInfoJSON{
		Name:          activist.Name,
		Email:         activist.Email,
		Phone:         activist.Phone,
		Pronouns:      activist.Pronouns,
		PreferredName: activist.PreferredName,
	}
}

func GetActivistListBasicJSON(db *sqlx.DB, chapterID int) []ActivistBasicInfoJSON {
	var activists []ActivistBasicInfo

	// Order the activists by the last even they've been to.
	err := db.Select(&activists, `
SELECT a.name, a.email, a.phone, a.pronouns, a.preferred_name FROM activists a
LEFT OUTER JOIN event_attendance ea ON a.id = ea.activist_id
LEFT OUTER JOIN events e ON e.id = ea.event_id
WHERE a.hidden = 0 AND a.chapter_id = ?
GROUP BY a.id
ORDER BY MAX(e.date) DESC`, chapterID)
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

type ActivistSpokeInfo struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Cell      string `db:"cell"`
	LastEvent string `db:"last_event"`
}

type NewActivistPendingWorkshopSpokeInfo struct {
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	Cell       string `db:"cell"`
	FirstEvent string `db:"first_event"`
}

func GetChapterMemberSpokeInfo(db *sqlx.DB, chapterID int) ([]ActivistSpokeInfo, error) {
	const query = `
		SELECT
			IF(preferred_name <> '', preferred_name, substring_index(name, " ", 1)) as first_name,
			SUBSTRING(name, LOCATE(' ', name)+1) as last_name,
			phone as cell
		FROM activists
		WHERE
		    chapter_id = ? and
			activist_level in ('chapter member', 'organizer')
			and hidden = 0
	`

	var activists []ActivistSpokeInfo
	err := db.Select(&activists, query, chapterID)
	if err != nil {
		return []ActivistSpokeInfo{}, err
	}

	return activists, nil
}

func GetSupporterSpokeInfo(db *sqlx.DB, chapterID int, startDate, endDate string) ([]ActivistSpokeInfo, error) {
	query := `
		SELECT
			IF(preferred_name <> '', preferred_name, substring_index(name, " ", 1)) as first_name,
			SUBSTRING(name, LOCATE(' ', name)+1) as last_name,
			phone as cell,
		    @last_event := (
				SELECT max(e.date) AS max_date
				FROM event_attendance ea
				JOIN activists inner_a ON inner_a.id = ea.activist_id
				JOIN events e ON e.id = ea.event_id
				WHERE inner_a.id = activists.id
			) AS last_event
		FROM activists
		WHERE
		    chapter_id = ?
		    and activist_level = 'supporter'
			and hidden = 0
	`
	args := []interface{}{chapterID}

	var havingClause []string
	if startDate != "" {
		havingClause = append(havingClause, "last_event >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		havingClause = append(havingClause, "last_event <= ?")
		args = append(args, endDate)
	}
	if len(havingClause) != 0 {
		query += " HAVING " + strings.Join(havingClause, " AND ")
	}

	var activists []ActivistSpokeInfo
	err := db.Select(&activists, query, args...)
	if err != nil {
		return []ActivistSpokeInfo{}, err
	}

	return activists, nil
}

func GetNewActivistsSpokeInfo(db *sqlx.DB, chapterID int, startDate, endDate string) ([]ActivistSpokeInfo, error) {
	last_event_subquery := `
		(
			SELECT MAX(e.date)
			FROM event_attendance ea
			JOIN events e ON ea.event_id = e.id
			WHERE ea.activist_id = activists.id
		)`

	query := `
		SELECT
			IF(preferred_name <> '', preferred_name, substring_index(name, " ", 1)) as first_name,
			SUBSTRING(name, LOCATE(' ', name)+1) as last_name,
			phone as cell,
			@last_event := ` + last_event_subquery + ` AS last_event
		FROM activists
		WHERE
			chapter_id = ?
			AND hidden = 0
			AND (
					SELECT COUNT(*) 
					FROM event_attendance ea
					JOIN events e ON ea.event_id = e.id
					WHERE ea.activist_id = activists.id 
				) <= 3
			AND ` + last_event_subquery + ` BETWEEN ? AND ?
	`
	args := []interface{}{chapterID, startDate, endDate}

	var activists []ActivistSpokeInfo
	err := db.Select(&activists, query, args...)
	if err != nil {
		return []ActivistSpokeInfo{}, err
	}

	return activists, nil
}

func GetNewActivistsPendingWorkshopSpokeInfo(db *sqlx.DB, chapterID int, startDate, endDate string) ([]NewActivistPendingWorkshopSpokeInfo, error) {
	first_event_subquery := `
		(
			SELECT MIN(e.date)
			FROM event_attendance ea
			JOIN events e ON ea.event_id = e.id
			WHERE ea.activist_id = activists.id
		)`

	query := `
		SELECT
			IF(preferred_name <> '', preferred_name, substring_index(name, " ", 1)) as first_name,
			SUBSTRING(name, LOCATE(' ', name)+1) as last_name,
			phone as cell,
			@first_event := ` + first_event_subquery + ` AS first_event
		FROM activists
		WHERE
			chapter_id = ?
			AND hidden = 0
			AND training0 IS NULL
			AND activist_level = 'supporter'
			AND ` + first_event_subquery + ` BETWEEN ? AND ?
	`
	args := []interface{}{chapterID, startDate, endDate}

	var activists []NewActivistPendingWorkshopSpokeInfo
	err := db.Select(&activists, query, args...)
	if err != nil {
		return []NewActivistPendingWorkshopSpokeInfo{}, err
	}

	return activists, nil
}

type CommunityProspectHubSpotInfo struct {
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Email        string `db:"email"`
	Phone        string `db:"phone"`
	Zip          string `db:"zip"`
	Source       string `db:"source"`
	InterestDate string `db:"interest_date"`
}

func GetCommunityProspectHubSpotInfo(db *sqlx.DB, chapterID int) ([]CommunityProspectHubSpotInfo, error) {
	var activists []CommunityProspectHubSpotInfo

	// Order the activists by the last even they've been to.
	err := db.Select(&activists, `
		SELECT 
			IF(preferred_name <> '', preferred_name, substring_index(name, " ", 1)) as first_name,
			SUBSTRING(name, LOCATE(' ', name)) as last_name,
			email, phone, IFNULL(location,'') as zip, source,
			interest_date
		FROM activists
		WHERE (source like '%form%' or source like 'petition%' or source like 'eventbrite%' or source='dxe-signup' or source='arc-signup') and source not like '%application%' and source != "Check-in Form"
		and activist_level = 'supporter'
		and interest_date >= DATE_SUB(now(), INTERVAL 3 MONTH)
		and activists.id not in (select distinct activist_id from event_attendance)
		and hidden = 0
		and chapter_id = ?
		ORDER BY interest_date desc
`, chapterID)
	if err != nil {
		return []CommunityProspectHubSpotInfo{}, err
	}

	return activists, nil
}

func CleanActivistData(body io.Reader, db *sqlx.DB) (ActivistExtra, error) {
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
	validConsentQuiz := true
	if activistJSON.ConsentQuiz == "" {
		// Not specified so insert null value into database
		validConsentQuiz = false
	}
	validTrainingProtest := true
	if activistJSON.TrainingProtest == "" {
		// Not specified so insert null value into database
		validTrainingProtest = false
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
	validQuiz := true
	if activistJSON.Quiz == "" {
		// Not specified so insert null value into database
		validQuiz = false
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
	validDiscordID := true
	if activistJSON.DiscordID == "" {
		validDiscordID = false
	}
	validFollowupDate := true
	if activistJSON.FollowupDate == "" {
		validFollowupDate = false
	}

	var assignedToInt int
	assignedToName := strings.TrimSpace(activistJSON.AssignedToName)
	if assignedToName != "" {
		users, err := getUsers(db, GetUserOptions{Name: assignedToName})
		if err != nil {
			panic(err)
		}
		if len(users) == 0 {
			return ActivistExtra{}, errors.New("invalid name for 'assigned to' field")
		}
		assignedToInt = users[0].ID
	}

	activistExtra := ActivistExtra{
		Activist: Activist{
			Email:         strings.TrimSpace(activistJSON.Email),
			Facebook:      strings.TrimSpace(activistJSON.Facebook),
			ID:            activistJSON.ID,
			ChapterID:     activistJSON.ChapterID,
			Location:      sql.NullString{String: strings.TrimSpace(activistJSON.Location), Valid: validLoc},
			Name:          strings.TrimSpace(activistJSON.Name),
			PreferredName: strings.TrimSpace(activistJSON.PreferredName),
			Phone:         strings.TrimSpace(activistJSON.Phone),
			Pronouns:      strings.TrimSpace(activistJSON.Pronouns),
			Language:      strings.TrimSpace(activistJSON.Language),
			Accessibility: strings.TrimSpace(activistJSON.Accessibility),
			Birthday:      sql.NullString{String: strings.TrimSpace(activistJSON.Birthday), Valid: validBirthday},
			Coords: Coords{
				Lat: activistJSON.Lat,
				Lng: activistJSON.Lng,
			},
		},
		ActivistMembershipData: ActivistMembershipData{
			ActivistLevel: strings.TrimSpace(activistJSON.ActivistLevel),
			Source:        strings.TrimSpace(activistJSON.Source),
			Hiatus:        activistJSON.Hiatus,
		},
		ActivistConnectionData: ActivistConnectionData{
			Connector:       strings.TrimSpace(activistJSON.Connector),
			Training0:       sql.NullString{String: strings.TrimSpace(activistJSON.Training0), Valid: validTraining0},
			Training1:       sql.NullString{String: strings.TrimSpace(activistJSON.Training1), Valid: validTraining1},
			Training4:       sql.NullString{String: strings.TrimSpace(activistJSON.Training4), Valid: validTraining4},
			Training5:       sql.NullString{String: strings.TrimSpace(activistJSON.Training5), Valid: validTraining5},
			Training6:       sql.NullString{String: strings.TrimSpace(activistJSON.Training6), Valid: validTraining6},
			ConsentQuiz:     sql.NullString{String: strings.TrimSpace(activistJSON.ConsentQuiz), Valid: validConsentQuiz},
			TrainingProtest: sql.NullString{String: strings.TrimSpace(activistJSON.TrainingProtest), Valid: validTrainingProtest},
			DevInterest:     strings.TrimSpace(activistJSON.DevInterest),
			Quiz:            sql.NullString{String: strings.TrimSpace(activistJSON.Quiz), Valid: validQuiz},

			CMFirstEmail:          sql.NullString{String: strings.TrimSpace(activistJSON.CMFirstEmail), Valid: validCMFirstEmail},
			CMApprovalEmail:       sql.NullString{String: strings.TrimSpace(activistJSON.CMApprovalEmail), Valid: validCMApprovalEmail},
			ProspectOrganizer:     activistJSON.ProspectOrganizer,
			ProspectChapterMember: activistJSON.ProspectChapterMember,
			ReferralFriends:       strings.TrimSpace(activistJSON.ReferralFriends),
			ReferralApply:         strings.TrimSpace(activistJSON.ReferralApply),
			ReferralOutlet:        strings.TrimSpace(activistJSON.ReferralOutlet),
			InterestDate:          sql.NullString{String: strings.TrimSpace(activistJSON.InterestDate), Valid: validInterestDate},
			MPI:                   activistJSON.MPI,
			Notes:                 sql.NullString{String: strings.TrimSpace(activistJSON.Notes), Valid: validNotes},
			VisionWall:            strings.TrimSpace(activistJSON.VisionWall),
			MPPRequirements:       strings.TrimSpace(activistJSON.MPPRequirements),
			VotingAgreement:       activistJSON.VotingAgreement,
			ActivistAddress: ActivistAddress{
				StreetAddress: strings.TrimSpace(activistJSON.StreetAddress),
				City:          strings.TrimSpace(activistJSON.City),
				State:         strings.TrimSpace(activistJSON.State),
			},
			DiscordID:    sql.NullString{String: strings.TrimSpace(activistJSON.DiscordID), Valid: validDiscordID},
			AssignedTo:   assignedToInt,
			FollowupDate: sql.NullString{String: strings.TrimSpace(activistJSON.FollowupDate), Valid: validFollowupDate},
		},
	}

	if err := validateActivist(activistExtra); err != nil {
		return ActivistExtra{}, err
	}

	return activistExtra, nil

}

var validActivistLevels = map[string]bool{
	"Supporter":             true,
	"Chapter Member":        true,
	"Organizer":             true,
	"Non-Local":             true,
	"Global Network Member": true,
}

func validateActivist(a ActivistExtra) error {
	if !validActivistLevels[a.ActivistLevel] {
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

	// remove anything from name field that isn't a character, space, number, or period
	reg, err := regexp.Compile("[^a-zA-Z0-9. ]+")
	if err != nil {
		return GetActivistOptions{}, errors.New("Error validating activist name regex")
	}
	a.Name = reg.ReplaceAllString(a.Name, "")

	// remove anything from date fields that aren't numbers or dashes
	reg, err = regexp.Compile("[^0-9-]+")
	if err != nil {
		return GetActivistOptions{}, errors.New("Error validating activist date regex")
	}
	a.LastEventDateFrom = reg.ReplaceAllString(a.LastEventDateFrom, "")
	a.LastEventDateTo = reg.ReplaceAllString(a.LastEventDateTo, "")
	a.InterestDateFrom = reg.ReplaceAllString(a.InterestDateFrom, "")
	a.InterestDateTo = reg.ReplaceAllString(a.InterestDateTo, "")

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
//   - Current
//   - New
//   - Former
//   - No attendance
//
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

func resetFollowupDate(db *sqlx.DB, activistID int) error {
	res, err := db.Exec(`UPDATE activists SET followup_date = null WHERE id = ?
		`, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update (reset) activist follow-up date")
	}
	if rows, _ := res.RowsAffected(); rows != 1 {
		return errors.New("failed to update (reset) activist follow-up date (no rows affected)")
	}
	return nil
}

func setFollowupDate(db *sqlx.DB, activistID, days int) error {
	_, err := db.Exec(`UPDATE activists SET followup_date = DATE_ADD(CURDATE(), INTERVAL ? DAY) WHERE id = ?
		`, days, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update (set) activist follow-up date")
	}
	return nil
}

func assignActivistToUser(db *sqlx.DB, activistID, userID int) error {
	_, err := db.Exec(`UPDATE activists SET assigned_to = ? WHERE id = ?
		`, userID, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update activist assigned_to value")
	}
	return nil
}
