package specs

import (
	"testing"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
)

func TestUserSignupRequestValidate(t *testing.T) {
	testCases := []struct {
		Name          string
		Req           UserSignupRequest
		ExpectedError error
	}{
		{
			Name:          "empty request",
			Req:           UserSignupRequest{},
			ExpectedError: errors.ErrInvalidBody,
		},
		{
			Name: "valid request",
			Req: UserSignupRequest{
				Name:     "something",
				Email:    "test@example.com",
				Password: "password!123",
			},
			ExpectedError: nil,
		},
		{
			Name: "missing name field",
			Req: UserSignupRequest{
				Email:    "test@gmail.com",
				Password: "testingP",
			},
			ExpectedError: errors.ErrMissingNameInRequest,
		},
		{
			Name: "missing email field",
			Req: UserSignupRequest{
				Name:     "test",
				Password: "123456",
			},
			ExpectedError: errors.ErrMissingEmailInRequest,
		},
		{
			Name: "missing password field",
			Req: UserSignupRequest{
				Name:  "test",
				Email: "test@gmail.co,",
			},
			ExpectedError: errors.ErrMissingPasswordInRequest,
		},
		{
			Name: "too short password",
			Req: UserSignupRequest{
				Name:     "test",
				Email:    "test@gmail.com",
				Password: "123456",
			},
			ExpectedError: errors.ErrInvalidBody,
		},
		{
			Name: "invalid email",
			Req: UserSignupRequest{
				Name:     "test",
				Email:    "something-invalid",
				Password: "anything",
			},
			ExpectedError: errors.ErrInvalidEmail,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Req.Validate()
			if err != tc.ExpectedError {
				t.Errorf("Expected Error: %v, Got: %v\n", tc.ExpectedError, err)
			}
		})
	}
}

func TestUserLoginRequestValidate(t *testing.T) {
	testCases := []struct {
		Name          string
		Req           UserLoginRequest
		ExpectedError error
	}{
		{
			Name:          "empty request",
			Req:           UserLoginRequest{},
			ExpectedError: errors.ErrInvalidBody,
		},
		{
			Name: "valid request",
			Req: UserLoginRequest{
				Email:    "test@gmail.com",
				Password: "password",
			},
			ExpectedError: nil,
		},
		{
			Name: "missing email in request",
			Req: UserLoginRequest{
				Password: "password123",
			},
			ExpectedError: errors.ErrMissingEmailInRequest,
		},
		{
			Name: "missing password in request",
			Req: UserLoginRequest{
				Email: "test@gmail.com",
			},
			ExpectedError: errors.ErrMissingPasswordInRequest,
		},
		{
			Name: "invalid email in request",
			Req: UserLoginRequest{
				Email:    "something",
				Password: "password123",
			},
			ExpectedError: errors.ErrInvalidEmail,
		},
		{
			Name: "too short password",
			Req: UserLoginRequest{
				Email:    "test@gmail.com",
				Password: "123456",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Req.Validate()
			if err != tc.ExpectedError {
				t.Errorf("Expected Error: %v, Got: %v\n", tc.ExpectedError, err)
			}
		})
	}
}

func TestCreateTransactionRequestValidate(t *testing.T) {
	testCases := []struct {
		Name          string
		Req           CreateTransactionRequest
		ExpectedError error
	}{
		{
			Name: "valid request",
			Req: CreateTransactionRequest{
				Amount: 1000.0,
				Mode:   "UPI",
			},
			ExpectedError: nil,
		},
		{
			Name:          "empty request",
			Req:           CreateTransactionRequest{},
			ExpectedError: errors.ErrInvalidBody,
		},
		{
			Name: "negative amount",
			Req: CreateTransactionRequest{
				Amount: -500.0,
				Mode:   "NETBANKING",
			},
			ExpectedError: errors.ErrAmountOutOfRange,
		},
		{
			Name: "too large amount",
			Req: CreateTransactionRequest{
				Amount: 1e308,
				Mode:   "NETBANKING",
			},
			ExpectedError: errors.ErrAmountOutOfRange,
		},
		{
			Name: "amount missing in transaction request",
			Req: CreateTransactionRequest{
				Mode: "UPI",
			},
			ExpectedError: errors.ErrMissingAmountInRequest,
		},
		{
			Name: "mode missing in transaction request",
			Req: CreateTransactionRequest{
				Amount: 50.00,
			},
			ExpectedError: errors.ErrMissingModeInRequest,
		},
		{
			Name: "invalid mode",
			Req: CreateTransactionRequest{
				Amount: 500.0,
				Mode:   "invalid",
			},
			ExpectedError: errors.ErrInvalidPaymentMode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Req.Validate()
			if err != tc.ExpectedError {
				t.Errorf("Expected Error: %v, Got: %v\n", tc.ExpectedError, err)
			}
		})
	}
}
