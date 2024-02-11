package entities

type LoginInput struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type SignupInput struct {
	Name      string `json:"name"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

type UserData struct {
	ID       int64  `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
