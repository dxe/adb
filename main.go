package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"github.com/dxe/adb/facebook_events"
	"github.com/dxe/adb/mailinglist_sync"
	"github.com/dxe/adb/members"
	"github.com/dxe/adb/model"
	"github.com/dxe/adb/survey_mailer"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

func flashMessageSuccess(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash_message_success",
		Value:    message,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func flashMesssageError(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash_message_error",
		Value:    message,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

var sessionStore = sessions.NewCookieStore([]byte(config.CookieSecret))

func getAuthedADBUser(db *sqlx.DB, r *http.Request) (adbUser model.ADBUser, authed bool) {
	// In dev, just return the test user.
	if !config.IsProd {
		return model.DevTestUser, true
	}

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

func router() (*mux.Router, *sqlx.DB) {
	db := model.NewDB(config.DBDataSource())
	main := MainController{db: db}
	csrfMiddleware := csrf.Protect(
		[]byte(config.CsrfAuthKey),
		csrf.Secure(config.IsProd), // disable secure flag in dev
		csrf.Path("/"),
	)

	router := mux.NewRouter()
	members.Route(router.PathPrefix("/members").Subrouter(), db)

	admin := router.PathPrefix("").Subrouter()
	admin.Use(csrfMiddleware)

	// Unauthed pages
	router.HandleFunc("/login", main.LoginHandler)
	router.HandleFunc("/logout", main.LogoutHandler)

	// Error pages
	router.HandleFunc("/403", main.ForbiddenHandler)

	// Authed pages
	router.Handle("/", alice.New(main.authAttendanceMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/update_event/{event_id:[0-9]+}", alice.New(main.authAttendanceMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/new_connection", alice.New(main.authOrganizerMiddleware).ThenFunc(main.UpdateConnectionHandler))
	router.Handle("/update_connection/{event_id:[0-9]+}", alice.New(main.authOrganizerMiddleware).ThenFunc(main.UpdateConnectionHandler))
	router.Handle("/update_event/{event_id:[0-9]+}", alice.New(main.authAttendanceMiddleware).ThenFunc(main.UpdateEventHandler))
	router.Handle("/list_events", alice.New(main.authAttendanceMiddleware).ThenFunc(main.ListEventsHandler))
	router.Handle("/list_connections", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListConnectionsHandler))
	router.Handle("/list_activists", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsHandler))
	router.Handle("/community_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCommunityProspectsHandler))
	router.Handle("/activist_pool", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsPoolHandler))
	router.Handle("/activist_recruitment", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsRecruitmentHandler))
	router.Handle("/activist_actionteam", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsActionTeamHandler))
	router.Handle("/activist_development", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsDevelopmentHandler))
	router.Handle("/organizer_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListOrganizerProspectsHandler))
	router.Handle("/senior_organizer_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListSeniorOrganizerProspectsHandler))
	router.Handle("/senior_organizer_development", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListSeniorOrganizerDevelopmentHandler))
	router.Handle("/chapter_member_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListChapterMemberProspectsHandler))
	router.Handle("/chapter_member_development", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListChapterMemberDevelopmentHandler))
	router.Handle("/circle_member_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCircleMemberProspectsHandler))
	router.Handle("/circle_members", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCircleMembersHandler))
	router.Handle("/leaderboard", alice.New(main.authOrganizerMiddleware).ThenFunc(main.LeaderboardHandler))
	router.Handle("/list_working_groups", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListWorkingGroupsHandler))
	router.Handle("/list_circles", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCirclesHandler))

	// Authed Admin pages
	admin.Handle("/admin/users", alice.New(main.authAdminMiddleware).ThenFunc(main.ListUsersHandler))

	// Unauthed API
	router.HandleFunc("/tokensignin", main.TokenSignInHandler)
	router.HandleFunc("/fb_events/{page_id:[0-9]+}", main.ListFBEventsHandler)

	// Defunct Unauthed API
	//router.HandleFunc(config.Route0, main.TransposedEventsDataJsonHandler)
	//router.HandleFunc("/wallboard_mpi", main.newPowerWallboard)                    // new endpoint for arc tv to get mpi
	//router.HandleFunc("/wallboard_chaptermembers", main.newChapterMemberWallboard) // new endpoint for arc tv to get chapter members
	//router.HandleFunc(config.Route2, main.ActivistListHandler)                     // used for connections google sheet

	// Authed API
	router.Handle("/activist_names/get", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.AutocompleteActivistsHandler))
	router.Handle("/activist_names/get_organizers", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.AutocompleteOrganizersHandler))
	router.Handle("/event/get/{event_id:[0-9]+}", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventGetHandler))
	router.Handle("/event/save", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventSaveHandler))
	router.Handle("/connection/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ConnectionSaveHandler))
	router.Handle("/event/list", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventListHandler))
	router.Handle("/event/delete", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventDeleteHandler))
	router.Handle("/activist/list", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistListHandler))
	router.Handle("/activist/list_basic", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.ActivistListBasicHandler))
	router.Handle("/activist/list_range", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistInfiniteScrollHandler))
	router.Handle("/activist/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistSaveHandler))
	router.Handle("/activist/hide", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistHideHandler))
	router.Handle("/activist/merge", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistMergeHandler))
	router.Handle("/working_group/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.WorkingGroupSaveHandler))
	router.Handle("/working_group/list", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.WorkingGroupListHandler))
	router.Handle("/working_group/delete", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.WorkingGroupDeleteHandler))
	router.Handle("/circle/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CircleGroupSaveHandler))
	router.Handle("/circle/list", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CircleGroupListHandler))
	router.Handle("/circle/delete", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CircleGroupDeleteHandler))

	// Authed Admin API
	admin.Handle("/user/list", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UserListHandler))
	admin.Handle("/user/save", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UserSaveHandler))
	admin.Handle("/user/delete", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UserDeleteHandler))
	// Authed Admin API for managing Users Roles
	admin.Handle("/users-roles/add", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UsersRolesAddHandler))
	admin.Handle("/users-roles/remove", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UsersRolesRemoveHandler))

	// Pprof debug routes
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	if config.IsProd {
		router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		router.PathPrefix("/dist").Handler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))
	} else {
		router.PathPrefix("/static").Handler(noCacheHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
		router.PathPrefix("/dist").Handler(noCacheHandler(http.StripPrefix("/dist/", http.FileServer(http.Dir("dist")))))
	}
	return router, db
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) authRoleMiddleware(h http.Handler, allowedRoles []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, authed := getAuthedADBUser(c.db, r)
		if !authed {
			// Delete the cookie if it doesn't auth.
			c := &http.Cookie{
				Name:     "auth-session",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, c)

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if !userIsAllowed(allowedRoles, user) {
			http.Redirect(w, r.WithContext(setUserContext(r, user)), "/403", http.StatusFound)
			return
		}

		// Request is authed at this point.
		h.ServeHTTP(w, r.WithContext(setUserContext(r, user)))
	})
}

