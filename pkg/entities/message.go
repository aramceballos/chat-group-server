package entities

import "encoding/json"

type Message struct {
	ID        int             `json:"id"`
	UserID    int64           `json:"user_id"`
	ChannelID int             `json:"channel_id"`
	Body      json.RawMessage `json:"body"`
	CreatedAt string          `json:"created_at"`
	User      User            `json:"user,omitempty"`
}
