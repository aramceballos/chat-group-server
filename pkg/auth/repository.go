package auth

import (
	"database/sql"

	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Repository interface {
	GetUserById(id string) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
	GetUserByUsername(username string) (*entities.User, error)
	CreateUser(user entities.User) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db,
	}
}

func (r *repository) GetUserById(id string) (*entities.User, error) {
	var user entities.User
	var username sql.NullString
	var email sql.NullString
	err := r.db.QueryRow("SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &username, &email, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	if username.Valid {
		user.UserName = username.String
	}

	if email.Valid {
		user.Email = email.String
	}

	return &user, nil
}

func (r *repository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.QueryRow("SELECT id, name, username, email, password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.UserName, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(username string) (*entities.User, error) {
	var user entities.User
	err := r.db.QueryRow("SELECT id, name, username, email, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Name, &user.UserName, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) CreateUser(user entities.User) error {
	_, err := r.db.Exec("INSERT INTO users (name, email, username, password, avatar_url) VALUES ($1, $2, $3, $4, $5)", user.Name, user.Email, user.UserName, user.Password, user.AvatarURL)
	if err != nil {
		return err
	}

	return nil
}
