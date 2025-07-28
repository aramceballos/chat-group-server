package chat

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func setupMockRepository(t *testing.T) (*sql.DB, sqlmock.Sqlmock, Repository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	mock.ExpectPrepare("SELECT EXISTS \\(SELECT 1 FROM memberships WHERE channel_id = \\$1 AND user_id = \\$2\\)")
	mock.ExpectPrepare("SELECT id, name, avatar_url, created_at FROM users WHERE id = \\$1")
	mock.ExpectPrepare("INSERT INTO messages \\(channel_id, user_id, body\\) VALUES \\(\\$1, \\$2, \\$3::jsonb\\) RETURNING id, user_id, channel_id, body, created_at")

	repo := NewRepository(db)
	assert.NotNil(t, repo)

	return db, mock, repo
}

func TestNewRepository(t *testing.T) {
	db, mock, repo := setupMockRepository(t)
	defer db.Close()

	assert.NotNil(t, repo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserMembership(t *testing.T) {
	db, mock, repo := setupMockRepository(t)
	defer db.Close()

	t.Run("user is a member", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM memberships WHERE channel_id = \\$1 AND user_id = \\$2\\)").
			WithArgs(1, 1).
			WillReturnRows(row)

		exists, err := repo.CheckUserMembership(1, 1)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("user is not a member", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM memberships WHERE channel_id = \\$1 AND user_id = \\$2\\)").
			WithArgs(1, 10).
			WillReturnRows(row)

		exists, err := repo.CheckUserMembership(1, 10)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM memberships WHERE channel_id = \\$1 AND user_id = \\$2\\)").
			WillReturnError(errors.New("database error"))

		exists, err := repo.CheckUserMembership(1, 10)
		assert.Error(t, err)
		assert.False(t, exists)
	})
}

func TestFetchUserById(t *testing.T) {
	db, mock, repo := setupMockRepository(t)
	defer db.Close()

	t.Run("user exists", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "name", "avatar_url", "created_at"}).
			AddRow(1, "John Doe", "http://example.com/avatar.jpg", "2023-01-01 00:00:00")

		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(row)

		user, err := repo.FetchUserById(1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "http://example.com/avatar.jpg", user.AvatarURL)
		assert.Equal(t, "2023-01-01 00:00:00", user.CreatedAt)
	})

	t.Run("user does not exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users WHERE id = \\$1").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.FetchUserById(1)
		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, avatar_url, created_at FROM users WHERE id = \\$1").
			WillReturnError(errors.New("database error"))

		user, err := repo.FetchUserById(1)
		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
	})
}

func TestInsertMessage(t *testing.T) {
	db, mock, repo := setupMockRepository(t)
	defer db.Close()

	t.Run("success insert", func(t *testing.T) {
		messageBodyMock := []byte("{\"type\": \"text\",\"content\": \"Sii\"}")
		row := sqlmock.NewRows([]string{"id", "user_id", "channel_id", "body", "created_at"}).
			AddRow(1, 1, 3, messageBodyMock, "2023-01-01 00:00:00")

		mock.ExpectQuery("INSERT INTO messages \\(channel_id, user_id, body\\) VALUES \\(\\$1, \\$2, \\$3::jsonb\\) RETURNING id, user_id, channel_id, body, created_at").
			WithArgs(1, 1, messageBodyMock).
			WillReturnRows(row)

		message, err := repo.InsertMessage(1, 1, messageBodyMock)
		assert.NoError(t, err)
		assert.Equal(t, 1, message.ID)
		assert.Equal(t, int64(1), message.UserID)
		assert.Equal(t, 3, message.ChannelID)
		assert.Equal(t, messageBodyMock, []byte(message.Body))
		assert.Equal(t, "2023-01-01 00:00:00", message.CreatedAt)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO messages \\(channel_id, user_id, body\\) VALUES \\(\\$1, \\$2, \\$3::jsonb\\) RETURNING id, user_id, channel_id, body, created_at").
			WillReturnError(errors.New("database error"))

		message, err := repo.InsertMessage(1, 1, []byte{})
		assert.Error(t, err)
		assert.Equal(t, entities.Message{}, message)
	})
}
