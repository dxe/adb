package transport

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type QueryActivistResultJSON struct {
	Activists  []model.ActivistJSON    `json:"activists"`
	Pagination QueryActivistPagination `json:"pagination"`
}

type QueryActivistPagination struct {
	NextCursor string `json:"next_cursor"`
}

// ActivistPatchInput represents PATCH /api/activists/{id} request body fields.
// Only non-nil fields are applied.
// Directly setting Lat/Lng is not implemented, but these fields can be updated
// via changes to other location fields.
// Updating chapter ID and dev application date/type is not supported.
type ActivistPatchInput struct {
	// Activist fields
	Email         *string `json:"email"`
	Facebook      *string `json:"facebook"`
	Name          *string `json:"name"`
	PreferredName *string `json:"preferred_name"`
	Phone         *string `json:"phone"`
	Pronouns      *string `json:"pronouns"`
	Language      *string `json:"language"`
	Accessibility *string `json:"accessibility"`
	Birthday      *string `json:"dob"`
	Location      *string `json:"location"`

	// ActivistMembershipData fields
	ActivistLevel *string `json:"activist_level"`
	Source        *string `json:"source"`
	Hiatus        *bool   `json:"hiatus"`

	// ActivistConnectionData fields (user-editable)
	Connector             *string `json:"connector"`
	Training0             *string `json:"training0"`
	Training1             *string `json:"training1"`
	Training4             *string `json:"training4"`
	Training5             *string `json:"training5"`
	Training6             *string `json:"training6"`
	ConsentQuiz           *string `json:"consent_quiz"`
	TrainingProtest       *string `json:"training_protest"`
	DevQuiz               *string `json:"dev_quiz"`
	DevInterest           *string `json:"dev_interest"`
	CMFirstEmail          *string `json:"cm_first_email"`
	CMApprovalEmail       *string `json:"cm_approval_email"`
	ProspectOrganizer     *bool   `json:"prospect_organizer"`
	ProspectChapterMember *bool   `json:"prospect_chapter_member"`
	ReferralFriends       *string `json:"referral_friends"`
	ReferralApply         *string `json:"referral_apply"`
	ReferralOutlet        *string `json:"referral_outlet"`
	InterestDate          *string `json:"interest_date"`
	Notes                 *string `json:"notes"`
	VisionWall            *string `json:"vision_wall"`
	VotingAgreement       *bool   `json:"voting_agreement"`
	StreetAddress         *string `json:"street_address"`
	City                  *string `json:"city"`
	State                 *string `json:"state"`
	AssignedTo            *int    `json:"assigned_to"`
	FollowupDate          *string `json:"followup_date"`
}

// ToPatchData converts transport PATCH input into model patch fields.
func (p ActivistPatchInput) ToPatchData() model.ActivistPatchData {
	var d model.ActivistPatchData

	addString := func(name model.ActivistColumnName, ptr *string) {
		if ptr != nil {
			d.Append(name, strings.TrimSpace(*ptr))
		}
	}
	addNullableString := func(name model.ActivistColumnName, ptr *string) {
		if ptr != nil {
			v := strings.TrimSpace(*ptr)
			d.Append(name, sql.NullString{String: v, Valid: v != ""})
		}
	}
	addBool := func(name model.ActivistColumnName, ptr *bool) {
		if ptr != nil {
			d.Append(name, *ptr)
		}
	}
	addInt := func(name model.ActivistColumnName, ptr *int) {
		if ptr != nil {
			d.Append(name, *ptr)
		}
	}

	addString(model.ColEmail, p.Email)
	addString(model.ColFacebook, p.Facebook)
	addString(model.ColName, p.Name)
	addString(model.ColPreferredName, p.PreferredName)
	addString(model.ColPhone, p.Phone)
	addString(model.ColPronouns, p.Pronouns)
	addString(model.ColLanguage, p.Language)
	addString(model.ColAccessibility, p.Accessibility)
	addNullableString(model.ColDOB, p.Birthday)
	addNullableString(model.ColLocation, p.Location)

	addString(model.ColActivistLevel, p.ActivistLevel)
	addString(model.ColSource, p.Source)
	addBool(model.ColHiatus, p.Hiatus)

	addString(model.ColConnector, p.Connector)
	addNullableString(model.ColTraining0, p.Training0)
	addNullableString(model.ColTraining1, p.Training1)
	addNullableString(model.ColTraining4, p.Training4)
	addNullableString(model.ColTraining5, p.Training5)
	addNullableString(model.ColTraining6, p.Training6)
	addNullableString(model.ColConsentQuiz, p.ConsentQuiz)
	addNullableString(model.ColTrainingProtest, p.TrainingProtest)
	addNullableString(model.ColDevQuiz, p.DevQuiz)
	addString(model.ColDevInterest, p.DevInterest)
	addNullableString(model.ColCMFirstEmail, p.CMFirstEmail)
	addNullableString(model.ColCMApprovalEmail, p.CMApprovalEmail)
	addBool(model.ColProspectOrganizer, p.ProspectOrganizer)
	addBool(model.ColProspectChapterMbr, p.ProspectChapterMember)
	addString(model.ColReferralFriends, p.ReferralFriends)
	addString(model.ColReferralApply, p.ReferralApply)
	addString(model.ColReferralOutlet, p.ReferralOutlet)
	addNullableString(model.ColInterestDate, p.InterestDate)
	addNullableString(model.ColNotes, p.Notes)
	addString(model.ColVisionWall, p.VisionWall)
	addBool(model.ColVotingAgreement, p.VotingAgreement)
	addString(model.ColStreetAddress, p.StreetAddress)
	addString(model.ColCity, p.City)
	addString(model.ColState, p.State)
	addInt(model.ColAssignedTo, p.AssignedTo)
	addNullableString(model.ColFollowupDate, p.FollowupDate)

	return d
}

func ActivistsSearchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	result, err := model.QueryActivists(authedUser, options, repo)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, QueryActivistResultJSON{
		Activists: model.BuildActivistJSONArray(result.Activists),
		Pagination: QueryActivistPagination{
			NextCursor: result.Pagination.NextCursor,
		},
	})
}

func ActivistsCountHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistCountOptions
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	count, err := model.CountActivists(authedUser, options, repo)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, map[string]int{"count": count})
}

// ActivistsExportHandler streams the full result set for the given query options
// as a CSV file. The CSV columns are the requested API columns in the same
// order. Rows are streamed directly from the database; response headers are
// not written until the first row arrives, so validation/auth errors still
// surface as JSON.
func ActivistsExportHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	cols := append([]model.ActivistColumnName{}, options.Shape.Columns...)
	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = string(c)
	}
	buildRow := func(a model.ActivistJSON) []string {
		return activistCSVRow(a, cols)
	}

	streamActivistsCSV(w, authedUser, options, repo, header, buildRow)
}

// ActivistsExportSpokeHandler streams a CSV in the Spoke dialer layout
// (first_name, last_name, cell) derived from name, preferred_name, and phone.
// The columns are server-controlled: clients must send an empty columns list,
// and the filters/sort in the request body are applied as usual.
func ActivistsExportSpokeHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	if len(options.Shape.Columns) != 0 {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("spoke export must be requested with an empty columns list"))
		return
	}
	options.Shape.Columns = []model.ActivistColumnName{
		model.ColName,
		model.ColPreferredName,
		model.ColPhone,
	}
	// chapter_name is required by QueryActivistShape validation when no chapter
	// filter is set, even though it doesn't appear in the spoke CSV.
	if options.Shape.Filters.ChapterId == 0 {
		options.Shape.Columns = append(options.Shape.Columns, model.ColChapterName)
	}

	streamActivistsCSV(w, authedUser, options, repo,
		[]string{"first_name", "last_name", "cell"},
		activistCSVRowSpoke,
	)
}

