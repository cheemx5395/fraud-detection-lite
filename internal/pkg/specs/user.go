package specs

import (
	"time"
)

// UserSignupRequest struct represents a request to create a user
type UserSignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
}

// User struct represents details of a user profile.
type UserSignupResponse struct {
	Message   string    `json:"message"`
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Mobile    string    `json:"mobile"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserLoginRequest struct represents a request to log-in the user
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginResponse struct represents response to send to successful login of user
type UserLoginResponse struct {
	Message string `json:"message"`
	ID      int32  `json:"id"`
	Token   string `json:"token"`
}
