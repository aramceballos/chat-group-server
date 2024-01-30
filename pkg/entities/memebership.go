package entities

type Membership struct {
	ID        int64 `json:"id"`
	UserID    int64 `json:"user_id"`
	ChannelID int   `json:"channel_id"`
	User      User  `json:"-"`
}
