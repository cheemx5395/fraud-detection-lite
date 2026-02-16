package service

import (
	"context"
	"os"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UserService struct {
	db     *repository.Queries
	rd     *redis.Client
	logger *zap.Logger
}

func NewUserService(queries *repository.Queries, rd *redis.Client, logger *zap.Logger) *UserService {
	return &UserService{
		db:     queries,
		rd:     rd,
		logger: logger,
	}
}

func (s *UserService) Signup(ctx context.Context, req specs.UserSignupRequest) (specs.UserSignupResponse, error) {
	hashedPass, err := helpers.HashPassword(req.Password)
	if err != nil {
		return specs.UserSignupResponse{}, err
	}

	user, err := s.db.CreateUser(ctx, repository.CreateUserParams{
		Name:       req.Name,
		Email:      req.Email,
		HashedPass: hashedPass,
	})
	if err != nil {
		return specs.UserSignupResponse{}, err
	}

	res := specs.UserSignupResponse{
		Message: "Signup Success!",
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
	}

	return res, nil
}

func (s *UserService) Login(ctx context.Context, req specs.UserLoginRequest) (specs.UserLoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return specs.UserLoginResponse{}, errors.ErrUserNotFound
	}

	if err := helpers.CheckPasswordHash(req.Password, user.HashedPass); err != nil {
		return specs.UserLoginResponse{}, errors.ErrWrongPassword
	}

	// generate token
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret" // Fallback or error? defaulting for "lite" version if not set
	}

	// 24 hours expiration
	token, err := helpers.MakeJWT(user.ID, user.Name, user.Email, secret, 24*time.Hour)
	if err != nil {
		return specs.UserLoginResponse{}, err
	}

	return specs.UserLoginResponse{
		Message: "Login Success!",
		Token:   token,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, claims *specs.UserTokenClaims) error {
	// Block/Blacklist the token in Redis
	// Key: "blacklist:<jti>"
	// Expiry: Time remaining until token expires
	timeLeft := time.Until(claims.ExpiresAt.Time)
	if timeLeft < 0 {
		return nil // Already expired
	}

	err := s.rd.Set(ctx, "blacklist:"+claims.ID, "true", timeLeft).Err()
	if err != nil {
		s.logger.Error("failed to blacklist token", zap.Error(err))
		return errors.ErrLogoutFailed
	}

	return nil
}
