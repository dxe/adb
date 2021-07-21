package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dxe/adb/members"

	"github.com/dxe/adb/international_mailer"

	oidc "github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"github.com/dxe/adb/discord"
	"github.com/dxe/adb/event_sync"
	"github.com/dxe/adb/form_processor"
	"github.com/dxe/adb/google_groups_sync"
	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/dxe/adb/survey_mailer"
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

type latLng struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// getIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		// if contains ":" then split at ":" and return [0]
		return stripPort(forwarded)
	}
	return stripPort(r.RemoteAddr)
}

// removes port # from an ip string (e.g. 1.1.1.1:80 to 1.1.1.1)
func stripPort(ip string) string {
	if strings.Contains(ip, ":") {
		return strings.Split(ip, ":")[0]
	}
	return ip
}

func generateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
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

func getAuthedADBChapter(db *sqlx.DB, r *http.Request) int {
	user, authed := getAuthedADBUser(db, r)
	if !authed {
		panic("Tried getting chapter for unauthorized user.")
	}
	return user.ChapterID
}

func noCacheHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
}

func (c MainController) corsAllowGetMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET")
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

	admin := router.PathPrefix("").Subrouter()
	admin.Use(csrfMiddleware)

	// Unauthed pages
	router.HandleFunc("/login", main.LoginHandler)
	router.HandleFunc("/logout", main.LogoutHandler)
	router.HandleFunc("/apply", main.ApplicationFormHandler)
	router.HandleFunc("/interest", main.InterestFormHandler)
	router.HandleFunc("/international", main.InternationalFormHandler)
	router.HandleFunc("/international_actions/{id:[0-9]+}/{token:[a-zA-Z0-9]+}", main.InternationalActionsFormHandler)

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
	router.Handle("/list_activists", alice.New(main.authOrganizerOrNonSFBayMiddleware).ThenFunc(main.ListActivistsHandler))
	router.Handle("/community_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCommunityProspectsHandler))
	router.Handle("/community_prospects_followup", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCommunityProspectsFollowupHandler))
	router.Handle("/activist_development", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListActivistsDevelopmentHandler))
	router.Handle("/organizer_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListOrganizerProspectsHandler))
	router.Handle("/chapter_member_prospects", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListChapterMemberProspectsHandler))
	router.Handle("/chapter_member_development", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListChapterMemberDevelopmentHandler))
	router.Handle("/leaderboard", alice.New(main.authOrganizerMiddleware).ThenFunc(main.LeaderboardHandler))
	router.Handle("/list_working_groups", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListWorkingGroupsHandler))
	router.Handle("/list_circles", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListCirclesHandler))
	router.Handle("/list_geocircles", alice.New(main.authOrganizerMiddleware).ThenFunc(main.ListGeoCirclesHandler))

	// Authed Admin pages
	admin.Handle("/admin/users", alice.New(main.authAdminMiddleware).ThenFunc(main.ListUsersHandler))
	admin.Handle("/list_chapters", alice.New(main.authAdminMiddleware).ThenFunc(main.ListChaptersHandler))

	// Unauthed API (internal)
	router.HandleFunc("/tokensignin", main.TokenSignInHandler)
	router.HandleFunc("/international_actions", main.InternationalActionsFormHandler)

	// Unauthed API (public)
	router.Handle("/health", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.HealthStatusHandler))
	router.Handle("/external_events/{page_id:[0-9]+}", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.ListFBEventsHandler))
	router.Handle("/chapters/{lat:[0-9.\\-]+},{lng:[0-9.\\-]+}", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.FindNearestChaptersHandler))
	router.Handle("/regions", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.ListAllChaptersByRegion))
	router.Handle("/chapters", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.ListAllChapters))
	router.Handle("/circles", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.CircleGroupNormalListHandler)) // TODO: maybe the public endpoints should return less info
	router.Handle("/geocircles", alice.New(main.corsAllowGetMiddleware).ThenFunc(main.CircleGroupGeoListHandler)) // TODO: maybe the public endpoints should return less info

	// Authed API
	router.Handle("/activist_names/get", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.AutocompleteActivistsHandler))
	router.Handle("/activist_names/get_organizers", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.AutocompleteOrganizersHandler))
	router.Handle("/activist_names/get_chaptermembers", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.AutocompleteChapterMembersHandler))
	router.Handle("/event/get/{event_id:[0-9]+}", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventGetHandler))
	router.Handle("/event/save", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventSaveHandler))
	router.Handle("/connection/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ConnectionSaveHandler))
	router.Handle("/event/list", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventListHandler))
	router.Handle("/event/delete", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.EventDeleteHandler))
	router.Handle("/activist/list", alice.New(main.apiOrganizerOrNonSFBayAuthMiddleware).ThenFunc(main.ActivistListHandler))
	router.Handle("/activist/list_basic", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.ActivistListBasicHandler))
	router.Handle("/activist/list_range", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistInfiniteScrollHandler))
	router.Handle("/activist/save", alice.New(main.apiOrganizerOrNonSFBayAuthMiddleware).ThenFunc(main.ActivistSaveHandler))
	router.Handle("/activist/hide", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ActivistHideHandler))
	router.Handle("/activist/merge", alice.New(main.apiOrganizerOrNonSFBayAuthMiddleware).ThenFunc(main.ActivistMergeHandler))
	router.Handle("/working_group/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.WorkingGroupSaveHandler))
	router.Handle("/working_group/list", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.WorkingGroupListHandler))
	router.Handle("/working_group/delete", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.WorkingGroupDeleteHandler))
	router.Handle("/circle/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CircleGroupSaveHandler))
	router.Handle("/circle/list", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.CircleGroupListHandler))
	router.Handle("/circle/delete", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CircleGroupDeleteHandler))
	router.Handle("/interaction/save", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.InteractionSaveHandler))
	router.Handle("/interaction/list", alice.New(main.apiAttendanceAuthMiddleware).ThenFunc(main.InteractionListHandler))
	router.Handle("/interaction/delete", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.InteractionDeleteHandler))
	router.Handle("/csv/chapter_member_spoke", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.ChapterMemberSpokeCSVHandler))
	router.Handle("/csv/community_prospects_hubspot", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.CommunityProspectHubSpotCSVHandler))
	router.Handle("/csv/international_organizers", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.InternationalOrganizersCSVHandler))
	router.Handle("/user/list", alice.New(main.apiOrganizerAuthMiddleware).ThenFunc(main.UserListHandler))

	// Authed Admin API
	admin.Handle("/user/save", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UserSaveHandler))
	admin.Handle("/user/delete", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UserDeleteHandler))
	admin.Handle("/chapter/list", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.ChapterListHandler))
	admin.Handle("/chapter/delete", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.ChapterDeleteHandler))
	admin.Handle("/chapter/save", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.ChapterSaveHandler))
	admin.Handle("/users-roles/add", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UsersRolesAddHandler))
	admin.Handle("/users-roles/remove", alice.New(main.apiAdminAuthMiddleware).ThenFunc(main.UsersRolesRemoveHandler))

	// Discord API
	router.Handle("/discord/list", alice.New(main.discordBotAuthMiddleware).ThenFunc(main.DiscordListHandler))
	router.Handle("/discord/status", alice.New(main.discordBotAuthMiddleware).ThenFunc(main.DiscordStatusHandler))
	router.Handle("/discord/generate", alice.New(main.discordBotAuthMiddleware).ThenFunc(main.DiscordGenerateHandler))
	router.HandleFunc("/discord/confirm/{id:[0-9]+}/{token:[a-zA-Z0-9]+}", main.DiscordConfirmHandler)
	router.HandleFunc("/discord/confirm_new/{id:[0-9]+}/{token:[a-zA-Z0-9]+}", main.DiscordConfirmNewHandler)
	router.HandleFunc("/discord/confirm_new", main.DiscordConfirmNewHandler)
	router.Handle("/discord/get_message/{message:[a-zA-Z]+}", alice.New(main.discordBotAuthMiddleware).ThenFunc(main.DiscordGetMessageHandler))
	router.Handle("/discord/set_message/{message:[a-zA-Z]+}", alice.New(main.discordBotAuthMiddleware).ThenFunc(main.DiscordSetMessageHandler))

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
	return c.authRoleMiddleware(h, []string{"admin", "organizer", "attendance", "non-sfbay"})
}

