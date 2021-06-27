package members

import (
	"encoding/csv"
	"errors"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *server) roster() {
	email, err := s.googleEmail()
	if err != nil {
		s.redirect(absURL("/login"))
		return
	}

	if !isAdmin(email) {
		s.error(errors.New("unauthorized"))
		return
	}

	if q, ok := s.r.URL.Query()["month"]; ok && len(q) == 1 {
		month, err := strconv.Atoi(q[0])
		if err != nil {
			s.error(err)
			return
		}
		if month < 190000 {
			s.error(errors.New("invalid month query; must be YYYYMM syntax"))
			return
		}
		s.rosterDownload(month)
		return
	}

	s.render(rosterTmpl, nil)
}

func (s *server) rosterDownload(queryMonth int) {
	// MySQL doesn't have a proper boolean data type, and it's
	// json_object seems to have some arbitrary heuristics for
	// deciding when to encode a boolean expression as 0/1 vs
	// true/false.
	type mysqlBool = int

	var data []struct {
		ID            int
		Name          string
		Email         string
		ActivistLevel string

		Eligible mysqlBool

		// Past3 and Past12 are how many of the past 3
		// and 12 months, respectively, the activist
		// fulfilled MPI requirements.
		MPIPast3, MPIPast12 int

		// CMApproved6 is the day the member was
		// approved as a chapter member.
		CMApproval string

		// VotingAgreement is whether the member has
		// signing the voting agreement.
		VotingAgreement mysqlBool
	}

	// This query would be more natural if attendance could be
	// computed using a subquery like working groups, but because
	// of the two-level aggregation, we'd actually need a
	// sub-subquery; and subqueries can only access variables from
	// the immediately outer context.
	// TODO: consider not hard-coding the SF Bay chapter ID in the WHERE clause
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
  'Email',         r.email,
  'ActivistLevel', r.activist_level,

  'Eligible', case r.activist_level
    when 'Organizer'      then r.mpi_past3 >= 2
    when 'Chapter Member' then r.mpi_past12 >= 8 and r.cm_approved6 and r.voting_agreement
    else                       false
  end,

  'VotingAgreement', r.voting_agreement,
  'MPIPast3',        r.mpi_past3,
  'MPIPast12',       r.mpi_past12,
  'CMApproval',      r.cm_approval_email
))
from roster r
`

	if err := s.queryJSON(&data, q, queryMonth); err != nil {
		s.error(err)
		return
	}

	sort.Slice(data.Members, func(i, j int) bool {
		return data.Members[i].Name < data.Members[j].Name
	})

	extra := ""
	if year, month, day := time.Now().Date(); queryMonth > year*100+int(month) {
		extra = fmt.Sprintf(" (Tentative as of %04d-%02d-%02d)", year, month, day)
	}
	filename := fmt.Sprintf("%s %04d Roster%s.csv", time.Month(queryMonth%100), queryMonth/100, 0, extra)

	h := s.w.Header()
	h.Set("Content-Type", "text/csv")
	// https://stackoverflow.com/a/68154942/2342228
	h.Set("Content-Disposition", `attachment; filename="`+quoteEscaper.Replace(filename)+`"`)

	yesNo := func(b bool) string {
		if b {
			return "Yes"
		}
		return "No"
	}

	w := csv.NewWriter(s.w)
	w.Write([]string{"ID", "Name", "Email", "Activist Level", "Eligible", "MPI (3 months)", "MPI (12 months)", "CM Approval", "Voting Agreement"})
	for _, member := range data.Members {
		w.Write([]string{fmt.Sprint(member.ID), member.Name, member.Email, member.ActivistLevel, yesNo(member.Eligible != 0), fmt.Sprint(member.MPIPast3), fmt.Sprint(member.MPIPast12), member.CMApproval, yesNo(member.VotingAgreement != 0)})
	}
	w.Flush()

	if err := w.Error(); err != nil {
		// TODO(mdempsky): Anything we can do about this?
		// We've already written the HTTP header at this
		// point. Can't change response code to 5xx.
		log.Printf("error writing csv: %v", err)
	}
}

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

var rosterTmpl = template.Must(template.New("roster").Parse(`
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
<h1>DxE SF Bay Area Eligible Voter Rosters</h1>

<p>Available rosters:</p>

<ul>
<li><a href="/roster?month=202106">June 2021</a></li>
<li><a href="/roster?month=202107">July 2021</a> (tentative!)</li>
</ul>

</div>
</body>
`))
