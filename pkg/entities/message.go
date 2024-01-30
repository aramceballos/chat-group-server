package entities

type Message struct {
	ID        int    `json:"id"`
	UserID    int64  `json:"user_id"`
	ChannelID int    `json:"channel_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user,omitempty"`
}
