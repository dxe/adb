package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/net/context"

	oidc "github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	"github.com/urfave/negroni"
)

func flashMessageSuccess(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "flash_message_success",
		Value: message,
		Path:  "/",
	})
}

func flashMesssageError(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "flash_message_error",
		Value: message,
		Path:  "/",
	})
}

var sessionStore = sessions.NewCookieStore([]byte(config.CookieSecret))

func getAuthedADBUser(db *sqlx.DB, r *http.Request) (adbUser model.ADBUser, authed bool) {
	// First, check the cookie.
	authSession, err := sessionStore.New(r, "auth-session")
	if err != nil {
		// the cookie secret has changed
		return model.ADBUser{}, false
	}
	authed, ok := authSession.Values["authed"].(bool)
	// We should always set "authed" to true, see setAuthSession,
	// but check it just in case.
	if !ok || !authed {
		return model.ADBUser{}, false
	}
	adbUserID, ok := authSession.Values["adbuserid"].(int)
	if !ok {
		return model.ADBUser{}, false
	}

	// Then, check that the user is still authed.
	adbUser, err = model.GetADBUser(db, adbUserID, "")
	if err != nil {
		return model.ADBUser{}, false
	}
	if adbUser.Disabled {
		return model.ADBUser{}, false
	}

	return adbUser, true
}

func setAuthSession(w http.ResponseWriter, r *http.Request, adbUser model.ADBUser) error {
	if adbUser.Disabled {
		return nil
	}

	authSession, err := sessionStore.Get(r, "auth-session")
	if err != nil {
		return err
	}
	authSession.Options = &sessions.Options{
		Path: "/",
		// MaxAge is 30 days in seconds
		MaxAge: 30 * // days
			24 * // hours
			60 * // minutes
			60, // seconds
		HttpOnly: true,
	}
	authSession.Values["authed"] = true
	authSession.Values["adbuserid"] = adbUser.ID
	return sessionStore.Save(r, w, authSession)
}

func noCacheHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
}

func router() *mux.Router {
	db := model.NewDB(config.DBDataSource())
	main := MainController{db: db}

	router := mux.NewRouter()
	// Unauthed pages
	router.HandleFunc("/login", main.LoginHandler)

	// Authed pages
	router.Handle("/", alice.New(main.authMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/update_event/{event_id:[0-9]+}", alice.New(main.authMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/list_events", alice.New(main.authMiddleware).ThenFunc(main.ListEventsHandler))
	router.Handle("/list_activists", alice.New(main.authMiddleware).ThenFunc(main.ListActivistsHandler))
	router.Handle("/leaderboard", alice.New(main.authMiddleware).ThenFunc(main.LeaderboardHandler))
	router.Handle("/power", alice.New(main.authMiddleware).ThenFunc(main.PowerHandler)) // TODO: rename

	// Unauthed API
	router.HandleFunc("/tokensignin", main.TokenSignInHandler)
	router.HandleFunc(config.Route0, main.TransposedEventsDataJsonHandler)
	router.HandleFunc(config.Route1, main.PowerWallboard)      // used for showing power on wallboard at ARC
	router.HandleFunc(config.Route2, main.ActivistListHandler) // used for connections google sheet

	// Authed API
	router.Handle("/activist_names/get", alice.New(main.apiAuthMiddleware).ThenFunc(main.AutocompleteActivistsHandler))
	router.Handle("/event/save", alice.New(main.apiAuthMiddleware).ThenFunc(main.EventSaveHandler))
	router.Handle("/event/list", alice.New(main.apiAuthMiddleware).ThenFunc(main.EventListHandler))
	router.Handle("/event/delete", alice.New(main.apiAuthMiddleware).ThenFunc(main.EventDeleteHandler))
	router.Handle("/activist/list", alice.New(main.apiAuthMiddleware).ThenFunc(main.ActivistListHandler))
	router.Handle("/activist/save", alice.New(main.apiAuthMiddleware).ThenFunc(main.ActivistSaveHandler))
	router.Handle("/activist/hide", alice.New(main.apiAuthMiddleware).ThenFunc(main.ActivistHideHandler))
	router.Handle("/activist/merge", alice.New(main.apiAuthMiddleware).ThenFunc(main.ActivistMergeHandler))
	router.Handle("/leaderboard/list", alice.New(main.apiAuthMiddleware).ThenFunc(main.LeaderboardListHandler))

	if config.IsProd {
		router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		router.PathPrefix("/dist").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))
	} else {
		router.PathPrefix("/static").Handler(noCacheHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
		router.PathPrefix("/dist").Handler(noCacheHandler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist")))))
	}
	return router
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, authed := getAuthedADBUser(c.db, r)
		if !authed {
			// Delete the cookie if it doesn't auth.
			c := &http.Cookie{
				Name:     "auth-session",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			}
			http.SetCookie(w, c)

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Request is authed at this point.
		h.ServeHTTP(w, r)
	})
}

