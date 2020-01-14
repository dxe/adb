package members

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

func Route(r *mux.Router, db *sqlx.DB) {
	s := &server{db}
	r.HandleFunc("/", s.index)
	r.HandleFunc("/login", s.login)
	r.HandleFunc("/auth", s.auth)
}

type server struct {
	db *sqlx.DB
}

func (s *server) queryJSON(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	var buf []byte
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&buf); err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}

type month struct {
	Month        int     `json:"month"`
	Events       []event `json:"events"`
	Community    int     `json:"community"`     // sloppy boolean
	DirectAction int     `json:"direct_action"` // sloppy boolean
}

type event struct {
	Name         string `json:"name"`
	Date         string `json:"date"`          // "YYYY-MM-DD"
	Community    int    `json:"community"`     // sloppy boolean
	DirectAction int    `json:"direct_action"` // sloppy boolean
}

func (s *server) events(ctx context.Context, email string) ([]month, error) {
	const q = `
select coalesce(json_arrayagg(json_object(
  'month', month,
  'events', events,
  'community', community,
  'direct_action', direct_action
)), '[]') from (
  select e.month, max(e.community) as community, max(e.direct_action) as direct_action,
         json_arrayagg(json_object(
           'name', e.name,
           'date', e.date,
           'community', e.community,
           'direct_action', e.direct_action
         )) as events

  from (
    select id, name, date, extract(year_month from date) as month,
           event_type in ('Circle', 'Community', 'Training') as community,
           event_type in ('Action', 'Campaign Action', 'Frontline Surveillance', 'Outreach', 'Sanctuary') as direct_action
    from events
  ) e
  join event_attendance ea
    on (e.id = ea.event_id)
  join activists a
    on (ea.activist_id = a.id)

  where a.email = ?
  group by e.month
) x
`

	var res []month
	if err := s.queryJSON(ctx, &res, q, email); err != nil {
		return nil, err
	}

	// Sort in descending order by date. (MySQL doesn't allow us
	// to control json_arrayagg's aggregation order.)
	sort.Slice(res, func(i, j int) bool { return res[i].Month > res[j].Month })
	for k := range res {
		events := res[k].Events
		sort.Slice(events, func(i, j int) bool { return events[i].Date > events[j].Date })
	}

	return res, nil
}

// Cookie names.
const (
	membersIDToken = "members_id_token"
	membersState   = "members_state"
)

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(membersIDToken)
	var token *oidc.IDToken
	if err == nil {
		token, err = verifier.Verify(r.Context(), c.Value)
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err == nil {
		err = token.Claims(&claims)
	}
	if err != nil || !claims.EmailVerified {
		http.Redirect(w, r, absURL("/login"), http.StatusFound)
		return
	}

	evs, err := s.events(r.Context(), claims.Email)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	count := 0
	for _, ev := range evs {
		count += len(ev.Events)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = indexTmpl.Execute(w, map[string]interface{}{
		"Email":      claims.Email,
		"Count":      count,
		"Attendance": evs,
	})
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	state, err := nonce()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   membersState,
		Value:  state,
		MaxAge: 3600,
	})

	var opts []oauth2.AuthCodeOption
	force := r.URL.Query()["force"] != nil
	if force {
		// If the user is currently only signed into one
		// Google Account, we need to set
		// prompt=select_account to force the account chooser
		// dialog to appear. Otherwise, Google will just
		// redirect back to us again immediately.
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "select_account"))
	}

	http.Redirect(w, r, conf.AuthCodeURL(state, opts...), http.StatusFound)
}

func (s *server) auth(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(membersState)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	if c.Value != r.FormValue("state") {
		fmt.Fprintln(w, "state mismatch")
		return
	}

	token, err := conf.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	idToken := token.Extra("id_token").(string)
	http.SetCookie(w, &http.Cookie{
		Name:   membersIDToken,
		Value:  idToken,
		MaxAge: 3600,
	})
	http.Redirect(w, r, absURL("/"), http.StatusFound)
}

var indexTmpl = template.Must(template.New("index").Funcs(template.FuncMap{
	"monthfmt": func(n int) string { return fmt.Sprintf("%d-%02d", n/100, n%100) },
}).Parse(`
<!doctype html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css?family=Source+Sans+Pro&display=swap" rel="stylesheet">

<style>
body {
  font-family: 'Source Sans Pro', sans-serif;
}

.wrap {
  max-width: 40em;
  margin-left: auto;
  margin-right: auto;
}

table {
  padding-top: 2em;
  border-spacing: 0;
  font-variant-numeric: tabular-nums;
}

tr.month {
  background-color: #ddd;
  font-weight: bold;
}

tr.month td {
  text-align: center;
}

tr.mpi {
  background-color: #beb;
}

td {
  padding: 0.375em;
}

td:nth-child(3) {
  white-space: nowrap;
}

.green { background-color: #beb; }
.gray { background-color: #ddd; }
</style>
</head>

<body>
<div class="wrap">

<p>Hello, <b>{{.Email}}</b>! (Not you? <a href="login?force">Click here</a> to login as someone else.)</p>

<p>Below is a list of <b>{{.Count}}</b> events you've attended with DxE SF.</p>

<p>🏙️ indicates the event is a "community" event;<br>
📣 indicates a "direct action" event.</p>

<p>A <b class="green">green</b> bar indicates you met the MPI requirements for that month;<br>
a <b class="gray">gray</b> bar indicates you did not.</p>

<p>Questions or feedback? Email <a href="mailto:tech@dxe.io">tech@dxe.io</a>.</p>

<table>
{{range .Attendance}}
<tr class="month {{if and .Community .DirectAction}}mpi{{end}}">
  <td>{{if .Community}}🏙️{{end}}</td>
  <td>{{if .DirectAction}}📣{{end}}</td>
  <td colspan=2>{{monthfmt .Month}}</td>
</tr>
{{range .Events}}
<tr>
  <td>{{if .Community}}🏙️{{end}}</td>
  <td>{{if .DirectAction}}📣{{end}}</td>
  <td>{{.Date}}</td>
  <td>{{.Name}}</td>
</tr>
{{end}}
{{end}}
</table>

</div>
</body>
</html>
`))

var conf, verifier = func() (*oauth2.Config, *oidc.IDTokenVerifier) {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}
	conf := &oauth2.Config{
		ClientID:     config.MembersClientID,
		ClientSecret: config.MembersClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  absURL("/auth"),
		Scopes:       []string{"email"},
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.MembersClientID,
	})
	return conf, verifier
}()

// nonce returns a 256-bit random hex string.
func nonce() (string, error) {
	var buf [32]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func absURL(path string) string {
	if config.IsProd {
		return "https://members.dxesf.org" + path
	}
	return "http://localhost:8080/members" + path
}
