package specs

import (
	"regexp"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

// UserSignupRequest struct represents a request to create a user
type UserSignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserSignupRequest) Validate() error {
	switch {
	case r.Email == "" && r.Name == "" && r.Password == "":
		return errors.ErrInvalidBody
	case r.Email == "":
		return errors.ErrMissingEmailInRequest
	case r.Name == "":
		return errors.ErrMissingNameInRequest
	case r.Password == "":
		return errors.ErrMissingPasswordInRequest
	}

	emailRegex := regexp.MustCompile(constants.EmailRegex)
	if !emailRegex.MatchString(r.Email) {
		return errors.ErrInvalidEmail
	}

	if len(r.Password) < 8 {
		return errors.ErrInvalidBody
	}

	return nil
}

// UserSignupResponse to represent signup response
type UserSignupResponse struct {
	Message string `json:"message"`
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

// User struct represents details of a user profile.
type UserResponse struct {
	Message   string    `json:"message"`
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserLoginRequest struct represents a request to log-in the user
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserLoginRequest) Validate() error {
	switch {
	case r.Email == "" && r.Password == "":
		return errors.ErrInvalidBody
	case r.Email == "":
		return errors.ErrMissingEmailInRequest
	case r.Password == "":
		return errors.ErrMissingPasswordInRequest
	}

	emailRegex := regexp.MustCompile(constants.EmailRegex)
	if !emailRegex.MatchString(r.Email) {
		return errors.ErrInvalidEmail
	}

	return nil
}

// UserLoginResponse struct represents response to send to successful login of user
type UserLoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type UserTokenClaims struct {
	UserID int32  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