func (c MainController) apiAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, authed := getAuthedADBUser(c.db, r)
		if !authed {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		// Request is authed at this point.
		h.ServeHTTP(w, r)
	})
}

func (c MainController) TokenSignInHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	unverifiedIdToken := r.PostFormValue("idtoken")

	tokenCtx := context.Background()

	provider, err := oidc.NewProvider(tokenCtx, "https://accounts.google.com")
	if err != nil {
		panic(err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: "975059814880-lfffftbpt7fdl14cevtve8sjvh015udc.apps.googleusercontent.com",
	})

	idToken, err := verifier.Verify(tokenCtx, unverifiedIdToken)
	if err != nil {
		panic(err)
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		panic(err)
	}

	adbUser, err := model.GetADBUser(c.db, 0, claims.Email)
	if err != nil || adbUser.Disabled {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "Email is not valid",
		})
		return
	}
	// Email is valid
	setAuthSession(w, r, adbUser)
	writeJSON(w, map[string]interface{}{
		"redirect": true,
	})
}

func (c MainController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "login", PageData{PageName: "Login"})
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "event_list", PageData{PageName: "EventList"})
}

func (c MainController) ListActivistsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "activist_list", PageData{PageName: "ActivistList"})
}

func (c MainController) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "leaderboard", PageData{PageName: "Leaderboard"})
}

var templates = template.Must(template.New("").Funcs(
	template.FuncMap{
		"formatdate": func(date time.Time) string {
			return date.Format(model.EventDateLayout)
		},
		"datenotzero": func(date time.Time) bool {
			return !time.Time{}.Equal(date)
		},
	}).ParseGlob("templates/*.html"))

type PageData struct {
	PageName string
	Data     interface{}
}

// Render a page. All templates that load a header expect a PageData
// object.
func renderPage(w io.Writer, name string, pageData PageData) {
	renderTemplate(w, name, pageData)
}

// Generic function to render a template. Most of the time, you want
// to use `renderPage` instead.
func renderTemplate(w io.Writer, name string, data interface{}) {
	if err := templates.ExecuteTemplate(w, name+".html", data); err != nil {
		panic(err)
	}
}

func writeJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
}

/* Accepts a non-nil error and sends an error response */
func sendErrorMessage(w io.Writer, err error) {
	if err == nil {
		panic(errors.Wrap(err, "Cannot send error message if error is nil"))
	}
	fmt.Printf("ERROR: %+v\n", err)
	writeJSON(w, map[string]string{
		"status":  "error",
		"message": err.Error(),
	})
}

func (c MainController) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var event model.Event
	if eventIDStr, ok := vars["event_id"]; ok {
		eventID, err := strconv.Atoi(eventIDStr)
		if err != nil {
			panic(err)
		}
		event, err = model.GetEvent(c.db, model.GetEventOptions{EventID: eventID})
		if err != nil {
			panic(err)
		}
	}

	renderPage(w, "event_new", PageData{
		PageName: "NewEvent",
		Data: map[string]interface{}{
			"Event": event,
		},
	})
}

