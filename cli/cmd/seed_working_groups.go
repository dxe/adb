package cmd

import (
	"fmt"

	"github.com/dxe/adb/cli/internal/db"
	"github.com/dxe/adb/pkg/shared"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// Seeds a small set of realistic working groups along with the activists on their rosters. Unlike the activist seed,
// this data is fixed rather than randomized so that the working groups page always shows the same groups, point
// people, members, and non-members on the mailing list.
//
// The command is idempotent: working groups are matched by name and activists by (name, chapter), so re-running it
// refreshes the seeded data in place instead of creating duplicates.

var seedWorkingGroupsChapterID int

// Activist levels assigned to seeded roster entries. Point people must be organizers to appear in the point person
// autocomplete on the working groups page (see GetAutocompleteOrganizerNames).
const (
	seedPointPersonLevel = "Organizer"
	seedMemberLevel      = "Chapter Member"
	seedNonMemberLevel   = "Supporter"
)

// seedWorkingGroup describes one working group and its roster.
type seedWorkingGroup struct {
	Name            string
	GroupEmail      string
	Description     string
	MeetingTime     string
	MeetingLocation string
	PointPerson     string
	// Members are on the roster and the mailing list.
	Members []string
	// NonMembers are on the mailing list only.
	NonMembers []string
}

// seedWorkingGroupMember is one flattened roster entry ready to be persisted.
type seedWorkingGroupMember struct {
	Name                   string
	ActivistLevel          string
	PointPerson            bool
	NonMemberOnMailingList bool
}

var seedWorkingGroupData = []seedWorkingGroup{
	{
		Name:            "Tech",
		GroupEmail:      "tech@directactioneverywhere.com",
		Description:     "Designs, engineers, & supports DxE's websites & applications.",
		MeetingTime:     "Wednesdays, 7:30pm (varies)",
		MeetingLocation: "ARC/Online",
		PointPerson:     "Alexander Taylor",
		Members:         []string{"Jake Hobbs", "Matthew Zirbel"},
	},
	{
		Name:            "Communications",
		GroupEmail:      "communications@directactioneverywhere.com",
		Description:     "This team doesn't meet but coordinates as needed.",
		MeetingTime:     "N/A",
		MeetingLocation: "N/A",
		PointPerson:     "Cassie King",
		NonMembers:      []string{"Zoe Rosenberg"},
	},
}

func init() {
	seedCmd.AddCommand(seedWorkingGroupsCmd)
	seedWorkingGroupsCmd.Flags().IntVar(&seedWorkingGroupsChapterID, "chapter-id", shared.SFBayChapterIdDevTest, "Chapter ID to assign to seeded activists")
}

var seedWorkingGroupsCmd = &cobra.Command{
	Use:   "working-groups",
	Short: "Seed working groups along with the activists on their rosters",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = conn.Close() }()

		memberCount := 0
		for _, wg := range seedWorkingGroupData {
			members, err := seedOneWorkingGroup(conn, wg, seedWorkingGroupsChapterID)
			if err != nil {
				return err
			}
			memberCount += members
		}

		fmt.Printf(
			"Seeded %d working groups with %d members in chapter %d\n",
			len(seedWorkingGroupData), memberCount, seedWorkingGroupsChapterID,
		)
		return nil
	},
}

// roster returns the working group's point person, members, and non-members as a single list of roster entries.
func (wg seedWorkingGroup) roster() []seedWorkingGroupMember {
	members := make([]seedWorkingGroupMember, 0, 1+len(wg.Members)+len(wg.NonMembers))
	if wg.PointPerson != "" {
		members = append(members, seedWorkingGroupMember{
			Name:          wg.PointPerson,
			ActivistLevel: seedPointPersonLevel,
			PointPerson:   true,
		})
	}
	for _, name := range wg.Members {
		members = append(members, seedWorkingGroupMember{
			Name:          name,
			ActivistLevel: seedMemberLevel,
		})
	}
	for _, name := range wg.NonMembers {
		members = append(members, seedWorkingGroupMember{
			Name:                   name,
			ActivistLevel:          seedNonMemberLevel,
			NonMemberOnMailingList: true,
		})
	}
	return members
}

