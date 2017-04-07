package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/directactioneverywhere/adb/model"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	"github.com/urfave/negroni"
)

var isProd bool

func init() {
	prod := flag.Bool("prod", false, "Run in production mode")
	flag.Parse()
	isProd = *prod
}

var sessionStore = sessions.NewCookieStore([]byte("replace-with-real-auth-secret"))

func isAuthed(req *http.Request) bool {
	authSession, err := sessionStore.Get(req, "auth-session")
	if err != nil {
		panic(err)
	}
	authed, ok := authSession.Values["authed"].(bool)
	// We should always set "authed" to true, see setAuthSession.
	return ok && authed
}

func setAuthSession(w http.ResponseWriter, req *http.Request, email string) error {
	authSession, err := sessionStore.Get(req, "auth-session")
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
	authSession.Values["email"] = email
	return sessionStore.Save(req, w, authSession)
}

func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !isAuthed(req) {
			http.Redirect(w, req, "/login", http.StatusFound)
			return
		}
		// Request is authed at this point.
		h.ServeHTTP(w, req)
	})
}

func apiAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !isAuthed(req) {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		// Request is authed at this point.
		h.ServeHTTP(w, req)
	})
}

var validEmails = map[string]bool{
	"samer@directactioneverywhere.com": true,
	"jake@directactioneverywhere.com":  true,
	"jakehobbs@gmail.com":              true,
	"kowshik.sundararajan@gmail.com":   true,
	"rydermeehan@gmail.com":            true,
	"jeffdavidson53@gmail.com":         true,
	"scott.r.paterson@gmail.com":       true,
	"wayne@directactioneverywhere.com": true,
	"nosefrog@gmail.com":               true,
	"samer@dropbox.com":                true,
	"matthew@dempsky.org":              true,
}

// TODO: Make this read from the database instead.
func isValidEmail(email string) bool {
	_, ok := validEmails[email]
	return ok
}

var devDataSource = "adb_user:adbpassword@/adb_db?parseTime=true"
var prodDataSource = "dxe_adb_go:L!oQ{JXlq82Nw-GqX:f4@/adb2?parseTime=true"

func router() *mux.Router {
	var db *sqlx.DB
	if isProd {
		db = model.NewDB(prodDataSource)
	} else {
		db = model.NewDB(devDataSource)
	}
	main := MainController{db: db}

	router := mux.NewRouter()
	// Unauthed pages
	router.HandleFunc("/login", main.LoginHandler)

	// Authed paged
	router.Handle("/", alice.New(authMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/update_event/{event_id:[0-9]+}", alice.New(authMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/list_events", alice.New(authMiddleware).ThenFunc(main.ListEventsHandler))
	router.Handle("/transposed_events_data", alice.New(authMiddleware).ThenFunc(main.TransposedEventsDataHandler))

	// Unauthed API
	router.HandleFunc("/tokensignin", main.TokenSignInHandler)

	// Authed API
	router.Handle("/activist_names/get", alice.New(apiAuthMiddleware).ThenFunc(main.AutocompleteActivistsHandler))
	router.Handle("/event/save", alice.New(apiAuthMiddleware).ThenFunc(main.EventSaveHandler))

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) TokenSignInHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		panic(err)
	}

	unverifiedIdToken := req.PostFormValue("idtoken")

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
	if !isValidEmail(claims.Email) {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "Email is not valid",
		})
		return
	}
	// Email is valid
	setAuthSession(w, req, claims.Email)
	writeJSON(w, map[string]interface{}{
		"redirect": true,
	})
}

func (c MainController) LoginHandler(w http.ResponseWriter, req *http.Request) {
	renderTemplate(w, "login", nil)
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, req *http.Request) {
	events, err := model.GetEvents(c.db)
	if err != nil {
		panic(err)
	}
	renderTemplate(w, "event_list", map[string]interface{}{
		"Events": events,
	})
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

func (c MainController) UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var event model.Event
	if eventIDStr, ok := vars["event_id"]; ok {
		eventID, err := strconv.Atoi(eventIDStr)
		if err != nil {
			panic(err)
		}
		event, err = model.GetEvent(c.db, eventID)
		if err != nil {
			panic(err)
		}
	}

	renderTemplate(w, "event_new", map[string]interface{}{
		"Event": event,
	})
}

func (c MainController) TransposedEventsDataHandler(w http.ResponseWriter, req *http.Request) {
	events, err := model.GetEvents(c.db)
	if err != nil {
		panic(err)
	}
	renderTemplate(w, "transposed_events_data", map[string]interface{}{
		"Events": events,
	})
}

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	names := model.GetAutocompleteNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) EventSaveHandler(w http.ResponseWriter, req *http.Request) {
	event, err := model.CleanEventData(c.db, req.Body)
	if err != nil {
		writeJSON(w, map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Events with no event ID are new events.
	isNewEvent := event.ID == 0

	eventID, err := model.InsertUpdateEvent(c.db, event)
	if err != nil {
		writeJSON(w, map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	out := map[string]string{
		"status":   "success",
		"redirect": "",
	}
	if isNewEvent {
		out["redirect"] = fmt.Sprintf("/update_event/%d", eventID)
	}
	writeJSON(w, out)
}

func main() {
	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r := router()
	n.UseHandler(r)

	var port string
	if isProd {
		port = "6060"
	} else {
		port = "8080"
	}

	fmt.Println("Listening on localhost:" + port)
	http.ListenAndServe(":"+port, n)
}