func (c MainController) authOrganizerOrNonSFBayMiddleware(h http.Handler) http.Handler {
	return c.authRoleMiddleware(h, []string{"admin", "organizer", "non-sfbay"})
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
		if r.Role == "non-sfbay" {
			mainRole = "non-sfbay"
			break
		}

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
	return c.apiRoleMiddleware(h, []string{"admin", "organizer", "attendance", "non-sfbay"})
}

func (c MainController) apiOrganizerOrNonSFBayAuthMiddleware(h http.Handler) http.Handler {
	return c.authRoleMiddleware(h, []string{"admin", "organizer", "non-sfbay"})
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
		log.Println(
			"WARNING: failed to construct OIDC provider (no Internet connection?). Some features may not work. Error:",
			err,
		)
		return nil
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
	if err != nil {
		log.Println(err.Error())
	}

	if err != nil {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "Email is not valid",
		})
		return
	}

	if adbUser.Disabled {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "Account is disabled",
		})
		return
	}

	if adbUser.ChapterID == 0 {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "No chapter assigned",
		})
		return
	}

	if len(adbUser.Roles) == 0 {
		writeJSON(w, map[string]interface{}{
			"redirect": false,
			"message":  "No roles assigned",
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
			Description: "Everyone whose Level is Supporter whose Source is a Petition or Form (excluding Application Form and Check-in Form), has not had an interaction, and has not attended an event within the past year",
			View:        "community_prospects",
		},
	})
}

