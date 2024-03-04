package entities

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at"`
}

type UpdateUserInput struct {
	Name     string `json:"name" validate:"required,min=5" error:"name is required"`
	Username string `json:"username" validate:"required,min=5,max=20" error:"username is required"`
	Email    string `json:"email" validate:"required,min=5" error:"email is required"`
}