func (c MainController) authAttendanceMiddleware(h http.Handler) http.Handler {
	return c.authRoleMiddleware(h, []string{"admin", "organizer", "attendance"})
}

func (c MainController) authOrganizerMiddleware(h http.Handler) http.Handler {
	return c.authRoleMiddleware(h, []string{"admin", "organizer"})
}

func (c MainController) authAdminMiddleware(h http.Handler) http.Handler {
	return c.authRoleMiddleware(h, []string{"admin"})
}

func userIsAllowed(roles []string, user model.ADBUser) bool {

	for i := 0; i < len(roles); i++ {
		for _, r := range user.Roles {
			if r.Role == roles[i] {
				return true
			}
		}
	}

	return false
}

func getUserMainRole(user model.ADBUser) string {
	if len(user.Roles) == 0 {
		return ""
	}

	var mainRole string
	for _, r := range user.Roles {
		if r.Role == "admin" {
			mainRole = "admin"
			break
		}

		if r.Role == "organizer" {
			mainRole = "organizer"
		}

		if r.Role == "attendance" && mainRole != "organizer" {
			mainRole = "attendance"
		}
	}

	return mainRole
}

func getUserName(user model.ADBUser) string {

	var userName string

	userName = user.Name

	return userName
}

func getUserEmail(user model.ADBUser) string {

	var userEmail string

	userEmail = user.Email

	return userEmail
}