func (c MainController) ListCommunityProspectsFollowupHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "CommunityProspectsFollowup",
		Data: ActivistListData{
			Title:       "Community Prospects Follow-up",
			Description: "Everyone who is assigned to someone and has a follow-up date",
			View:        "community_prospects_followup",
		},
	})
}

func (c MainController) ListActivistsDevelopmentHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "activist_list", PageData{
		PageName: "OrganizerDevelopment",
		Data: ActivistListData{
			Title:       "Organizer Development",
			Description: "Everyone who is an Organizer",
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

func (c MainController) ListGeoCirclesHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "circles_list", PageData{PageName: "GeoCirclesList"})
}

func (c MainController) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "user_list", PageData{PageName: "UserList"})
}

func (c MainController) ListChaptersHandler(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "chapters_list", PageData{PageName: "ChaptersList"})
}

func (c MainController) ChapterListHandler(w http.ResponseWriter, r *http.Request) {
	chaps, err := model.GetAllChapters(c.db)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":   "success",
		"chapters": chaps,
	})
}

func (c MainController) ChapterSaveHandler(w http.ResponseWriter, r *http.Request) {
	chap, err := model.CleanChapterData(c.db, r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	var chapID int
	if chap.ChapterID == 0 {
		chap.EmailToken = generateToken()
		chapID, err = model.InsertChapter(c.db, chap)
	} else {
		chapID, err = model.UpdateChapter(c.db, chap)
	}
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	chapJSON, err := model.GetChapterByID(c.db, chapID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":  "success",
		"chapter": chapJSON,
	})
}

