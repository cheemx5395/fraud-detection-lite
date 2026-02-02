package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/redis/go-redis/v9"
)

// Signup returns an HTTP handler that signs up user using DB
func Signup(DB *repository.Queries) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeUserSignupRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, errors.ErrInvalidBody)
			return
		}

		hashedPass, err := helpers.HashPassword(req.Password)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		userParams := repository.CreateUserParams{
			Name:         req.Name,
			Email:        req.Email,
			MobileNumber: req.Mobile,
			HashedPass:   hashedPass,
		}

		user, err := DB.CreateUser(r.Context(), userParams)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusConflict, err)
			return
		}

		res := specs.UserSignupResponse{
			Message: "Signup Success!",
			ID:      user.ID,
		}

		middleware.SuccessResponse(w, 201, res)
	}
}

// Login returns an HTTP handler that logs the user into the system
func Login(DB *repository.Queries) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeUserLoginRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		if req.Email == "" || req.Password == "" {
			middleware.ErrorResponse(w, http.StatusNotFound, errors.ErrInvalidBody)
			return
		}

		user, err := DB.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusNotFound, errors.ErrUserNotFound)
			return
		}

		err = helpers.CheckPasswordHash(req.Password, user.HashedPass)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, errors.ErrWrongPassword)
			return
		}

		token, err := helpers.MakeJWT(user.ID, user.Name, user.Email, os.Getenv("JWT_SECRET"), constants.TokenExpiryDuration)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, errors.ErrGenerateToken)
			return
		}

		res := specs.UserLoginResponse{
			Message: "Logged in Successfully",
			Token:   token,
		}

		middleware.SuccessResponse(w, http.StatusOK, res)
	}
}

func Logout(DB *repository.Queries, RD *redis.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := helpers.GetClaimsFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, errors.ErrInvalidToken)
			return
		}

		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl <= 0 {
			middleware.ErrorResponse(w, http.StatusUnauthorized, errors.ErrExpiredToken)
			return
		}

		err = RD.Set(
			r.Context(),
			"blacklist:"+claims.ID,
			"true",
			ttl,
		).Err()

		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, errors.ErrLogoutFailed)
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	}
}
