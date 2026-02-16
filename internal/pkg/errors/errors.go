package errors

import "errors"

// Error Variables for replacing the errors wherever required
var (
	// Missing "required" parameters in requests
	ErrMissingNameInRequest     = errors.New("missing name in request body")
	ErrMissingEmailInRequest    = errors.New("missing email in request body")
	ErrMissingPasswordInRequest = errors.New("missing password in request body")
	ErrMissingAmountInRequest   = errors.New("missing amount in request body")
	ErrMissingModeInRequest     = errors.New("missing mode in request body")
	ErrMissingFileInRequest     = errors.New("missing file in request body")

	// validation errors on transactions
	ErrAmountOutOfRange   = errors.New("amount should be in range 1 to 10^13")
	ErrInvalidPaymentMode = errors.New("invalid mode to make transaction")

	// handler errors
	ErrMethodNotAllowed        = errors.New("method not allowed")
	ErrInvalidBody             = errors.New("invalid request body")
	ErrInvalidEmail            = errors.New("invalid email formatting")
	ErrUserNotFound            = errors.New("user with given parameter not found")
	ErrWrongPassword           = errors.New("password is incorrect")
	ErrGenerateToken           = errors.New("unable to generate token")
	ErrEmptyToken              = errors.New("empty token")
	ErrInvalidToken            = errors.New("invalid token")
	ErrExpiredToken            = errors.New("expired token")
	ErrLogoutFailed            = errors.New("token blacklisting failed")
	ErrAuthServiceUnavailable  = errors.New("redis down for authentication")
	ErrAuthInternalService     = errors.New("error in auth package")
	ErrInternalService         = errors.New("in-built function returning error")
	ErrTxnBlocked              = errors.New("invalid traansaction blocked")
	ErrDB                      = errors.New("error in operations on db")
	ErrNotFound                = errors.New("Set not found in Redis Server")
	ErrFailureInParsingExcel   = errors.New("failure in parsing excel file")
	ErrFailureInParsingCSV     = errors.New("failure in parsing csv file")
	ErrUnexpectedHeadersInFile = errors.New("unexpected headers in file")
)

// DB Related variables
var (
	ErrConnectionFailed = errors.New("error connecting to database")
)

var (
	ErrBackgroundJobFailed = errors.New("error in profile updating job")
)
