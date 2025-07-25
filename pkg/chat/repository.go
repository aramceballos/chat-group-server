package chat

import (
	"database/sql"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/gofiber/fiber/v2/log"
)

type Repository interface {
	CheckUserMembership(channelId int, userId int) (bool, error)
	FetchUserById(userId int) (entities.User, error)
	InsertMessage(channelId int, userId int, msgBody []byte) (entities.Message, error)
	Close() error
}

type repository struct {
	db                  *sql.DB
	checkMembershipStmt *sql.Stmt
	fetchUserByIdStmt   *sql.Stmt
	insertedMessageStmt *sql.Stmt
}

func NewRepository(db *sql.DB) Repository {
	repo := &repository{
		db,
		nil,
		nil,
		nil,
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
			stmt:  &r.checkMembershipStmt,
			query: "SELECT EXISTS (SELECT 1 FROM memberships WHERE channel_id = $1 AND user_id = $2);",
			name:  "check user membership",
		},
		{
			stmt:  &r.fetchUserByIdStmt,
			query: "SELECT id, name, avatar_url, created_at FROM users WHERE id = $1;",
			name:  "fetch user by id",
		},
		{
			stmt:  &r.insertedMessageStmt,
			query: "INSERT INTO messages (channel_id, user_id, body) VALUES ($1, $2, $3::jsonb) RETURNING id, user_id, channel_id, body, created_at;",
			name:  "insert message",
		},
	}

	for _, s := range statements {
		var err error
		*s.stmt, err = r.db.Prepare(s.query)
		if err != nil {
			log.Errorf("[chat repository error]: error preparing statement %s: %w", s.name, err)
			return err
		}
	}

	return nil
}

func (r *repository) Close() error {
	statements := []*sql.Stmt{
		r.checkMembershipStmt,
		r.fetchUserByIdStmt,
		r.insertedMessageStmt,
	}

	for _, statement := range statements {
		if statement != nil {
			if err := statement.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *repository) CheckUserMembership(channelId int, userId int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS (SELECT 1 FROM memberships WHERE channel_id = $1 AND user_id = $2)", channelId, userId).Scan(&exists)
	return exists, err
}

func (r *repository) FetchUserById(userId int) (entities.User, error) {
	user := entities.User{}
	err := r.db.QueryRow("SELECT id, name, avatar_url, created_at FROM users WHERE id = $1", userId).Scan(&user.ID, &user.Name, &user.AvatarURL, &user.CreatedAt)
	return user, err
}

func (r *repository) InsertMessage(channelId int, userId int, msgBody []byte) (entities.Message, error) {
	insertedMessage := entities.Message{}
	err := r.db.QueryRow("INSERT INTO messages (channel_id, user_id, body) VALUES ($1, $2, $3::jsonb) RETURNING id, user_id, channel_id, body, created_at", channelId, userId, msgBody).Scan(&insertedMessage.ID, &insertedMessage.UserID, &insertedMessage.ChannelID, &insertedMessage.Body, &insertedMessage.CreatedAt)
	return insertedMessage, err
}
