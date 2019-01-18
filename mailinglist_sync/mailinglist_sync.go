package mailinglist_sync

import (
	"context"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/admin/directory/v1"
)

func getAdminService() (*admin.Service, error) {
	key, err := ioutil.ReadFile(config.SyncMailingListsConfigFile)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read google auth key")
	}
	oauthConfig, err := google.JWTConfigFromJSON(key, "https://www.googleapis.com/auth/admin.directory.group")
	if err != nil {
		return nil, errors.Wrap(err, "Could not read JWT config from google auth key")
	}
	oauthConfig.Subject = config.SyncMailingListsOauthSubject

	client := oauthConfig.Client(context.Background())
	adminService, err := admin.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "Could not construct admin service")
	}

	return adminService, nil
}

func listMembers(adminService *admin.Service, groupEmail string) ([]string, error) {
	var memberEmails []string
	call := adminService.Members.List(groupEmail)
	err := call.Pages(context.Background(), func(members *admin.Members) error {
		for _, m := range members.Members {
			memberEmails = append(memberEmails, m.Email)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not page members for group %s", groupEmail)
	}
	return memberEmails, nil
}

func insertMember(adminService *admin.Service, groupEmail, memberEmail string) error {
	_, err := adminService.Members.Insert(groupEmail, &admin.Member{Email: memberEmail}).Do()
	return errors.Wrapf(err, "Could not insert member %s into group %s ", memberEmail, groupEmail)
}

func removeMember(adminService *admin.Service, groupEmail, memberEmail string) error {
	err := adminService.Members.Delete(groupEmail, memberEmail).Do()
	return errors.Wrapf(err, "Could not delete member %s from group %s", memberEmail, groupEmail)
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// getInsertAndRemoveEmails returns the list of emails that need to be
// inserted and/or removed to transform current into target.
func getInsertAndRemoveEmails(target, current []string) (insertEmails, removeEmails []string) {
	wgMembersEmailMap := map[string]bool{}
	for _, e := range target {
		normalized := normalizeEmail(e)
		wgMembersEmailMap[normalized] = true
	}

	listEmailMap := map[string]bool{}
	for _, e := range current {
		normalized := normalizeEmail(e)
		listEmailMap[normalized] = true
	}

	insertEmails = []string{}
	for wgMemberEmail := range wgMembersEmailMap {
		e := normalizeEmail(wgMemberEmail)
		if _, ok := listEmailMap[e]; !ok {
			insertEmails = append(insertEmails, e)
		}
	}
	removeEmails = []string{}
	for listEmail := range listEmailMap {
		e := normalizeEmail(listEmail)
		if _, ok := wgMembersEmailMap[e]; !ok {
			removeEmails = append(removeEmails, e)
		}
	}

	return insertEmails, removeEmails
}

func syncMailingList(adminService *admin.Service, groupEmail string, memberEmails []string) {
	listEmails, err := listMembers(adminService, groupEmail)
	if err != nil {
		// Don't continue processing if we can't get
		// the members list.
		log.Printf("Failed to list members of %v: %v", groupEmail, err)
		return
	}

	insertEmails, removeEmails := getInsertAndRemoveEmails(memberEmails, listEmails)
	if len(insertEmails) != 0 || len(removeEmails) != 0 {
		log.Printf("Syncing %v: +%q, -%q", groupEmail, insertEmails, removeEmails)
	}

	for _, e := range removeEmails {
		err := removeMember(adminService, groupEmail, e)
		if err != nil {
			log.Printf("Failed to remove %v from group %v", e, groupEmail)
			// Continue processing.
		}
	}
	for _, e := range insertEmails {
		err := insertMember(adminService, groupEmail, e)
		if err != nil {
			log.Printf("Failed to add %v to group %v", e, groupEmail)
			// Continue processing.
		}
	}
}

func syncWorkingGroupMailingLists(db *sqlx.DB, adminService *admin.Service) {
	wgs, err := model.GetWorkingGroups(db, model.WorkingGroupQueryOptions{})
	if err != nil {
		log.Println("Failed to query working groups")
		return
	}

	for _, wg := range wgs {
		var memberEmails []string
		for _, m := range wg.Members {
			email := normalizeEmail(m.ActivistEmail)
			if email == "" {
				log.Printf("Activist has no email, will not be synced to mailing list: %s\n", m.ActivistName)
				continue
			}
			memberEmails = append(memberEmails, email)
		}
		syncMailingList(adminService, wg.GroupEmail, memberEmails)
	}
}

func syncCircleHostMailingList(db *sqlx.DB, adminService *admin.Service) {
	// Sync circlehosts@directactioneverywhere.com to contain all
	// circle hosts.

	circles, err := model.GetCircleGroups(db, model.CircleGroupQueryOptions{})
	if err != nil {
		log.Println("Failed to query circles")
		return
	}

	// The circle host's email is used as the circle's group email.
	var emails []string
	for _, c := range circles {
		email := normalizeEmail(c.GroupEmail)
		if email == "" {
			log.Printf("Circle has no email, will not be synced to mailing list: %s", c.Name)
			continue
		}
		emails = append(emails, email)
	}

	syncMailingList(adminService, "circlehosts@directactioneverywhere.com", emails)
}

func syncChapterMemberMailingList(db *sqlx.DB, adminService *admin.Service) {
	// Sync chaptermembers@directactioneverywhere.com to contain all
	// activists that are considered a Chapter Member; i.e. Activists that
	// that have activist_level of "Chapter Member".

	members, err := model.GetChapterMembers(db)
	if err != nil {
		log.Println("Failed to query chapters")
		return
	}

	var emails []string
	for _, m := range members {
		email := normalizeEmail(m.Email)
		if email == "" {
			log.Printf("Activist has no email, will not be synced to mailing list: %s", m.Name)
			continue
		}

		emails = append(emails, email)
	}

	syncMailingList(adminService, "chaptermembers@directactioneverywhere.com", emails)
}

func syncMailingListsWrapper(db *sqlx.DB, adminService *admin.Service) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in syncMailingLists", r)
		}
	}()

	syncWorkingGroupMailingLists(db, adminService)
	syncCircleHostMailingList(db, adminService)
	syncChapterMemberMailingList(db, adminService)
}

// Syncs the mailing list every 5 minutes. Should be run in a
// goroutine.
func StartMailingListsSync(db *sqlx.DB) {
	adminService, err := getAdminService()
	if err != nil {
		// Just panic if we can't get an admin service so that
		// we don't accidentally mess this up without
		// realizing it.
		panic(err)
	}

	for {
		log.Println("Starting mailing lists sync")
		syncMailingListsWrapper(db, adminService)
		log.Println("Finished mailing lists sync")
		time.Sleep(5 * time.Minute)
	}
}