func (c MainController) TransposedEventsDataJsonHandler(w http.ResponseWriter, r *http.Request) {
	events, err := model.GetEventsJSON(c.db, model.GetEventOptions{
		OrderBy:   "e.date ASC",
		DateFrom:  "2017-01-01",
		DateTo:    "",
		EventType: "",
	})
	if err != nil {
		panic(err)
	}

	writeJSON(w, events)
}

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, r *http.Request) {
	names := model.GetAutocompleteNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) ActivistSaveHandler(w http.ResponseWriter, r *http.Request) {
	userExtra, err := model.CleanActivistData(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	activistID, err := model.UpdateActivistData(c.db, userExtra)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// Retrieve updated information from database and send in response body
	activist, err := model.GetUserJSON(c.db, model.GetUserOptions{ID: activistID})
	if err != nil {
		panic(err)
	}

	out := map[string]interface{}{
		"status":   "success",
		"activist": activist,
	}
	writeJSON(w, out)
}

func (c MainController) ActivistHideHandler(w http.ResponseWriter, r *http.Request) {
	var userID struct {
		ID int `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&userID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	err = model.HideUser(c.db, userID.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
	}
	writeJSON(w, out)
}

func (c MainController) ActivistMergeHandler(w http.ResponseWriter, r *http.Request) {
	var activistMergeData struct {
		CurrentActivistID  int    `json:"current_activist_id"`
		TargetActivistName string `json:"target_activist_name"`
	}
	err := json.NewDecoder(r.Body).Decode(&activistMergeData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// First, we need to get the activist ID for the target
	// activist.
	mergedUser, err := model.GetUser(c.db, activistMergeData.TargetActivistName)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "Could not fetch data for: %s", activistMergeData.TargetActivistName))
		return
	}

	err = model.MergeUser(c.db, activistMergeData.CurrentActivistID, mergedUser.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
	}
	writeJSON(w, out)
}

func (c MainController) EventSaveHandler(w http.ResponseWriter, r *http.Request) {
	event, err := model.CleanEventData(c.db, r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// Events with no event ID are new events.
	isNewEvent := event.ID == 0

	eventID, err := model.InsertUpdateEvent(c.db, event)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	attendees, err := model.GetEventAttendance(c.db, eventID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status":    "success",
		"redirect":  "",
		"attendees": attendees,
	}
	if isNewEvent {
		out["redirect"] = fmt.Sprintf("/update_event/%d", eventID)
	}
	writeJSON(w, out)
}

func (c MainController) EventListHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	eventName := r.PostFormValue("event_name")
	eventActivist := r.PostFormValue("event_activist")
	dateStart := r.PostFormValue("event_date_start")
	dateEnd := r.PostFormValue("event_date_end")
	eventType := r.PostFormValue("event_type")

	events, err := model.GetEventsJSON(c.db, model.GetEventOptions{
		OrderBy:        "e.date DESC, e.id DESC",
		DateFrom:       dateStart,
		DateTo:         dateEnd,
		EventType:      eventType,
		EventNameQuery: eventName,
		EventActivist:  eventActivist,
	})

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, events)
}

func (c MainController) EventDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}
	eventIDStr := r.PostFormValue("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		panic(err)
	}

	if err := model.DeleteEvent(c.db, eventID); err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"status": "success",
	})
}

func (c MainController) ActivistListHandler(w http.ResponseWriter, r *http.Request) {
	activists, err := model.GetUsersJSON(c.db, model.GetUserOptions{})
	if err != nil {
		panic(err)
	}

	writeJSON(w, activists)
}

func (c MainController) LeaderboardListHandler(w http.ResponseWriter, r *http.Request) {
	activists, err := model.GetLeaderboardUsersJSON(c.db)
	if err != nil {
		panic(err)
	}

	writeJSON(w, activists)
}

func (c MainController) PowerHandler(w http.ResponseWriter, r *http.Request) {
	power, err := model.GetPower(c.db)
	if err != nil {
		panic(err)
	}

	powerHist, err := model.GetPowerHistArray(c.db)
	if err != nil {
		panic(err)
	}

	powerMTD, err := model.GetPowerMTD(c.db)
	if err != nil {
		panic(err)
	}

	renderPage(w, "power", PageData{
		PageName: "Power",
		Data: map[string]interface{}{
			"Power":     power,
			"PowerHist": powerHist,
			"PowerMTD":  powerMTD,
		},
	})
}

func (c MainController) PowerWallboard(w http.ResponseWriter, r *http.Request) {
	power, err := model.GetPower(c.db)
	if err != nil {
		panic(err)
	}

	powerHist, err := model.GetPowerHistArray(c.db)
	if err != nil {
		panic(err)
	}

	powerMTD, err := model.GetPowerMTD(c.db)
	if err != nil {
		panic(err)
	}

	renderPage(w, "power_wallboard", PageData{
		PageName: "PowerWallboard",
		Data: map[string]interface{}{
			"Power":     power,
			"PowerHist": powerHist,
			"PowerMTD":  powerMTD,
		},
	})
}

func main() {
	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r := router()
	n.UseHandler(r)

	fmt.Println("Listening on localhost:" + config.Port)
	http.ListenAndServe(":"+config.Port, n)
}