// seedOneWorkingGroup upserts a working group and replaces its roster, creating any missing activists. It returns the
// number of roster entries written.
func seedOneWorkingGroup(conn *sqlx.DB, wg seedWorkingGroup, chapterID int) (int, error) {
	tx, err := conn.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction for working group %q: %w", wg.Name, err)
	}

	members, err := seedWorkingGroupTx(tx, wg, chapterID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to commit working group %q: %w", wg.Name, err)
	}
	return members, nil
}

func seedWorkingGroupTx(tx *sqlx.Tx, wg seedWorkingGroup, chapterID int) (int, error) {
	workingGroupID, err := upsertSeedWorkingGroup(tx, wg)
	if err != nil {
		return 0, err
	}

	// Replace the roster so that removing someone from the seed data removes them from the working group too.
	if _, err := tx.Exec(
		`DELETE FROM working_group_members WHERE working_group_id = ?`, workingGroupID,
	); err != nil {
		return 0, fmt.Errorf("failed to clear members of working group %q: %w", wg.Name, err)
	}

	members := wg.roster()
	for _, member := range members {
		activistID, err := upsertSeedActivist(tx, member, chapterID)
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(
			`INSERT INTO working_group_members (working_group_id, activist_id, point_person, non_member_on_mailing_list)
			 VALUES (?, ?, ?, ?)`,
			workingGroupID, activistID, member.PointPerson, member.NonMemberOnMailingList,
		); err != nil {
			return 0, fmt.Errorf("failed to add %q to working group %q: %w", member.Name, wg.Name, err)
		}
	}

	return len(members), nil
}

// upsertSeedWorkingGroup inserts the working group, or updates it in place when one already exists with the same
// name, and returns its ID.
func upsertSeedWorkingGroup(tx *sqlx.Tx, wg seedWorkingGroup) (int, error) {
	// Seeded groups are visible so that they show up wherever the app filters on visibility.
	if _, err := tx.Exec(
		`INSERT INTO working_groups (name, group_email, visible, description, meeting_time, meeting_location, coords)
		 VALUES (?, ?, 1, ?, ?, ?, '')
		 ON DUPLICATE KEY UPDATE
		   group_email = VALUES(group_email),
		   visible = VALUES(visible),
		   description = VALUES(description),
		   meeting_time = VALUES(meeting_time),
		   meeting_location = VALUES(meeting_location)`,
		wg.Name, wg.GroupEmail, wg.Description, wg.MeetingTime, wg.MeetingLocation,
	); err != nil {
		return 0, fmt.Errorf("failed to insert working group %q: %w", wg.Name, err)
	}

	var workingGroupID int
	if err := tx.Get(&workingGroupID, `SELECT id FROM working_groups WHERE name = ?`, wg.Name); err != nil {
		return 0, fmt.Errorf("failed to get ID of working group %q: %w", wg.Name, err)
	}
	return workingGroupID, nil
}

// upsertSeedActivist creates the roster entry's activist if they don't exist yet and returns their ID. Existing
// activists keep their contact info, but their level is set to the one the roster requires.
func upsertSeedActivist(tx *sqlx.Tx, member seedWorkingGroupMember, chapterID int) (int, error) {
	if _, err := tx.Exec(
		`INSERT INTO activists (name, email, chapter_id, activist_level) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE activist_level = VALUES(activist_level), hidden = 0`,
		member.Name, seedEmail(member.Name, chapterID), chapterID, member.ActivistLevel,
	); err != nil {
		return 0, fmt.Errorf("failed to insert activist %q: %w", member.Name, err)
	}

	var activistID int
	if err := tx.Get(
		&activistID, `SELECT id FROM activists WHERE name = ? AND chapter_id = ?`, member.Name, chapterID,
	); err != nil {
		return 0, fmt.Errorf("failed to get ID of activist %q: %w", member.Name, err)
	}
	return activistID, nil
}
