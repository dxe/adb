package members

import (
	"database/sql"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"github.com/dxe/adb/model"
)

var admins = map[string]bool{
	"antonelleracelis@gmail.com": true,
	"matthew@dempsky.org":        true,
}

func isAdmin(email string) bool {
	// TODO(mdempsky): Use adb_users instead?
	return strings.HasSuffix(email, "@directactioneverywhere.com") || admins[email]
}

func (s *server) index() {
	email, err := s.googleEmail()
	if err != nil {
		s.redirect(absURL("/login"))
		return
	}

	if isAdmin(email) {
		if q := s.r.URL.Query()["email"]; len(q) >= 1 && q[0] != "" {
			email = q[0]
		}
	}

	// MySQL doesn't have a proper boolean data type, and it's
	// json_object seems to have some arbitrary heuristics for
	// deciding when to encode a boolean expression as 0/1 vs
	// true/false.
	var data struct {
		Name          string
		Pronouns      string
		Email         string
		Phone         string
		Location      string
		Facebook      string
		Birthday      string
		ActivistLevel string

		Organizer       int // boolean
		ChapterMember   int // boolean
		VotingAgreement int // boolean

		// Past3 and Past12 are how many of the past 3 and 12
		// months, respectively, the activist fulfilled MPI
		// requirements.
		//
		// Approved6 is whether the member was an approved
		// Chapter Member for the prior 6 months.
		ThisPast3, ThisPast12, ThisApproved6 int
		NextPast3, NextPast12, NextApproved6 int
		ThisMonth, NextMonth                 string // "Month Year"

		WorkingGroups []string

		Total      int
		Attendance []struct {
			Month        int // YYYYMM
			MPI          int // boolean
			Community    int // boolean
			DirectAction int // boolean
			Events       []struct {
				Date         string // "YYYY-MM-DD"
				Name         string
				Community    int // boolean
				DirectAction int // boolean
			}
		}
	}

	// This query would be more natural if attendance could be
	// computed using a subquery like working groups, but because
	// of the two-level aggregation, we'd actually need a
	// sub-subquery; and subqueries can only access variables from
	// the immediately outer context.
	// TODO: consider not hard-coding the SF Bay chapter ID in the WHERE clause
	const q = `
select json_object(
  'Name',          x.name,
  'Pronouns',      x.pronouns,
  'Email',         x.email,
  'Phone',         x.phone,
  'Location',      x.location,
  'Facebook',      x.facebook,
  'Birthday',      x.dob,
  'ActivistLevel', x.activist_level,

  'Organizer',       x.activist_level in ('Organizer'),
  'ChapterMember',   x.activist_level in ('Chapter Member', 'Organizer'),
  'VotingAgreement', x.voting_agreement,

  'ThisMonth',     date_format(now(), '%M %Y'),
  'ThisPast3',     sum(x.mpi and x.month >= extract(year_month from date_add(now(), interval -3 month))
                             and x.month <  extract(year_month from date_add(now(), interval 0 month))),
  'ThisPast12',    sum(x.mpi and x.month >= extract(year_month from date_add(now(), interval -12 month))
                             and x.month <  extract(year_month from date_add(now(), interval 0 month))),
  'ThisApproved6', cm_approval_email < date_format(date_sub(now(), interval 6 month), '%Y-%m-01'),

  'NextMonth',     date_format(date_add(now(), interval 1 month), '%M %Y'),
  'NextPast3',     sum(x.mpi and x.month >= extract(year_month from date_add(now(), interval -2 month))
                             and x.month <  extract(year_month from date_add(now(), interval 1 month))),
  'NextPast12',    sum(x.mpi and x.month >= extract(year_month from date_add(now(), interval -11 month))
                             and x.month <  extract(year_month from date_add(now(), interval 1 month))),
  'NextApproved6', cm_approval_email < date_format(date_sub(now(), interval 5 month), '%Y-%m-01'),

  'WorkingGroups', (
    select json_arrayagg(w.name)
    from working_groups w
    join working_group_members m on (w.id = m.working_group_id)
    where m.activist_id = x.id
  ),

  'Total', sum(x.subtotal),
  'Attendance', if(sum(x.subtotal) = 0, null,
    json_arrayagg(json_object(
      'Month', x.month,
      'MPI', x.mpi,
      'Community', x.community,
      'DirectAction', x.direct_action,
      'Events', x.events
    )))
)
from (
  select a.id, a.name, a.pronouns, a.email, a.phone, a.location, a.facebook, a.activist_level, a.dob, a.cm_approval_email, a.voting_agreement,
    e.month, count(e.id) as subtotal,
    max(e.community) as community, max(e.direct_action) as direct_action,
    (max(e.direct_action) and (max(e.community) or e.month >= 202001)) as mpi,
    json_arrayagg(json_object(
      'Date', e.date,
      'Name', e.name,
      'Community', e.community,
      'DirectAction', e.direct_action
    )) as events
  from activists a
  left join event_attendance ea on (a.id = ea.activist_id)
  left join (
          select id, date,
                 concat(name, if(event_type = 'Connection', ' (Connection)', '')) as name,
                 extract(year_month from date) as month,
                 event_type in ('Circle', 'Community', 'Training') as community,
                 event_type in ('Action', 'Campaign Action', 'Frontline Surveillance', 'Outreach', 'Animal Care') as direct_action
          from events
        ) e on (e.id = ea.event_id)
  where a.email = ?
    and a.chapter_id = ` + model.SFBayChapterIdStr + `
    and not a.hidden
  group by a.id, e.month
) x
group by x.id
`

	if err := s.queryJSON(&data, q, email); err != nil {
		if err == sql.ErrNoRows {
			s.render(absentTmpl, email)
		} else {
			s.error(err)
		}
		return
	}

	// Manually sort in descending order by date, as MySQL doesn't
	// allow control of json_arrayagg()'s aggregation order.
	sort.Slice(data.Attendance, func(i, j int) bool { return data.Attendance[i].Month > data.Attendance[j].Month })
	for k := range data.Attendance {
		events := data.Attendance[k].Events
		sort.Slice(events, func(i, j int) bool { return events[i].Date > events[j].Date })
	}

	sort.Strings(data.WorkingGroups)

	s.render(indexTmpl, &data)
}

