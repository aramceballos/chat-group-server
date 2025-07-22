package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Repository interface {
	FetchUsers() ([]entities.User, error)
	FetchUserById(id string) (entities.User, error)
	UpdateUser(userId string, user entities.UpdateUserInput) error
}

type repository struct {
	db                *sql.DB
	fetchUsersStmt    *sql.Stmt
	fetchUserByIdStmt *sql.Stmt
	checkEmailStmt    *sql.Stmt
	checkUsernameStmt *sql.Stmt
	updateUserStmt    *sql.Stmt
}

func NewRepository(db *sql.DB) Repository {
	repo := &repository{
		db:                db,
		fetchUsersStmt:    nil,
		fetchUserByIdStmt: nil,
		checkEmailStmt:    nil,
		checkUsernameStmt: nil,
		updateUserStmt:    nil,
	}
	repo.prepareStatements()
	return repo
}

func (r *repository) prepareStatements() error {
	var err error
	r.fetchUsersStmt, err = r.db.Prepare(`
		SELECT 
			id, name, avatar_url, created_at
		FROM users
	`)
	if err != nil {
		return fmt.Errorf("failed to repare fetch users statement: %w", err)
	}
	r.fetchUserByIdStmt, err = r.db.Prepare(`
		SELECT 
			id, name, username, email, avatar_url, created_at 
		FROM users 
		WHERE 
			id = $1
	`)
	if err != nil {
		return fmt.Errorf("failed to repare fetch users statement: %w", err)
	}
	r.checkEmailStmt, err = r.db.Prepare(`
		SELECT id 
		FROM users 
		WHERE 
			email = $1 
			AND 
			id != $2
	`)
	if err != nil {
		return fmt.Errorf("failed to repare fetch users statement: %w", err)
	}
	r.checkUsernameStmt, err = r.db.Prepare(`
		SELECT id 
		FROM users 
		WHERE 
			username = $1 AND id != $2
	`)
	if err != nil {
		return fmt.Errorf("failed to repare fetch users statement: %w", err)
	}
	r.updateUserStmt, err = r.db.Prepare(`
		UPDATE users 
		SET 
			name = $1, username = $2, email = $3
		WHERE id = $4
	`)
	if err != nil {
		return fmt.Errorf("failed to repare fetch users statement: %w", err)
	}

	return nil
}

func (r *repository) FetchUsers() ([]entities.User, error) {
	rows, err := r.fetchUsersStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []entities.User{}

	for rows.Next() {
		user := entities.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.AvatarURL, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, nil
}

func (r *repository) FetchUserById(id string) (entities.User, error) {
	var user entities.User
	var username sql.NullString
	var email sql.NullString
	err := r.db.QueryRow("", id).Scan(&user.ID, &user.Name, &username, &email, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return entities.User{}, err
	}

	if username.Valid {
		user.UserName = username.String
	}

	if email.Valid {
		user.Email = email.String
	}
	return user, nil
}

func (r *repository) UpdateUser(userId string, user entities.UpdateUserInput) error {
	var existingUser entities.User
	err := r.db.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", user.Email, userId).Scan(&existingUser.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingUser.ID != 0 {
		return errors.New("a user with this email already exists")
	}

	err = r.db.QueryRow("SELECT id FROM users WHERE username = $1 AND id != $2", user.Username, userId).Scan(&existingUser.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingUser.ID != 0 {
		return errors.New("a user with this username already exists")
	}

	_, err = r.db.Exec("UPDATE users SET name = $1, username = $2, email = $3 WHERE id = $4", user.Name, user.Username, user.Email, userId)
	if err != nil {
		return err
	}

	return nil
}