func (c MainController) apiRoleMiddleware(h http.Handler, allowedRoles []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, authed := getAuthedADBUser(c.db, r)
		if !authed {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		if !userIsAllowed(allowedRoles, user) {
			http.Error(w, http.StatusText(403), 403)
			return
		}

		// Request is authed at this point.
		h.ServeHTTP(w, r)
	})
}

func (c MainController) apiAttendanceAuthMiddleware(h http.Handler) http.Handler {
	return c.apiRoleMiddleware(h, []string{"admin", "organizer", "attendance"})
}

func (c MainController) apiOrganizerAuthMiddleware(h http.Handler) http.Handler {
	return c.apiRoleMiddleware(h, []string{"admin", "organizer"})
}

func (c MainController) apiAdminAuthMiddleware(h http.Handler) http.Handler {
	return c.apiRoleMiddleware(h, []string{"admin"})
}

func setUserContext(r *http.Request, user model.ADBUser) context.Context {
	return context.WithValue(r.Context(), "UserContext", user)
}

func getUserFromContext(ctx context.Context) model.ADBUser {
	var userctx interface{}
	userctx = ctx.Value("UserContext")

	if userctx == nil {
		return model.ADBUser{}
	}

	return userctx.(model.ADBUser)
}

var verifier = func() *oidc.IDTokenVerifier {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}
	return provider.Verifier(&oidc.Config{
		ClientID: "975059814880-lfffftbpt7fdl14cevtve8sjvh015udc.apps.googleusercontent.com",
	})
}()

func (c MainController) TokenSignInHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	idToken, err := verifier.Verify(r.Context(), r.PostFormValue("idtoken"))
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
	renderPage(w, r, "login", PageData{PageName: "Login"})
}

func (c MainController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "auth-session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	renderPage(w, r, "logout", PageData{PageName: "Logout"})
}

func (c MainController) ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "403", PageData{PageName: "403 - Forbidden"})
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "event_list", PageData{PageName: "EventList"})
}

func (c MainController) ListConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "connection_list", PageData{PageName: "ConnectionsList"})
}

type ActivistListData struct {
	Title       string
	Description string
	View        string
}

func (c MainController) ListActivistsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ActivistList",
		Data: ActivistListData{
			Title:       "All Activists",
			Description: "Everyone who has attended an event within the filtered range",
			View:        "all_activists",
		},
	})
}

func (c MainController) ListCommunityProspectsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "CommunityProspects",
		Data: ActivistListData{
			Title:       "Community Prospects",
			Description: "Everyone whose Level is Supporter or Circle Member, whose Source is a Form (other than the Circle Interest Form), Application, Fur Ban, or Petition that was submitted within the last 3 months",
			View:        "community_prospects",
		},
	})
}

func (c MainController) ListActivistsPoolHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ActivistPool",
		Data: ActivistListData{
			Title:       "Recruitment Connections",
			Description: "Inactive page",
			View:        "activist_pool",
		},
	})
}

func (c MainController) ListActivistsRecruitmentHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ActivistRecruitment",
		Data: ActivistListData{
			Title:       "Activist Recruitment",
			Description: "Inactive page",
			View:        "activist_recruitment",
		},
	})
}

func (c MainController) ListActivistsActionTeamHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ActivistActionTeam",
		Data: ActivistListData{
			Title:       "Action Team",
			Description: "Inactive page",
			View:        "action_team",
		},
	})
}

func (c MainController) ListActivistsDevelopmentHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "OrganizerDevelopment",
		Data: ActivistListData{
			Title:       "Organizer Development",
			Description: "Everyone who is an Organizer (including Senior Organizers)",
			View:        "development",
		},
	})
}

func (c MainController) ListChapterMemberDevelopmentHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ChapterMemberDevelopment",
		Data: ActivistListData{
			Title:       "Chapter Members",
			Description: "Everyone who is a Chapter Member (including Organizers)",
			View:        "chapter_member_development",
		},
	})
}

func (c MainController) ListOrganizerProspectsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "OrganizerProspects",
		Data: ActivistListData{
			Title:       "Organizer Prospects",
			Description: "Everyone who is a Prospective Organizer who is not an Organizer",
			View:        "organizer_prospects",
		},
	})
}