var absentTmpl = template.Must(template.New("absent").Parse(`
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
</style>
</head>

<body>
<div class="wrap">
<p>Sorry, we don't have any records associated with the email address "{{.}}".</p>
<p>You can try <a href="login?force">logging in with another account</a>
or email <a href="mailto:tech@dxe.io">tech@dxe.io</a> for help.</p>
</div>
</body>
`))

var indexTmpl = template.Must(template.New("index").Funcs(template.FuncMap{
	"monthfmt": func(n int) string { return fmt.Sprintf("%d-%02d", n/100, n%100) },
}).Parse(`
<!doctype html>
<html>
<head>
<title>DxE SF Activist Record</title>
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

h1, h2 {
  margin-top: 2em;
}

table {
  border-spacing: 0;
  font-variant-numeric: tabular-nums;
}

table.attendance {
  padding-top: 2em;
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

th, td {
  padding: 0.375em;
}

table.attendance td:nth-child(3) {
  white-space: nowrap;
}

table.profile td:nth-child(1), table.election, td:nth-child(1) {
  font-weight: bold;
}

.green { background-color: #beb; }
.gray { background-color: #ddd; }
</style>
</head>

<body>
<div class="wrap">

<h1>Activist Record</h1>

<p>Hello, <b>{{.Name}}</b>! (Not you? <a href="login?force">Click here</a> to login as someone else.)</p>

<p>Questions or feedback about this page? Email <a href="mailto:tech@dxe.io">tech@dxe.io</a>.</p>

<h2>Profile</h2>

<table class="profile">
<tr><td>Name:</td><td>{{.Name}}</td></tr>
<tr><td>Pronouns:</td><td>{{.Pronouns}}</td></tr>
<tr><td>Email:</td><td>{{.Email}}</td></tr>
<tr><td>Phone:</td><td>{{.Phone}}</td></tr>
<tr><td>Location:</td><td>{{.Location}}</td></tr>
<tr><td>Facebook Profile:</td><td><a href="{{.Facebook}}">{{.Facebook}}</a></td></tr>
<tr><td>Birthday:</td><td>{{.Birthday}}</td></tr>
<tr><td><a href="https://docs.google.com/document/d/1QnJXz8YuQeBL0cz4iK60mOvQfDN1vd7SBwvVhRFDHNc/preview">Activist Level</a>:</td><td>{{.ActivistLevel}}</td></tr>
</table>

<h2>Voter Eligibility</h2>

{{if .Organizer}}
<p>As an <a href="https://docs.dxesf.org/#33-organizers">Organizer</a>, you are eligible to vote if you've been on the Movement-Power Index for 2 of the past 3 full months or 8 of the past 12 full months at the time of the vote.</p>
<table class="elections">
<tr>
  <th>Month</th>
  <th>MPI of past 3 months</th>
  <th>MPI of past 12 months</th>
  <th>Eligible?</th>
</tr>
<tr>
  <td>{{.ThisMonth}}</td>
  <td>{{.ThisPast3}}</td>
  <td>{{.ThisPast12}}</td>
  <td>{{if or (ge .ThisPast3 2) (ge .ThisPast12 8)}}Yes{{else}}No{{end}}</td>
</tr>
<tr>
  <td>{{.NextMonth}}</td>
  <td>{{.NextPast3}}</td>
  <td>{{.NextPast12}}</td>
  <td>{{if or (ge .NextPast3 2) (ge .NextPast12 8)}}Yes{{else}}No{{end}}</td>
</tr>
</table>
{{else if .ChapterMember}}
<p>As a <a href="https://docs.dxesf.org/#32-chapter-members">Chapter Member</a>, you are eligible to vote if you've been a Chapter Member for the past 6 full months, on the Movement-Power Index for 8 of the past 12 full months at the time of the vote, and have signed the voting agreement.</p>

<table class="elections">
<tr>
  <th>Month</th>
  <th>CM for 6 months</th>
  <th>MPI of past 12 months</th>
  <th>Voting agreement</th>
  <th>Eligible?</th>
</tr>
<tr>
  <td>{{.ThisMonth}}</td>
  <td>{{if .ThisApproved6}}Yes{{else}}No{{end}}</td>
  <td>{{.ThisPast12}}</td>
  <td>{{if .VotingAgreement}}Yes{{else}}No{{end}}</td>
  <td>{{if and .ThisApproved6 (ge .ThisPast12 8) .VotingAgreement}}Yes{{else}}No{{end}}</td>
</tr>
<tr>
  <td>{{.NextMonth}}</td>
  <td>{{if .NextApproved6}}Yes{{else}}No{{end}}</td>
  <td>{{.NextPast12}}</td>
  <td>{{if .VotingAgreement}}Yes{{else}}No{{end}}</td>
  <td>{{if and .NextApproved6 (ge .NextPast12 8) .VotingAgreement}}Yes{{else}}No{{end}}</td>
</tr>
</table>
{{else}}
<p>Sorry, you are not eligible. You must be a Chapter Member to be eligible to vote.</p>
{{end}}

<h2>Working Groups</h2>

{{if .WorkingGroups}}
<ul>
{{range .WorkingGroups}}
<li>{{.}}</li>
{{end}}
</ul>
{{else}}
<p>None.</p>
{{end}}

<h2>Event Attendance</h2>

<p>Below are <b>{{.Total}}</b> events you've attended with DxE SF.</p>

<p>🏙️ indicates a "community" event;<br>
📣 indicates a "direct action" event.</p>

<p>A <b class="green">green</b> bar indicates you met MPI requirements that month;<br>
a <b class="gray">gray</b> bar indicates you did not.</p>

<table class="attendance">
{{range .Attendance}}
<tr class="month {{if .MPI}}mpi{{end}}">
  <td>{{if .Community}}🏙️{{else if (ge .Month 202001)}}🆓{{end}}</td>
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
