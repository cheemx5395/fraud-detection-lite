package handler

import (
	"context"
	"log"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Signup(ctx context.Context, req specs.UserSignupRequest) (specs.UserSignupResponse, error) {
	args := m.Called(ctx, req)
	log.Println(args...)
	return args.Get(0).(specs.UserSignupResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req specs.UserLoginRequest) (specs.UserLoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(specs.UserLoginResponse), args.Error(1)
}

func (m *MockUserService) Logout(ctx context.Context, claims *specs.UserTokenClaims) error {
	args := m.Called(ctx, claims)
	return args.Error(0)
}
