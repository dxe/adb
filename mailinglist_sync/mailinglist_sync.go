package mailinglist_sync

import (
	"context"
	"fmt"
	"io/ioutil"
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

func getInsertAndRemoveEmails(wgMembers []model.WorkingGroupMember, listEmails []string) (insertEmails, removeEmails []string) {
	wgMembersEmailMap := map[string]bool{}
	for _, a := range wgMembers {
		// Don't track empty emails.
		e := normalizeEmail(a.ActivistEmail)
		if e != "" {
			wgMembersEmailMap[e] = true
		}
	}
	listEmailMap := map[string]bool{}
	for _, e := range listEmails {
		listEmailMap[e] = true
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

func syncMailingLists(db *sqlx.DB, adminService *admin.Service) error {
	wgs, err := model.GetWorkingGroups(db, model.WorkingGroupQueryOptions{})
	if err != nil {
		return err
	}
	var errs []error
	for _, wg := range wgs {
		listEmails, err := listMembers(adminService, wg.GroupEmail)
		if err != nil {
			errs = append(errs, err)
			// Don't continue processing if we can't get
			// the members list.
			continue
		}

		insertEmails, removeEmails := getInsertAndRemoveEmails(wg.Members, listEmails)

		for _, e := range insertEmails {
			err = insertMember(adminService, wg.GroupEmail, e)
			if err != nil {
				errs = append(errs, err)
				// Fallthrough and try to insert other
				// members if there's a failure.
			}
		}
		for _, e := range removeEmails {
			err = removeMember(adminService, wg.GroupEmail, e)
			if err != nil {
				errs = append(errs, err)
				// Fallthrough and try to delete other
				// members if there's a failure.
			}
		}
	}

	if len(errs) != 0 {
		return errors.Errorf("Received errors during syncMailingLists: %+v", errs)
	}
	return nil
}

func syncMailingListsWrapper(db *sqlx.DB, adminService *admin.Service) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in syncMailingLists", r)
		}
	}()

	err := syncMailingLists(db, adminService)
	if err != nil {
		panic(err)
	}
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
		fmt.Println("Starting mailing lists sync")
		syncMailingListsWrapper(db, adminService)
		fmt.Println("Finished mailing lists sync")
		time.Sleep(5 * time.Minute)
	}
}
