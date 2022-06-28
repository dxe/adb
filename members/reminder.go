package members

import (
	"errors"
	"html/template"
	"sort"
	"time"
)

func (s *server) reminder() {
	email, err := s.googleEmail()
	if err != nil {
		s.loginDest("/reminder")
		return
	}

	if !isAdmin(email) {
		s.error(errors.New("unauthorized"))
		return
	}

	year, month, _ := time.Now().Date()
	month++ // next month
	if month == time.December+1 {
		year, month = year+1, time.January // carry to next year
	}

	members, err := s.members(year*100 + int(month))
	if err != nil {
		s.error(err)
		return
	}

	data := struct {
		Year    int
		Month   time.Month
		Members []nameAndEmail
	}{
		Year:    year,
		Month:   month,
		Members: members,
	}
	s.render(reminderTmpl, data)
}

type nameAndEmail struct {
	ID    int
	Name  string
	Email string
}

func (s *server) members(queryMonth int) ([]nameAndEmail, error) {
	// MySQL doesn't have a proper boolean data type, and it's
	// json_object seems to have some arbitrary heuristics for
	// deciding when to encode a boolean expression as 0/1 vs
	// true/false.
	type mysqlBool = int

	// Same query as in roster.go, except for the final clause:
	const q = `
with target as (select ?),
sfbay as (
  select *
  from activists a
  where a.chapter_id = 47
    and a.activist_level in ('Organizer', 'Chapter Member')
    and not a.hidden
),
raw_mpi as (
  select a.id, e.month, (max(e.direct_action) and (max(e.community) or e.month >= 202001)) as mpi,
    period_diff(e.month, (select * from target)) as month_delta
  from sfbay a
  left join event_attendance ea on (a.id = ea.activist_id)
  left join (
          select id,
                 extract(year_month from date) as month,
                 event_type in ('Circle', 'Community', 'Training') as community,
                 event_type in ('Action', 'Campaign Action', 'Frontline Surveillance', 'Outreach', 'Sanctuary') as direct_action
          from events
        ) e on (e.id = ea.event_id)
  group by a.id, e.month
),
roster as (
  select a.id, a.name, a.email, a.activist_level, a.cm_approval_email, a.voting_agreement,

  sum(x.mpi and x.month_delta >= -3  and x.month_delta < 0) as mpi_past3,
  sum(x.mpi and x.month_delta >= -12 and x.month_delta < 0) as mpi_past12,

  period_diff(extract(year_month from a.cm_approval_email), (select * from target)) < -6 as cm_approved6

  from sfbay a left join raw_mpi x using (id)
  group by id
)
select json_arrayagg(json_object(
  'ID',            r.id,
  'Name',          r.name,
  'Email',         r.email
))
from roster r
where r.activist_level = 'Chapter Member'
  and r.mpi_past12 >= 8 and r.cm_approved6 and not r.voting_agreement
`

	var members []nameAndEmail
	if err := s.queryJSON(&members, q, queryMonth); err != nil {
		return nil, err
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Name < members[j].Name
	})
	return members, nil
}

var reminderTmpl = template.Must(template.New("reminder").Parse(`
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
<h1>DxE SF Bay Area Chapter Members</h1>

<p>The Chapter Members below would be eligible to vote in {{.Month}} {{.Year}},
if they first sign the voting agreement:</p>

<ul>
{{range .Members}}
<li>{{.Name}} <code>&lt;{{.Email}}&gt;</code></li>
{{end}}
</ul>

</div>
</body>
`))
