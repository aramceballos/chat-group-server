package entities

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}