// streamActivistsCSV runs the query and writes header + rows as CSV. Response
// headers are deferred until the first row arrives so validation/auth errors
// can still surface as JSON.
func streamActivistsCSV(
	w http.ResponseWriter,
	authedUser model.ADBUser,
	options model.QueryActivistOptions,
	repo model.ActivistRepository,
	header []string,
	buildRow func(model.ActivistJSON) []string,
) {
	var cw *csv.Writer
	startCSV := func() error {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="activists.csv"`)
		cw = csv.NewWriter(w)
		return cw.Write(header)
	}

	err := model.StreamActivists(authedUser, options, repo, func(a model.ActivistExtra) error {
		if cw == nil {
			if err := startCSV(); err != nil {
				return fmt.Errorf("writing CSV header: %w", err)
			}
		}
		return cw.Write(buildRow(model.BuildActivistJSON(a)))
	})
	if err != nil {
		if cw == nil {
			// No bytes written yet — surface the error as a JSON response.
			if errors.Is(err, model.ErrValidation) {
				sendErrorMessage(w, http.StatusBadRequest, err)
			} else {
				sendErrorMessage(w, http.StatusInternalServerError, err)
			}
			return
		}
		cw.Flush()
		if flushErr := cw.Error(); flushErr != nil {
			log.Printf("activists CSV export: flush after stream error: %v", flushErr)
		}
		log.Printf("activists CSV export: %v", err)
		return
	}

	if cw == nil {
		// Query matched zero rows — still return a valid (header-only) CSV.
		if err := startCSV(); err != nil {
			log.Printf("activists CSV export: write header: %v", err)
			return
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("activists CSV export: flush: %v", err)
	}
}

// activistCSVRowSpoke renders a row in the Spoke dialer layout. first_name
// prefers preferred_name; otherwise it's the first whitespace-separated token
// of name. last_name is the remainder after the first space (empty if name has
// no space).
func activistCSVRowSpoke(a model.ActivistJSON) []string {
	firstName := a.PreferredName
	lastName := ""
	if i := strings.Index(a.Name, " "); i >= 0 {
		if firstName == "" {
			firstName = a.Name[:i]
		}
		lastName = a.Name[i+1:]
	} else if firstName == "" {
		firstName = a.Name
	}
	return []string{firstName, lastName, a.Phone}
}

// activistJSONFieldByJSONTag maps a json tag name to the corresponding
// reflect.StructField index on model.ActivistJSON. Built once at startup so
// each export row is a map lookup, not a struct walk.
var activistJSONFieldByJSONTag = func() map[string]int {
	m := map[string]int{}
	t := reflect.TypeOf(model.ActivistJSON{})
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name != "" && name != "-" {
			m[name] = i
		}
	}
	return m
}()

func activistCSVRow(a model.ActivistJSON, columns []model.ActivistColumnName) []string {
	v := reflect.ValueOf(a)
	out := make([]string, len(columns))
	for i, col := range columns {
		idx, ok := activistJSONFieldByJSONTag[string(col)]
		if !ok {
			continue
		}
		out[i] = formatCSVValue(v.Field(idx))
	}
	return out
}

// formatCSVValue converts v into its string representation suitable for CSV output.
// It handles string, bool, signed and unsigned integers, and floats; other kinds are
// rendered using fmt.Sprint of the underlying value.
func formatCSVValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		return fmt.Sprint(v.Interface())
	}
}

// ActivistsDebugQueryHandler accepts the same body as ActivistsSearchHandler,
// runs EXPLAIN ANALYZE on the underlying SQL, persists the resolved query and
// EXPLAIN ANALYZE output to the debug_sql_queries table, and returns the id
// ActivistsDebugQueryHandler decodes query options from the request body, executes a debug query that records the resolved SQL and EXPLAIN ANALYZE output, and responds with the inserted debug row id as JSON.
//
// It accepts the same request shape as the activists search endpoint and returns a JSON object `{"id": <row id>}` on success.
func ActivistsDebugQueryHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	id, err := model.DebugActivistQuery(authedUser, options, repo)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, map[string]int64{"id": id})
}

// ActivistPatchHandler applies a partial update to the activist identified by the `{id}` URL
// parameter and returns the updated activist as JSON.
//
// It validates the numeric activist id and the request body, applies the patch, then fetches
// and writes the updated activist as `{"activist": ...}`.
//
// Error responses:
//   - 400 Bad Request for an invalid id or malformed/invalid request body,
//   - 404 Not Found if no activist exists with the given id,
//   - 500 Internal Server Error for other failures.
func ActivistPatchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, db *sqlx.DB, repo model.ActivistRepository, userRepo model.UserRepository) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	activistID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("invalid activist id %s: %w", rawID, err))
		return
	}

	var input ActivistPatchInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	if err := model.PatchActivist(db, repo, userRepo, authedUser, activistID, input.ToPatchData()); err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else if errors.Is(err, model.ErrNotFound) {
			sendErrorMessage(w, http.StatusNotFound, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Return updated activist.
	activist, err := model.GetActivistJSONForUser(db, authedUser, model.GetActivistOptions{ID: activistID})
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, map[string]any{
		"activist": activist,
	})
}

func ActivistGetHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, db *sqlx.DB) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	activistID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("invalid activist id %s: %w", rawID, err))
		return
	}

	activist, err := model.GetActivistJSONForUser(db, authedUser, model.GetActivistOptions{ID: activistID})
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			sendErrorMessage(w, http.StatusNotFound, fmt.Errorf("no activist found with id %d", activistID))
		} else if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, map[string]any{
		"activist": activist,
	})
}
