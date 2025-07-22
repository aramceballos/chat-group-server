package user

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
}

func TestFetchUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "avatar_url", "created_at"}).
			AddRow(1, "John Doe", "http://example.com/avatar1.jpg", "2023-01-01 00:00:00").
			AddRow(2, "Jane Smith", "http://example.com/avatar2.jpg", "2023-01-02 00:00:00")

		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users").WillReturnRows(rows)

		users, err := repo.FetchUsers()
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "John Doe", users[0].Name)
		assert.Equal(t, "Jane Smith", users[1].Name)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users").WillReturnError(errors.New("database error"))

		users, err := repo.FetchUsers()
		assert.Error(t, err)
		assert.Nil(t, users)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "avatar_url", "created_at"}).
			AddRow("invalid", "John Doe", "http://example.com/avatar1.jpg", "2023-01-01 00:00:00")

		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users").WillReturnRows(rows)

		users, err := repo.FetchUsers()
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}

func TestFetchUserById(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("success with username and email", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "name", "username", "email", "avatar_url", "created_at"}).
			AddRow(1, "John Doe", "johndoe", "john@example.com", "http://example.com/avatar.jpg", "2023-01-01 00:00:00")

		mock.ExpectQuery("SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = \\$1").
			WithArgs("1").WillReturnRows(row)

		user, err := repo.FetchUserById("1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "johndoe", user.UserName)
		assert.Equal(t, "john@example.com", user.Email)
	})

	t.Run("success with null username and email", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "name", "username", "email", "avatar_url", "created_at"}).
			AddRow(1, "John Doe", nil, nil, "http://example.com/avatar.jpg", "2023-01-01 00:00:00")

		mock.ExpectQuery("SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = \\$1").
			WithArgs("1").WillReturnRows(row)

		user, err := repo.FetchUserById("1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "", user.UserName)
		assert.Equal(t, "", user.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = \\$1").
			WithArgs("999").WillReturnError(sql.ErrNoRows)

		user, err := repo.FetchUserById("999")
		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, username, email, avatar_url, created_at FROM users WHERE id = \\$1").
			WithArgs("1").WillReturnError(errors.New("database error"))

		user, err := repo.FetchUserById("1")
		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
	})
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("success", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "john.updated@example.com",
		}

		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1 AND id != \\$2").
			WithArgs(updateInput.Username, "1").WillReturnError(sql.ErrNoRows)

		mock.ExpectExec("UPDATE users SET name = \\$1, username = \\$2, email = \\$3 WHERE id = \\$4").
			WithArgs(updateInput.Name, updateInput.Username, updateInput.Email, "1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUser("1", updateInput)
		assert.NoError(t, err)
	})

	t.Run("email already exists", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "existing@example.com",
		}

		row := sqlmock.NewRows([]string{"id"}).AddRow(2)
		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnRows(row)

		err := repo.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Equal(t, "a user with this email already exists", err.Error())
	})

	t.Run("username already exists", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "existinguser",
			Email:    "john.updated@example.com",
		}

		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnError(sql.ErrNoRows)

		row := sqlmock.NewRows([]string{"id"}).AddRow(2)
		mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1 AND id != \\$2").
			WithArgs(updateInput.Username, "1").WillReturnRows(row)

		err := repo.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Equal(t, "a user with this username already exists", err.Error())
	})

	t.Run("database error on email check", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "john.updated@example.com",
		}

		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnError(errors.New("database error"))

		err := repo.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
	})

	t.Run("database error on username check", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "john.updated@example.com",
		}

		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1 AND id != \\$2").
			WithArgs(updateInput.Username, "1").WillReturnError(errors.New("database error"))

		err := repo.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
	})

	t.Run("database error on update", func(t *testing.T) {
		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "john.updated@example.com",
		}

		mock.ExpectQuery("SELECT id FROM users WHERE email = \\$1 AND id != \\$2").
			WithArgs(updateInput.Email, "1").WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT id FROM users WHERE username = \\$1 AND id != \\$2").
			WithArgs(updateInput.Username, "1").WillReturnError(sql.ErrNoRows)

		mock.ExpectExec("UPDATE users SET name = \\$1, username = \\$2, email = \\$3 WHERE id = \\$4").
			WithArgs(updateInput.Name, updateInput.Username, updateInput.Email, "1").
			WillReturnError(errors.New("update error"))

		err := repo.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
	})
}
