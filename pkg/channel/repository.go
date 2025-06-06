package channel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Repository interface {
	FetchChannels() ([]entities.Channel, error)
	FetchChannelById(id string) (entities.Channel, error)
	CreateChannel(channel entities.CreateChannelInput) error
	JoinChannel(userId string, input entities.JoinChannelInput) error
	LeaveChannel(userId string, input entities.JoinChannelInput) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db,
	}
}

func (r *repository) FetchChannels() ([]entities.Channel, error) {
	rows, err := r.db.Query("SELECT id, name, description, image_url FROM channels")
	if err != nil {
		return nil, err
	}

	var channels []entities.Channel
	for rows.Next() {
		var channel entities.Channel
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.Description, &channel.ImageURL); err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (r *repository) FetchChannelById(channelId string) (entities.Channel, error) {
	query := `
		SELECT
			c.id AS channel_id,
			c.name AS channel_name,
			c.description AS channel_description,
			c.image_url AS channel_image_url,
			m.id AS membership_id,
			m.user_id AS membership_user_id,
			m.role AS membership_role,
			COALESCE(u1.name, '') AS membership_user_name,
			COALESCE(u1.avatar_url, '') AS membership_user_avatar_url,
			COALESCE(u1.created_at::text, '') AS membership_user_created_at,
			msg.id AS message_id,
			msg.user_id AS message_user_id,
			COALESCE(u2.name, '') AS message_user_name,
			COALESCE(u2.avatar_url, '') AS message_user_avatar_url,
			COALESCE(u2.created_at::text, '') AS message_user_created_at,
			msg.content AS message_content,
			msg.created_at AS message_created_at
		FROM channels c
		LEFT JOIN memberships m ON c.id = m.channel_id
		LEFT JOIN users u1 ON m.user_id = u1.id
		LEFT JOIN messages msg ON c.id = msg.channel_id
		LEFT JOIN users u2 ON msg.user_id = u2.id
		WHERE c.id = $1
	`

	rows, err := r.db.Query(query, channelId)
	if err != nil {
		return entities.Channel{}, err
	}
	defer rows.Close()

	// Variables to store channel, memberships, and messages
	var channel entities.Channel
	memberships := make(map[int64]entities.Membership) // Map to store memberships by ID
	messages := make(map[int]entities.Message)         // Map to store messages by ID

	// Iterate through the rows
	for rows.Next() {
		var (
			channelID, messageID                                                                               int
			membershipID, membershipUserID                                                                     sql.NullInt64
			channelName, channelDescription, channelImageURL                                                   string
			messageUserID                                                                                      int64
			membershipRole, membershipUserName, membershipUserAvatarURL, messageUserAvatarURL, messageUserName string
			membershipUserCreatedAt, messageUserCreatedAt, messageCreatedAt                                    string
			messageBody                                                                                        json.RawMessage
		)

		// Scan the values into variables
		if err := rows.Scan(
			&channelID, &channelName, &channelDescription, &channelImageURL,
			&membershipID, &membershipUserID, &membershipRole, &membershipUserName, &membershipUserAvatarURL, &membershipUserCreatedAt,
			&messageID, &messageUserID, &messageUserName, &messageUserAvatarURL, &messageUserCreatedAt,
			&messageBody, &messageCreatedAt,
		); err != nil {
			return entities.Channel{}, err
		}

		// Check if the channel has already been populated
		if channel.ID == 0 {
			// Populate channel information if not done yet
			channel.ID = channelID
			channel.Name = channelName
			channel.Description = channelDescription
			channel.ImageURL = channelImageURL
		}

		// Check if the membership and user ID are valid
		if membershipID.Valid && membershipUserID.Valid {
			// Check if the membership has already been populated
			if _, ok := memberships[membershipID.Int64]; !ok {
				// Populate membership information if not done yet
				memberships[membershipID.Int64] = entities.Membership{
					ID:        membershipID.Int64,
					UserID:    membershipUserID.Int64,
					ChannelID: channelID,
					User: entities.User{
						ID:        membershipUserID.Int64,
						Name:      membershipUserName,
						AvatarURL: membershipUserAvatarURL,
						Role:      membershipRole,
						CreatedAt: membershipUserCreatedAt,
					},
				}
			}
		}

		// Check if the message has already been populated
		if _, ok := messages[messageID]; !ok {
			// Populate message information if not done yet
			messages[messageID] = entities.Message{
				ID:        messageID,
				UserID:    messageUserID,
				ChannelID: channelID,
				Body:      messageBody, // Use messageBody variable, which should be of type json.RawMessage
				CreatedAt: messageCreatedAt,
				User: entities.User{
					ID:        messageUserID,
					Name:      messageUserName,
					AvatarURL: messageUserAvatarURL,
					CreatedAt: messageUserCreatedAt,
				},
			}
		}
	}

	// Populate memberships and messages into the channel
	for _, membership := range memberships {
		channel.Members = append(channel.Members, membership.User)
	}

	for _, message := range messages {
		channel.Messages = append(channel.Messages, message)
	}

	return channel, nil
}

func (r *repository) CreateChannel(channel entities.CreateChannelInput) error {
	_, err := r.db.Exec("INSERT INTO channels (name, description, image_url) VALUES ($1, $2, $3)", channel.Name, channel.Description, channel.ImageURL)
	return err
}

func (r *repository) JoinChannel(userId string, input entities.JoinChannelInput) error {
	var existingMembershipID int64
	err := r.db.QueryRow("SELECT id FROM memberships WHERE user_id = $1 AND channel_id = $2", userId, input.ChannelID).Scan(&existingMembershipID)
	fmt.Println("id:", existingMembershipID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingMembershipID != 0 {
		return errors.New("membership already exists")
	}
	_, err = r.db.Exec("INSERT INTO memberships (user_id, channel_id, role) VALUES ($1, $2, 'user')", userId, input.ChannelID)
	return err
}

func (r *repository) LeaveChannel(userId string, input entities.JoinChannelInput) error {
	_, err := r.db.Exec("DELETE FROM memberships WHERE user_id = $1 AND channel_id = $2", userId, input.ChannelID)
	return err
}
