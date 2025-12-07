package persistence

import (
	"database/sql"
	"log"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DBUserRepository struct {
	db *sqlx.DB
}

var _ model.UserRepository = (*DBUserRepository)(nil)

func NewUserRepository(db *sqlx.DB) *DBUserRepository {
	return &DBUserRepository{db: db}
}

func (r *DBUserRepository) GetUser(id int, email string) (model.ADBUser, error) {
	query := `
SELECT
  id,
  email,
  name,
  disabled,
  chapter_id,
  @chapter_name := IFNULL((
    SELECT name
    FROM fb_pages
    WHERE fb_pages.chapter_id = adb_users.chapter_id
  ),"") AS chapter_name
FROM adb_users
`
	var queryArgs []interface{}
	if id != 0 {
		query += " WHERE id = ? "
		queryArgs = append(queryArgs, id)
	} else if email != "" {
		query += " WHERE email = ? "
		queryArgs = append(queryArgs, email)
	} else {
		return model.ADBUser{}, errors.New("Must supply id or email")
	}

	adbUser := &model.ADBUser{}
	if err := r.db.Get(adbUser, query, queryArgs...); err != nil {
		return model.ADBUser{}, errors.Wrapf(err, "cannot get adb user %d", id)
	}

	usersRoles, err := getUsersRoles(r.db)

	// We don't want non-SF Bay users to have access to any of the other roles, so just replace it.
	if adbUser.ChapterName != model.SFBayChapterName {
		usersRoles = []model.UserRole{{
			UserID: adbUser.ID,
			Role:   "non-sfbay",
		}}
	}

	if err != nil || len(usersRoles) == 0 {
		return *adbUser, nil
	}

	for _, r := range usersRoles {
		if r.UserID == adbUser.ID {
			adbUser.Roles = append(adbUser.Roles, r.Role)
		}
	}

	log.Println("[User access]", adbUser.Name, "-", adbUser.Email)

	return *adbUser, nil
}

// TODO(https://github.com/dxe/adb/issues/292): switch lookup by ID or exact match on name to GetUser
func (r *DBUserRepository) GetUsers(options model.GetUserOptions) ([]model.ADBUser, error) {
	users, err := getUsers(r.db, options)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return users, nil
	}

	if !options.PopulateRoles {
		return users, nil
	}

	usersRoles, err := getUsersRoles(r.db)

	if err != nil {
		return nil, err
	}

	userIDToIndex := map[int]int{}
	for i, user := range users {
		userIDToIndex[user.ID] = i
	}

	for _, r := range usersRoles {
		if a, ok := userIDToIndex[r.UserID]; ok {
			users[a].Roles = append(users[a].Roles, r.Role)
		}
	}

	return users, nil
}