func (c MainController) ListSeniorOrganizerProspectsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "SeniorOrganizerProspects",
		Data: ActivistListData{
			Title:       "Senior Organizer Prospects",
			Description: "Everyone who is a Prospective Senior Organizer who is not a Senior Organizer",
			View:        "senior_organizer_prospects",
		},
	})
}

func (c MainController) ListSeniorOrganizerDevelopmentHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "SeniorOrganizerDevelopment",
		Data: ActivistListData{
			Title:       "Senior Organizer Development",
			Description: "Everyone who is a Senior Organizer",
			View:        "senior_organizer_development",
		},
	})
}

func (c MainController) ListChapterMemberProspectsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "ChapterMemberProspects",
		Data: ActivistListData{
			Title:       "Chapter Member Prospects",
			Description: "Everyone who is a Chapter Member Prospect who is not a Chapter Member or Organizer",
			View:        "chapter_member_prospects",
		},
	})
}

func (c MainController) ListCircleMemberProspectsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "CircleMemberProspects",
		Data: ActivistListData{
			Title:       "Circle Member Prospects",
			Description: "Everyone interested in joining a circle who is not assigned to a circle",
			View:        "circle_member_prospects",
		},
	})
}

func (c MainController) ListCircleMembersHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "CircleMembers",
		Data: ActivistListData{
			Title:       "Circle Members",
			Description: "Everyone who is assigned to a circle",
			View:        "circle_members",
		},
	})
}

func (c MainController) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "Leaderboard",
		Data: ActivistListData{
			Title:       "Leaderboard",
			Description: "Everyone who has attended an event in the last 30 days",
			View:        "leaderboard",
		},
	})
}

func (c MainController) ListWorkingGroupsHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "working_group_list", PageData{PageName: "WorkingGroupList"})
}

func (c MainController) ListCirclesHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "circles_list", PageData{PageName: "CirclesList"})
}

func (c MainController) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "user_list", PageData{PageName: "UserList"})
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
	PageName  string
	Data      interface{}
	CsrfField string
	MainRole  string
	UserName  string
	UserEmail string
	// Filled in by renderPage.
	StaticResourcesHash string
}

// Render a page. All templates that load a header expect a PageData
// object.
func renderPage(w io.Writer, r *http.Request, name string, pageData PageData) {
	pageData.CsrfField = csrf.Token(r)
	pageData.StaticResourcesHash = config.StaticResourcesHash()
	pageData.MainRole = getUserMainRole(getUserFromContext(r.Context()))
	pageData.UserName = getUserName(getUserFromContext(r.Context()))
	pageData.UserEmail = getUserEmail(getUserFromContext(r.Context()))
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
	var eventID int
	if eventIDStr, ok := vars["event_id"]; ok {
		var err error
		eventID, err = strconv.Atoi(eventIDStr)
		if err != nil {
			panic(err)
		}
	}

	renderPage(w, r, "event_new", PageData{
		PageName: "NewEvent",
		Data: map[string]interface{}{
			"EventID": eventID,
		},
	})
}

func (c MainController) UpdateConnectionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var eventID int
	if eventIDStr, ok := vars["event_id"]; ok {
		var err error
		eventID, err = strconv.Atoi(eventIDStr)
		if err != nil {
			panic(err)
		}
	}

	renderPage(w, r, "connection_new", PageData{
		PageName: "NewConnection",
		Data: map[string]interface{}{
			"EventID": eventID,
		},
	})
}

// DEFUNCT
// func (c MainController) TransposedEventsDataJsonHandler(w http.ResponseWriter, r *http.Request) {
// 	events, err := model.GetEventsJSON(c.db, model.GetEventOptions{
// 		OrderBy:   "e.date ASC",
// 		DateFrom:  "2017-01-01",
// 		DateTo:    "",
// 		EventType: "",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	writeJSON(w, events)
// }

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, r *http.Request) {
	names := model.GetAutocompleteNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) AutocompleteOrganizersHandler(w http.ResponseWriter, r *http.Request) {
	names := model.GetAutocompleteOrganizerNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

// TODO Protect against non POST requests. Perhaps we can do this with the router...
func (c MainController) ActivistInfiniteScrollHandler(w http.ResponseWriter, r *http.Request) {
	activistOptions, err := model.GetActivistRangeOptions(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	activists, err := model.GetActivistRangeJSON(c.db, activistOptions)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":              "success",
		"activist_range_list": activists,
	})
}

