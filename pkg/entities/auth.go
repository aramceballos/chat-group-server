package entities

type LoginInput struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type UserData struct {
	ID       int64  `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