func (r *DBUserRepository) CreateUser(user model.ADBUser) (model.ADBUser, error) {
	if user.ID != 0 {
		return model.ADBUser{}, errors.New("User ID must be 0 when creating a user")
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Name = strings.TrimSpace(user.Name)

	tx, err := r.db.Beginx()
	if err != nil {
		return model.ADBUser{}, errors.Wrap(err, "failed to start create user transaction")
	}
	defer tx.Rollback()

	var existing struct {
		ID       int  `db:"id"`
		Disabled bool `db:"disabled"`
	}

	err = tx.Get(&existing, `SELECT id, disabled FROM adb_users WHERE email = ?`, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return model.ADBUser{}, errors.Wrapf(err, "failed to check existing user %s", user.Email)
	}
	if err == nil {
		if existing.Disabled {
			return model.ADBUser{}, errors.Errorf("user with email %s already exists and is suspended", user.Email)
		}
		return model.ADBUser{}, errors.Errorf("user with email %s already exists", user.Email)
	}

	result, err := tx.NamedExec(`
INSERT INTO adb_users (
  email,
  name,
  disabled,
  chapter_id
) VALUES (
  :email,
  :name,
  :disabled,
  :chapter_id
	)`, user)

	if err != nil {
		return model.ADBUser{}, errors.Wrapf(err, "Could not create user: %s", user.Email)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.ADBUser{}, errors.Wrapf(err, "Could not get LastInsertId for %s", user.Email)
	}

	userID := int(id)

	if err := syncUserRolesTx(tx, userID, user.Roles); err != nil {
		return model.ADBUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.ADBUser{}, errors.Wrap(err, "failed to commit create user transaction")
	}

	users, err := r.GetUsers(model.GetUserOptions{ID: userID, PopulateRoles: true})
	if err != nil {
		return model.ADBUser{}, err
	}
	if len(users) == 0 {
		return model.ADBUser{}, errors.Errorf("no user found with ID %d after create", userID)
	}
	return users[0], nil
}

func (r *DBUserRepository) UpdateUser(user model.ADBUser) (model.ADBUser, error) {
	if user.ID == 0 {
		return model.ADBUser{}, errors.New("User ID cannot be 0 when updating a user")
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Name = strings.TrimSpace(user.Name)

	tx, err := r.db.Beginx()
	if err != nil {
		return model.ADBUser{}, errors.Wrap(err, "failed to start update user transaction")
	}
	defer tx.Rollback()

	var existing struct {
		ID       int  `db:"id"`
		Disabled bool `db:"disabled"`
	}

	err = tx.Get(&existing, `SELECT id, disabled FROM adb_users WHERE email = ?`, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return model.ADBUser{}, errors.Wrapf(err, "failed to check existing user %s", user.Email)
	}
	if err == nil && existing.ID != user.ID {
		if existing.Disabled {
			return model.ADBUser{}, errors.Errorf("user with email %s already exists and is suspended", user.Email)
		}
		return model.ADBUser{}, errors.Errorf("user with email %s already exists", user.Email)
	}

	_, err = tx.NamedExec(`UPDATE adb_users
SET
  email = :email,
  name  = :name,
  disabled = :disabled,
  chapter_id = :chapter_id
	WHERE
	id = :id`, user)

	if err != nil {
		return model.ADBUser{}, errors.Wrap(err, "failed to update user data")
	}

	if err := syncUserRolesTx(tx, user.ID, user.Roles); err != nil {
		return model.ADBUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.ADBUser{}, errors.Wrap(err, "failed to commit update user transaction")
	}

	users, err := r.GetUsers(model.GetUserOptions{ID: user.ID, PopulateRoles: true})
	if err != nil {
		return model.ADBUser{}, err
	}
	if len(users) == 0 {
		return model.ADBUser{}, errors.Errorf("no user found with ID %d after update", user.ID)
	}
	return users[0], nil
}

func getUsers(db *sqlx.DB, options model.GetUserOptions) ([]model.ADBUser, error) {
	query := `
SELECT
  id,
  email,
  name,
  disabled,
  chapter_id
FROM adb_users
`

	var queryArgs []interface{}

	if options.ID != 0 && options.Name != "" {
		return nil, errors.New("You may provide ID or Name but not both.")
	}

	if options.ID != 0 {
		query += " WHERE id = ? "
		queryArgs = append(queryArgs, options.ID)
	}

	if options.Name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, options.Name)
	}

	query += " ORDER BY email "

	var users []model.ADBUser
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get users")
	}

	return users, nil
}

func getUsersRoles(db *sqlx.DB) ([]model.UserRole, error) {
	query := `
SELECT
  ur.user_id,
  ur.role
FROM users_roles ur
`

	var userRoles []model.UserRole
	err := db.Select(&userRoles, query)

	if err != nil {
		return nil, errors.Wrap(err, "failed to select UserRoles")
	}

	if len(userRoles) == 0 {
		return nil, nil
	}

	return userRoles, nil
}

func syncUserRolesTx(tx *sqlx.Tx, userID int, roles []string) error {
	existingRoles, err := getUserRolesTx(tx, userID)
	if err != nil {
		return err
	}

	existingSet := map[string]struct{}{}
	for _, r := range existingRoles {
		existingSet[r] = struct{}{}
	}

	desiredSet := map[string]struct{}{}
	for _, r := range roles {
		desiredSet[r] = struct{}{}
	}

	for role := range existingSet {
		if _, ok := desiredSet[role]; ok {
			continue
		}
		if _, err := tx.Exec(`DELETE FROM users_roles WHERE user_id = ? AND role = ?`, userID, role); err != nil {
			return errors.Wrapf(err, "failed to remove role %s for user %d", role, userID)
		}
	}

	for role := range desiredSet {
		if _, ok := existingSet[role]; ok {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO users_roles (user_id, role) VALUES (?, ?)`, userID, role); err != nil {
			return errors.Wrapf(err, "failed to add role %s for user %d", role, userID)
		}
	}

	return nil
}

func getUserRolesTx(tx *sqlx.Tx, userID int) ([]string, error) {
	var roles []string
	if err := tx.Select(&roles, `SELECT role FROM users_roles WHERE user_id = ?`, userID); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch roles for user %d", userID)
	}
	return roles, nil
}
