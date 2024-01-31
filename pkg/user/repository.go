package user

import (
	"database/sql"

	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Repository interface {
	FetchUsers() ([]entities.User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db,
	}
}

func (r *repository) FetchUsers() ([]entities.User, error) {
	rows, err := r.db.Query("SELECT id, name, avatar_url, created_at FROM users")
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
