package errors

import "errors"

// Error Variables for replacing the errors wherever required
var (
	ErrInvalidBody            = errors.New("invalid request body")
	ErrParameterMissing       = errors.New("parameter missing")
	ErrUserNotFound           = errors.New("user with given parameter not found")
	ErrWrongPassword          = errors.New("password is incorrect")
	ErrGenerateToken          = errors.New("unable to generate token")
	ErrEmptyToken             = errors.New("empty token")
	ErrInvalidToken           = errors.New("invlaid token")
	ErrExpiredToken           = errors.New("expired token")
	ErrLogoutFailed           = errors.New("token blacklisting failed")
	ErrAuthServiceUnavailable = errors.New("redis down for authentication")
)

// DB Related variables
var (
	ErrConnectionFailed = errors.New("error connecting to database")
)