func (c MainController) ActivistSaveHandler(w http.ResponseWriter, r *http.Request) {
	activistExtra, err := model.CleanActivistData(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// If the activist id is 0, that means they're creating a new
	// activist.
	var activistID int
	if activistExtra.ID == 0 {
		activistID, err = model.CreateActivist(c.db, activistExtra)
	} else {
		activistID, err = model.UpdateActivistData(c.db, activistExtra)
	}
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// Retrieve updated information from database and send in response body
	activist, err := model.GetActivistJSON(c.db, model.GetActivistOptions{ID: activistID})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status":   "success",
		"activist": activist,
	}
	writeJSON(w, out)
}

func (c MainController) ActivistHideHandler(w http.ResponseWriter, r *http.Request) {
	var activistID struct {
		ID int `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&activistID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	err = model.HideActivist(c.db, activistID.ID)
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
	mergedActivist, err := model.GetActivist(c.db, activistMergeData.TargetActivistName)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "Could not fetch data for: %s", activistMergeData.TargetActivistName))
		return
	}

	err = model.MergeActivist(c.db, activistMergeData.CurrentActivistID, mergedActivist.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
	}
	writeJSON(w, out)
}

func (c MainController) EventGetHandler(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(mux.Vars(r)["event_id"])
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	event, err := model.GetEvent(c.db, model.GetEventOptions{EventID: eventID})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
		"event":  event.ToJSON(),
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

func (c MainController) ConnectionSaveHandler(w http.ResponseWriter, r *http.Request) {
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
		out["redirect"] = fmt.Sprintf("/update_connection/%d", eventID)
	}
	writeJSON(w, out)
}

func (c MainController) EventListHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sendErrorMessage(w, err)
		return
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

func (c MainController) WorkingGroupSaveHandler(w http.ResponseWriter, r *http.Request) {
	wg, err := model.CleanWorkingGroupData(c.db, r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	var wgID int
	if wg.ID == 0 {
		wgID, err = model.CreateWorkingGroup(c.db, wg)
	} else {
		wgID, err = model.UpdateWorkingGroup(c.db, wg)
	}
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	wgJSON, err := model.GetWorkingGroupJSON(c.db, wgID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":        "success",
		"working_group": wgJSON,
	})
}

func (c MainController) WorkingGroupListHandler(w http.ResponseWriter, r *http.Request) {
	wgs, err := model.GetWorkingGroupsJSON(c.db)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":         "success",
		"working_groups": wgs,
	})
}

func (c MainController) WorkingGroupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		ID int `json:"working_group_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	err = model.DeleteWorkingGroup(c.db, requestData.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"status": "success",
	})
}

//start circle
func (c MainController) CircleGroupSaveHandler(w http.ResponseWriter, r *http.Request) {
	cir, err := model.CleanCircleGroupData(c.db, r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	var cirID int
	if cir.ID == 0 {
		cirID, err = model.CreateCircleGroup(c.db, cir)
	} else {
		cirID, err = model.UpdateCircleGroup(c.db, cir)
	}
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	cirJSON, err := model.GetCircleGroupJSON(c.db, cirID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"circle": cirJSON,
	})
}

func (c MainController) CircleGroupListHandler(w http.ResponseWriter, r *http.Request) {
	cirs, err := model.GetCircleGroupsJSON(c.db)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":         "success",
		"working_groups": cirs,
	})
}