func (c MainController) ChapterDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		ID int `json:"chapter_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	err = model.DeleteChapter(c.db, requestData.ID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"status": "success",
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
		"abbrev": func(input string) string {
			length := len(input)
			if length > 16 {
				return input[0:15] + "..."
			}
			return input
		},
	}).ParseGlob("templates/*.html"))

type UserChapter struct {
	ID   int
	Name string
}

type PageData struct {
	PageName    string
	Data        interface{}
	CsrfField   string
	MainRole    string
	UserName    string
	UserEmail   string
	UserChapter UserChapter
	// Filled in by renderPage.
	StaticResourcesHash string
	// Used on International & Discord Form pages
	GooglePlacesAPIKey string
	// Used on Discord Form page - TODO: handle this differently
	DiscordUser model.DiscordUser
	// Used on Int'l Actions form page
	Chapter model.ChapterWithToken
}

// Render a page. All templates that load a header expect a PageData
// object.
func renderPage(w io.Writer, r *http.Request, name string, pageData PageData) {
	pageData.CsrfField = csrf.Token(r)
	pageData.StaticResourcesHash = config.StaticResourcesHash()
	pageData.MainRole = getUserMainRole(getUserFromContext(r.Context()))
	pageData.UserName = getUserFromContext(r.Context()).Name
	pageData.UserEmail = getUserFromContext(r.Context()).Email
	pageData.UserChapter = UserChapter{
		ID:   getUserFromContext(r.Context()).ChapterID,
		Name: getUserFromContext(r.Context()).ChapterName,
	}
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
		log.Printf("Error writing JSON! %v", err.Error())
		//panic(err)
	}
}

/* Accepts a non-nil error and sends an error response */
func sendErrorMessage(w io.Writer, err error) {
	if err == nil {
		panic(errors.Wrap(err, "Cannot send error message if error is nil"))
	}
	log.Printf("ERROR: %+v\n", err.Error())
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

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	names := model.GetAutocompleteNames(c.db, chapter)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) AutocompleteOrganizersHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	names := model.GetAutocompleteOrganizerNames(c.db, chapter)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) AutocompleteChapterMembersHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	names := model.GetAutocompleteChapterMembersNames(c.db, chapter)
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
	// get requesting user's (for logging)
	user, _ := getAuthedADBUser(c.db, r)

	activistExtra, err := model.CleanActivistData(r.Body, c.db)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// If the activist id is 0, that means they're creating a new
	// activist.
	var activistID int
	if activistExtra.ID == 0 {
		activistExtra.ChapterID = user.ChapterID
		activistID, err = model.CreateActivist(c.db, activistExtra)
	} else {
		activistID, err = model.UpdateActivistData(c.db, activistExtra, user.Email)
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
	chapter := getAuthedADBChapter(c.db, r)

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
	mergedActivist, err := model.GetActivist(c.db, activistMergeData.TargetActivistName, chapter)
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
	chapter := getAuthedADBChapter(c.db, r)

	eventID, err := strconv.Atoi(mux.Vars(r)["event_id"])
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	// We pass in the chapter ID to make sure people don't get events that don't belong to their chapter.
	event, err := model.GetEvent(c.db, model.GetEventOptions{EventID: eventID, ChapterID: chapter})
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
	chapter := getAuthedADBChapter(c.db, r)

	event, err := model.CleanEventData(c.db, r.Body, chapter)
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
	chapter := getAuthedADBChapter(c.db, r)

	event, err := model.CleanEventData(c.db, r.Body, chapter)
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
	chapter := getAuthedADBChapter(c.db, r)

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
		ChapterID:      chapter,
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

	chapter := getAuthedADBChapter(c.db, r)

	// We pass in the chapter ID to make sure people don't delete event that don't belong to their chapter.
	if err := model.DeleteEvent(c.db, eventID, chapter); err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"status": "success",
	})
}

func (c MainController) WorkingGroupSaveHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	wg, err := model.CleanWorkingGroupData(c.db, r.Body, chapter)
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

func (c MainController) CircleGroupSaveHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	cir, err := model.CleanCircleGroupData(c.db, r.Body, chapter)
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
	cirs, err := model.GetCircleGroupsJSON(c.db, 0, false)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":        "success",
		"circle_groups": cirs,
	})
}

func (c MainController) CircleGroupNormalListHandler(w http.ResponseWriter, r *http.Request) {
	cirs, err := model.GetCircleGroupsJSON(c.db, 1, true)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":        "success",
		"circle_groups": cirs,
	})
}

func (c MainController) CircleGroupGeoListHandler(w http.ResponseWriter, r *http.Request) {
	cirs, err := model.GetCircleGroupsJSON(c.db, 2, true)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":        "success",
		"circle_groups": cirs,
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

func (c MainController) ActivistListHandler(w http.ResponseWriter, r *http.Request) {
	options, err := model.CleanGetActivistOptions(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	user, _ := getAuthedADBUser(c.db, r)
	options.ChapterID = user.ChapterID

	if options.AssignedToCurrentUser {
		options.AssignedTo = user.ID
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
	chapter := getAuthedADBChapter(c.db, r)

	activists := model.GetActivistListBasicJSON(c.db, chapter)

	out := map[string]interface{}{
		"status":    "success",
		"activists": activists,
	}

	writeJSON(w, out)
}

func (c MainController) ChapterMemberSpokeCSVHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	activists, err := model.GetActivistSpokeInfo(c.db, chapter)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=chapter_members_spoke.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Transfer-Encoding", "chunked")

	writer := csv.NewWriter(w)
	err = writer.Write([]string{"first_name", "last_name", "cell"})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	for _, activist := range activists {
		err := writer.Write([]string{activist.FirstName, activist.LastName, activist.Cell})
		if err != nil {
			sendErrorMessage(w, err)
			return
		}
	}
	writer.Flush()

}

func (c MainController) CommunityProspectHubSpotCSVHandler(w http.ResponseWriter, r *http.Request) {
	chapter := getAuthedADBChapter(c.db, r)

	activists, err := model.GetCommunityProspectHubSpotInfo(c.db, chapter)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=community_prospects.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Transfer-Encoding", "chunked")

	writer := csv.NewWriter(w)
	err = writer.Write([]string{"first_name", "last_name", "email", "phone", "zip", "source", "interest_date"})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	for _, activist := range activists {
		err := writer.Write([]string{activist.FirstName, activist.LastName, activist.Email, activist.Phone, activist.Zip, activist.Source, activist.InterestDate})
		if err != nil {
			sendErrorMessage(w, err)
			return
		}
	}
	writer.Flush()

}

func (c MainController) InternationalOrganizersCSVHandler(w http.ResponseWriter, r *http.Request) {
	chapters, err := model.GetAllChapters(c.db)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=international_organizers.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Transfer-Encoding", "chunked")

	writer := csv.NewWriter(w)
	err = writer.Write([]string{"chapter", "country", "email", "mentor", "organizer_name", "organizer_email", "organizer_fb", "organizer_phone"})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	for _, chap := range chapters {
		if len(chap.Organizers) == 0 {
			err := writer.Write([]string{chap.Name, chap.Country, chap.Email, chap.Mentor, "", "", "", ""})
			if err != nil {
				sendErrorMessage(w, err)
				return
			}
		}
		for _, org := range chap.Organizers {
			err := writer.Write([]string{chap.Name, chap.Country, chap.Email, chap.Mentor, org.Name, org.Email, org.Facebook, org.Phone})
			if err != nil {
				sendErrorMessage(w, err)
				return
			}
		}
	}
	writer.Flush()
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

func (c MainController) HealthStatusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status": "healthy",
	})
}

func (c MainController) ListFBEventsHandler(w http.ResponseWriter, r *http.Request) {
	// page ID (required)
	vars := mux.Vars(r)
	var pageID int
	if pageIDStr, ok := vars["page_id"]; ok {
		var err error
		pageID, err = strconv.Atoi(pageIDStr)
		if err != nil {
			panic(err)
		}
	}

	// start & end time (optional)
	u, _ := url.Parse(r.URL.String())
	params := u.Query()

	var startTime, endTime time.Time
	var err error

	if startTimeStr, ok := params["start_time"]; ok {
		startTime, err = time.Parse(time.RFC3339, startTimeStr[0])
		if err != nil {
			writeJSON(w, map[string]string{
				"error": "start_time format incorrect",
			})
			return
		}
	}

	if endTimeStr, ok := params["end_time"]; ok {
		endTime, err = time.Parse(time.RFC3339, endTimeStr[0])
		if err != nil {
			writeJSON(w, map[string]string{
				"error": "end_time format incorrect",
			})
			return
		}
	}

	localEventsFound := false

	// run query to get local events
	events, err := model.GetExternalEvents(c.db, pageID, startTime, endTime, false)
	if err != nil {
		panic(err)
	}

	// check if any local events were returned
	if len(events) > 0 {
		localEventsFound = true
	}

	if !localEventsFound {
		// get online SF Bay + ALOA events instead
		events, err = model.GetExternalEvents(c.db, 0, startTime, endTime, true)
		if err != nil {
			panic(err)
		}
	}

	// return json
	writeJSON(w, map[string]interface{}{
		"local_events_found": localEventsFound,
		"events":             events,
	})
}

func (c MainController) FindNearestChaptersHandler(w http.ResponseWriter, r *http.Request) {
	// get lat, lng
	vars := mux.Vars(r)
	var lat float64
	var lng float64
	if latStr, ok := vars["lat"]; ok {
		var err error
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			panic(err)
		}
	}
	if lngStr, ok := vars["lng"]; ok {
		var err error
		lng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			panic(err)
		}
	}

	// if lat & lng = 0, then get location using IP address
	if lat == 0 && lng == 0 {
		if config.IPGeolocationKey == "" {
			writeJSON(w, map[string]string{
				"status":  "error",
				"message": "Geolocation API key not configured",
			})
			return
		}

		ip := getIP(r)

		path := "https://api.ipgeolocation.io/ipgeo?apiKey=" + config.IPGeolocationKey + "&ip=" + ip + "&fields=latitude,longitude"
		resp, err := http.Get(path)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Error status from IP Geolocation: %v\n", resp.Status)
			panic(resp.StatusCode)
		}
		loc := latLng{}
		err = json.NewDecoder(resp.Body).Decode(&loc)
		if err != nil {
			panic(err)
		}
		lat, err = strconv.ParseFloat(loc.Latitude, 64)
		lng, err = strconv.ParseFloat(loc.Longitude, 64)
		if err != nil {
			panic(err)
		}
	}

	// run query
	pages, err := model.FindNearestChapters(c.db, lat, lng)
	if err != nil {
		panic(err)
	}

	// return json
	writeJSON(w, pages)
}

func (c MainController) ListAllChaptersByRegion(w http.ResponseWriter, r *http.Request) {
	// run query
	pages, err := model.GetAllChaptersByRegion(c.db)
	if err != nil {
		panic(err)
	}

	// return json
	writeJSON(w, pages)
}

func (c MainController) ListAllChapters(w http.ResponseWriter, r *http.Request) {
	// run query
	chapters, err := model.GetAllChapterInfo(c.db)
	if err != nil {
		panic(err)
	}

	// return json
	writeJSON(w, chapters)
}

func (c MainController) discordBotAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		// check that discord auth matches (shared secret w/ discord bot)
		discordAuth := r.PostFormValue("auth")
		if subtle.ConstantTimeCompare([]byte(discordAuth), []byte(config.DiscordSecret)) != 1 {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// TODO: move some of the discord logic out of main.go
func (c MainController) DiscordListHandler(w http.ResponseWriter, r *http.Request) {
	// this function provides a list of all activists who have a confirmed discord id
	activists, err := model.GetActivistsWithDiscordID(c.db)
	if err != nil {
		panic(err.Error())
	}

	writeJSON(w, map[string]interface{}{
		"activists": activists,
	})
}

func (c MainController) DiscordStatusHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"status": "error parsing form",
		})
		return
	}

	discordUserID, err := strconv.Atoi(r.PostFormValue("id"))
	if err != nil {
		panic(err)
	}

	// see if a record exists for this user id in the discord_users table (pending, confirmed, or not found)
	status, err := model.GetDiscordUserStatus(c.db, discordUserID)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Discord user status:", status)

	// return the status found from database
	writeJSON(w, map[string]interface{}{
		"status": status,
	})
}

func (c MainController) DiscordGenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if config.DiscordFromEmail == "" {
		log.Println("ERROR: Discord From Email is not configured! Unable to send verification email.")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	var user model.DiscordUser
	userID, err := strconv.Atoi(r.PostFormValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	user.ID = userID
	user.Email = r.PostFormValue("email")
	user.Token = generateToken()

	// check that user isn't already confirmed
	status, err := model.GetDiscordUserStatus(c.db, user.ID)
	if err != nil {
		panic(err.Error())
	}
	if status == model.Confirmed {
		writeJSON(w, map[string]interface{}{
			"status": "already confirmed",
		})
		return
	}

	// Get activists already associated with this email.
	activists, err := model.GetActivistsByEmail(c.db, user.Email)

	// ensure there aren't multiple activist records w/ this email to avoid issues later on
	if len(activists) > 1 {
		writeJSON(w, map[string]interface{}{
			"status": "too many activists",
		})
		return
	}

	// INSERT/REPLACE into database (only allows one record per discord ID)
	err = model.InsertOrUpdateDiscordUser(c.db, user)
	if err != nil {
		panic(err.Error())
	}

	confirmPath := "confirm"
	if len(activists) == 0 {
		confirmPath = "confirm_new"
	}

	// trigger email to be sent w/ verification link
	subjectText := "Please verify your email"
	// TODO: make nicer looking email
	confirmLink := config.UrlPath + "/discord/" + confirmPath + "/" + strconv.Itoa(user.ID) + "/" + user.Token
	bodyHtml := `<p>Hello,</p><p>Please click the link below to verify your email address.</p><p><a href="` + confirmLink + `">CONFIRM</a></p><p>Cheers,<br />The DxE Discord Bot</p><br /><br /><br /><p><em>This email was sent to you by DxE to verify your email address for Discord. If you did not request this verification, please do not click the above link.</em></p>`
	err = mailer.Send(mailer.Message{
		FromName:    "DxE Discord",
		FromAddress: config.DiscordFromEmail,
		ToAddress:   user.Email,
		Subject:     subjectText,
		BodyHTML:    bodyHtml,
	})
	if err != nil {
		panic(err.Error())
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
	})
}

func (c MainController) DiscordConfirmNewHandler(w http.ResponseWriter, r *http.Request) {
	var user model.DiscordUser

	if r.Method == "GET" {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			panic(err)
		}
		user.ID = id
		user.Token = vars["token"]

		renderPage(w, r, "form_discord", PageData{
			PageName:           "FormDiscord",
			GooglePlacesAPIKey: config.GooglePlacesAPIKey,
			DiscordUser:        user,
		})
	}

	if r.Method == "POST" {
		var formData model.DiscordFormData

		err := json.NewDecoder(r.Body).Decode(&formData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(formData.ID)
		if err != nil {
			panic(err)
		}
		user.ID = id
		user.Token = formData.Token
		userName := formData.FirstName + " " + formData.LastName

		// get the email associated with the token
		formData.Email, err = model.GetEmailFromDiscordToken(c.db, user.Token)
		if err != nil {
			panic(err.Error())
		}
		if formData.Email == "" {
			writeJSON(w, map[string]interface{}{
				"status": "invalid token",
			})
			return
		}

		err = model.SubmitDiscordForm(c.db, formData)
		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		err = model.ConfirmDiscordUser(c.db, user)
		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "invalid token",
				"message": err.Error(),
			})
			return
		}

		// get user status
		status, err := model.GetDiscordUserStatus(c.db, user.ID)
		if err != nil {
			panic(err.Error())
		}
		log.Println("Discord user confirmation status:", status)

		// TODO: refactor this so there is no duplicate code
		if status == model.Confirmed {

			err = discord.UpdateNickname(user.ID, userName)
			if err != nil {
				log.Println("Error updating Discord nickname!", err)
			}

			err = discord.AddUserRole(user.ID, "New user")
			if err != nil {
				log.Println(err)
			}

			welcomeMessage := "Your email has been confirmed. I've added you to the DxE Global channels. Welcome! With that out of the way, introduce yourself in <#814302017026654228> to meet our community."
			err = discord.SendMessage(user.ID, welcomeMessage)
			if err != nil {
				log.Println("Error sending Discord welcome message!", welcomeMessage, err)
			}

			// send email to alert discord mods
			emailBody := userName + " (New User) confirmed their account on Discord.<br/>If they are already in the ADB using a different name or email, please add their Discord ID (" + strconv.Itoa(user.ID) + ") to the ADB manually."
			err = mailer.Send(mailer.Message{
				FromName:    "DxE Discord",
				FromAddress: config.DiscordFromEmail,
				ToAddress:   config.DiscordModeratorEmail,
				Subject:     "Discord user confirmed",
				BodyHTML:    emailBody,
			})
			if err != nil {
				log.Println("Error sending Discord alert email to moderators!", err)
			}

			writeJSON(w, map[string]interface{}{
				"status": "success",
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"status": "error confirming user",
		})

	}
}

func (c MainController) DiscordConfirmHandler(w http.ResponseWriter, r *http.Request) {

	var user model.DiscordUser

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(err)
	}
	user.ID = id
	user.Token = vars["token"]

	// try to confirm user
	err = model.ConfirmDiscordUser(c.db, user)
	if err != nil {
		panic(err)
	}

	// get user status
	status, err := model.GetDiscordUserStatus(c.db, user.ID)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Discord user confirmation status:", status)

	// if status = confirmed, we are good to try to update things
	if status == model.Confirmed {

		// get the email associated with the token
		user.Email, err = model.GetEmailFromDiscordToken(c.db, user.Token)
		if err != nil {
			panic(err.Error())
		}

		// lookup full activist record using the email
		activists, err := model.GetActivistsByEmail(c.db, user.Email)
		if err != nil {
			panic(err.Error())
		}
		if len(activists) < 1 {
			renderPage(w, r, "discord", PageData{
				PageName: "Error",
				Data: map[string]interface{}{
					"message": "Your email address is not associated with any activist. Please reach out to tech@dxe.io for assistance.",
				},
			})
			return
		}
		if len(activists) > 1 {
			renderPage(w, r, "discord", PageData{
				PageName: "Error",
				Data: map[string]interface{}{
					"message": "There are multiple activists associated with your email address. Please reach out to tech@dxe.io for assistance.",
				},
			})
			return
		}

		// modify the discord_id in the selected activist record
		activists[0].DiscordID = sql.NullString{String: strings.TrimSpace(strconv.Itoa(user.ID)), Valid: true}
		// save the updated activist record to the database
		_, err = model.UpdateActivistData(c.db, activists[0], "SYSTEM")
		if err != nil {
			panic(err.Error())
		}

		// set name in discord to same name as ADB
		err = discord.UpdateNickname(user.ID, activists[0].Name)
		if err != nil {
			log.Println("Error updating Discord nickname!", err)
		}

		// add roles & send welcome message
		welcomeMessage := ""
		var rolesToAdd []string

		switch activists[0].ActivistLevel {
		case "Chapter Member":
			rolesToAdd = append(rolesToAdd, "Verified")
			rolesToAdd = append(rolesToAdd, "SF Bay Chapter Member")
			welcomeMessage = "Your email has been confirmed. I've added you to the Chapter Member channels. Welcome!"
		case "Organizer":
			rolesToAdd = append(rolesToAdd, "Verified")
			rolesToAdd = append(rolesToAdd, "SF Bay Chapter Member")
			rolesToAdd = append(rolesToAdd, "Organizer")
			welcomeMessage = "Your email has been confirmed. I've added you to the Chapter Member and Organizer channels. Welcome!"
		default:
			rolesToAdd = append(rolesToAdd, "Verified")
			welcomeMessage = "Your email has been confirmed. I've added you to the Global channels. Welcome! (It seems that you are not a Chapter Member in the SF Bay Area, so I did not add you to the SF Bay channels. Please email discord-mods@dxe.io if that doesn't seem right.)"
		}

		err = discord.AddUserRoles(user.ID, rolesToAdd)
		if err != nil {
			log.Println(err)
		}

		welcomeMessage += " With that out of the way, introduce yourself in <#814302017026654228> to meet our community."

		err = discord.SendMessage(user.ID, welcomeMessage)
		if err != nil {
			log.Println("Error sending Discord welcome message!", welcomeMessage, err)
		}

		// send email to alert discord mods to add this person to working group roles
		emailBody := activists[0].Name + " (" + activists[0].ActivistLevel + ") confirmed their account on Discord. Please manually add their working group roles if needed."
		err = mailer.Send(mailer.Message{
			FromName:    "DxE Discord",
			FromAddress: config.DiscordFromEmail,
			ToAddress:   config.DiscordModeratorEmail,
			Subject:     "Discord user confirmed",
			BodyHTML:    emailBody,
		})
		if err != nil {
			log.Println("Error sending Discord alert email to moderators!", err)
		}

		// TODO: handle working groups
		// at first, maybe just see if they are in a WG & tell them to message the leader to join

		// render page saying confirmation successful & link back to discord
		renderPage(w, r, "discord", PageData{
			PageName: "Success",
			Data: map[string]interface{}{
				"message": "Your email has been confirmed.",
			},
		})

		return
	}

	// render error page
	renderPage(w, r, "discord", PageData{
		PageName: "Error",
		Data: map[string]interface{}{
			"message": "There was a problem verifying your email. Please try again or contact " + template.HTML(`<a href="mailto:`+config.DiscordModeratorEmail+`">`+config.DiscordModeratorEmail+`</a>`) + ".",
		},
	})
	return
}

func (c MainController) DiscordGetMessageHandler(w http.ResponseWriter, r *http.Request) {
	messageName := mux.Vars(r)["message"]

	message, err := model.GetDiscordMessage(c.db, messageName)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status":  "success",
		"message": message,
	}
	writeJSON(w, out)
}

func (c MainController) DiscordSetMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	userID, err := strconv.Atoi(r.PostFormValue("user"))
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	var message model.DiscordMessage
	message.Name = mux.Vars(r)["message"]
	message.Text = r.PostFormValue("text")
	message.UpdatedBy = userID

	err = model.SetDiscordMessage(c.db, message)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	out := map[string]interface{}{
		"status": "success",
	}
	writeJSON(w, out)
}

func (c MainController) ApplicationFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderPage(w, r, "form_application", PageData{
			PageName: "FormApplication",
		})
	}
	if r.Method == "POST" {

		// TODO: verify the request is coming from our frontend & not someone else

		var formData model.ApplicationFormData

		err := json.NewDecoder(r.Body).Decode(&formData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = model.SubmitApplicationForm(c.db, formData)

		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"status": "success",
		})
	}
}

func (c MainController) InterestFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderPage(w, r, "form_interest", PageData{
			PageName: "FormInterest",
		})
	}
	if r.Method == "POST" {

		// TODO: verify the request is coming from our frontend & not someone else

		var formData model.InterestFormData

		err := json.NewDecoder(r.Body).Decode(&formData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = model.SubmitInterestForm(c.db, formData)

		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"status": "success",
		})
	}
}

func (c MainController) InternationalFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderPage(w, r, "form_international", PageData{
			PageName:           "FormInternational",
			GooglePlacesAPIKey: config.GooglePlacesAPIKey,
		})
	}
	if r.Method == "POST" {

		// TODO: verify the request is coming from our frontend & not someone else

		var formData model.InternationalFormData

		err := json.NewDecoder(r.Body).Decode(&formData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = model.SubmitInternationalForm(c.db, formData)

		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"status": "success",
		})
	}
}

func (c MainController) InternationalActionsFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		token := vars["token"]
		chapID, err := strconv.Atoi(vars["id"])
		if err != nil {
			panic(err)
		}

		// TODO: check that the token and ID match a chapter in the database
		chapFromDB, err := model.GetChapterByID(c.db, chapID)
		if err != nil {
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": "Invalid chapter.",
			})
			return
		}

		if chapFromDB.EmailToken != token {
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": "Invalid token.",
			})
			return
		}

		renderPage(w, r, "form_international_actions", PageData{
			PageName: "FormInternationalActions",
			Chapter:  chapFromDB,
		})
	}

	if r.Method == "POST" {
		var formData model.InternationalActionFormData

		err := json.NewDecoder(r.Body).Decode(&formData)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = model.SubmitInternationalActionForm(c.db, formData)

		if err != nil {
			log.Println(err.Error())
			log.Println(formData)
			writeJSON(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"status": "success",
		})
	}
}

func (c MainController) InteractionSaveHandler(w http.ResponseWriter, r *http.Request) {
	var interaction model.Interaction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		sendErrorMessage(w, err)
		return
	}

	if interaction.UserID == 0 {
		// new interaction, so create it using the current user's id
		user, _ := getAuthedADBUser(c.db, r)
		interaction.UserID = user.ID
	}

	if err := model.SaveInteraction(c.db, interaction); err != nil {
		sendErrorMessage(w, err)
		return
	}

	// get the updated activist record
	activist, err := model.GetActivistJSON(c.db, model.GetActivistOptions{
		ID: interaction.ActivistID,
	})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":   "success",
		"activist": activist,
	})
}

func (c MainController) InteractionListHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		ActivistID int `json:"activist_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		sendErrorMessage(w, err)
		return
	}

	interactions, err := model.ListActivistInteractions(c.db, reqData.ActivistID)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":       "success",
		"interactions": interactions,
	})
}

