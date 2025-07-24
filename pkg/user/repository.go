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
	Close() error
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
		db: db,
	}
	if err := repo.prepareStatements(); err != nil {
		panic(err.Error())
	}
	return repo
}

func (r *repository) prepareStatements() error {
	statements := []struct {
		stmt  **sql.Stmt
		query string
		name  string
	}{
		{
			stmt:  &r.fetchUsersStmt,
			query: "SELECT id, name, avatar_url, created_at FROM users;",
			name:  "fetch users",
		},
		{
			stmt:  &r.fetchUserByIdStmt,
			query: "SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = $1;",
			name:  "fetch user by id",
		},
		{
			stmt:  &r.checkEmailStmt,
			query: "SELECT id FROM users WHERE email = $1 AND id != $2;",
			name:  "check user existance by email",
		},
		{
			stmt:  &r.checkUsernameStmt,
			query: "SELECT id FROM users WHERE username = $1 AND id != $2;",
			name:  "check user existance by email",
		},
		{
			stmt:  &r.updateUserStmt,
			query: "UPDATE users SET name = $1, username = $2, email = $3 WHERE id = $4;",
			name:  "update user information",
		},
	}

	for _, s := range statements {
		stmt, err := r.db.Prepare(s.query)
		if err != nil {
			return fmt.Errorf("faile to prepare %s statement: %w", s.name, err)
		}
		*s.stmt = stmt
	}

	return nil
}

func (r *repository) Close() error {
	statements := []*sql.Stmt{
		r.fetchUsersStmt,
		r.fetchUserByIdStmt,
		r.checkEmailStmt,
		r.checkUsernameStmt,
		r.updateUserStmt,
	}

	for _, stmt := range statements {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				return err
			}
		}
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
	err := r.fetchUserByIdStmt.QueryRow(id).Scan(&user.ID, &user.Name, &username, &email, &user.AvatarURL, &user.CreatedAt)
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
	err := r.checkEmailStmt.QueryRow(user.Email, userId).Scan(&existingUser.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingUser.ID != 0 {
		return errors.New("a user with this email already exists")
	}

	err = r.checkUsernameStmt.QueryRow(user.Username, userId).Scan(&existingUser.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingUser.ID != 0 {
		return errors.New("a user with this username already exists")
	}

	_, err = r.updateUserStmt.Exec(user.Name, user.Username, user.Email, userId)
	if err != nil {
		return err
	}

	return nil
}