func (c MainController) CircleGroupDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		ID int `json:"circle_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	err = model.DeleteCircleGroup(c.db, requestData.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"status": "success",
	})
}

//end circle

func (c MainController) ActivistListHandler(w http.ResponseWriter, r *http.Request) {
	options, err := model.CleanGetActivistOptions(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	activists, err := model.GetActivistsJSON(c.db, options)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":        "success",
		"activist_list": activists,
	})
}

func (c MainController) ActivistListBasicHandler(w http.ResponseWriter, r *http.Request) {
	activists := model.GetActivistListBasicJSON(c.db)

	out := map[string]interface{}{
		"status":    "success",
		"activists": activists,
	}

	writeJSON(w, out)
}

func (c MainController) UserListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := model.GetUsersJSON(c.db)

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, users)
}

func (c MainController) UserSaveHandler(w http.ResponseWriter, r *http.Request) {
	user, err := model.CleanUserData(r.Body)

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// Check if we're updating an existing User or creating one

	var userID int
	if user.ID == 0 {
		// new user
		userID, err = model.CreateUser(c.db, user)
	} else {
		userID, err = model.UpdateUser(c.db, user)
	}

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// Retrieve updated User Data and send back in response

	userJSON, err := model.GetUserJSON(c.db, model.GetUserOptions{ID: userID})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
		"user":   userJSON,
	}

	writeJSON(w, out)
}

func (c MainController) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, err := model.CleanUserData(r.Body)

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	userID, err := model.RemoveUser(c.db, user.ID)

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
		"userID": userID,
	}

	writeJSON(w, out)
}

// DEFUNCT
// func (c MainController) newPowerWallboard(w http.ResponseWriter, r *http.Request) {
// 	power, err := model.GetPower(c.db)
// 	if err != nil {
// 		panic(err)
// 	}

// 	writeJSON(w, map[string]interface{}{
// 		"status": "success",
// 		"Power":  power,
// 	})
// }

// func (c MainController) newChapterMemberWallboard(w http.ResponseWriter, r *http.Request) {
// 	members, err := model.GetActiveChapterMembers(c.db)
// 	if err != nil {
// 		panic(err)
// 	}

// 	writeJSON(w, map[string]interface{}{
// 		"status":  "success",
// 		"Members": members,
// 	})
// }

func (c MainController) UsersRolesAddHandler(w http.ResponseWriter, r *http.Request) {
	var userRoleData struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}

	err := json.NewDecoder(r.Body).Decode(&userRoleData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	userRole := model.UserRole{
		UserID: userRoleData.UserID,
		Role:   userRoleData.Role,
	}

	userId, err := model.CreateUserRole(c.db, userRole)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":  "success",
		"user_id": userId,
	})
}

func (c MainController) UsersRolesRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var userRoleData struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}

	err := json.NewDecoder(r.Body).Decode(&userRoleData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	userRole := model.UserRole{
		UserID: userRoleData.UserID,
		Role:   userRoleData.Role,
	}

	userId, err := model.RemoveUserRole(c.db, userRole)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":  "success",
		"user_id": userId,
	})
}

func (c MainController) ListFBEventsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var pageID int
	if pageIDStr, ok := vars["page_id"]; ok {
		var err error
		pageID, err = strconv.Atoi(pageIDStr)
		if err != nil {
			panic(err)
		}
	}

	events, err := model.GetFacebookEvents(c.db, pageID)
	if err != nil {
		panic(err)
	}

	writeJSON(w, events)
}

// func (c MainController) newChapterMemberWallboard(w http.ResponseWriter, r *http.Request) {
// 	members, err := model.GetActiveChapterMembers(c.db)
// 	if err != nil {
// 		panic(err)
// 	}

// 	writeJSON(w, map[string]interface{}{
// 		"status":  "success",
// 		"Members": members,
// 	})
// }

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://dc89e0cef6204791a1f199564aec911c@sentry.io/1820804",
	})

	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r, db := router()

	// Start syncing mailing lists in the background if we have
	// the environment set up.
	if config.SyncMailingListsConfigFile != "" {
		go mailinglist_sync.StartMailingListsSync(db)
	}

	// Start running survey mailer in the background if we have
	// the environment set up.
	if config.SurveyMissingEmail != "" && config.SurveyFromEmail != "" && config.AWSAccessKey != "" && config.AWSSecretKey != "" && config.AWSSESEndpoint != "" {
		go survey_mailer.StartSurveyMailer(db)
	}

	// Start syncing Facebook events
	go facebook_events.StartFacebookSync(db)

	// Set up server
	n.UseHandler(r)

	fmt.Println("IsProd =", config.IsProd)
	fmt.Println("Listening on localhost:" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, n))
}