func (c MainController) InteractionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		ID         int `json:"id"`
		ActivistID int `json:"activist_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		sendErrorMessage(w, err)
		return
	}

	if err := model.DeleteInteraction(c.db, reqData.ID); err != nil {
		sendErrorMessage(w, err)
		return
	}

	// get the updated activist record
	activist, err := model.GetActivistJSON(c.db, model.GetActivistOptions{
		ID: reqData.ActivistID,
	})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":   "success",
		"activist": activist,
	})
}

func main() {
	var isProdArgument = flag.Bool("prod", false, "whether to run as production")
	var logLevel = flag.Int(
		"logLevel",
		1,
		"log level (see https://github.com/rs/zerolog#leveled-logging)",
	)
	flag.Parse()
	config.SetCommandLineFlags(*isProdArgument, *logLevel)
	log.Println("IsProd =", config.IsProd)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	r, db := router()
	n.UseHandler(r)

	if config.RunBackgroundJobs {
		go google_groups_sync.StartMailingListsSync(db)
		go survey_mailer.StartSurveyMailer(db)
		go international_mailer.StartInternationalMailer(db)
		go international_mailer.StartInternationalActionFormProcessor(db)
		go event_sync.StartExternalEventSync(db)
		go form_processor.StartFormProcessor(db)
	}

	go func() {
		// Set up a second router on a different port for Members.
		membersRouter := mux.NewRouter()
		mn := negroni.New()
		mn.Use(negroni.NewRecovery())
		mn.Use(negroni.NewLogger())
		members.Route(membersRouter, db)
		mn.UseHandler(membersRouter)
		log.Println("Members webserver listening on localhost:" + config.MembersPort)
		log.Fatal(http.ListenAndServe(":"+config.MembersPort, mn))
	}()

	log.Println("Main webserver listening on localhost:" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, n))

}
