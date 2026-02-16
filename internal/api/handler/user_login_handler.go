package handler

import (
	"context"
	"errors"
	"net/http"

	pkgerrors "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
)

type userServiceInterface interface {
	Signup(ctx context.Context, req specs.UserSignupRequest) (specs.UserSignupResponse, error)
	Login(ctx context.Context, req specs.UserLoginRequest) (specs.UserLoginResponse, error)
	Logout(ctx context.Context, claims *specs.UserTokenClaims) error
}

// Signup returns an HTTP handler that signs up user using DB
func Signup(s userServiceInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.ErrorResponse(w, http.StatusMethodNotAllowed, pkgerrors.ErrMethodNotAllowed)
		}

		req, err := decodeUserSignupRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		if err := req.Validate(); err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		res, err := s.Signup(r.Context(), req)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		middleware.SuccessResponse(w, http.StatusCreated, res)
	}
}

// Login returns an HTTP handler that logs the user into the system
func Login(s userServiceInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.ErrorResponse(w, http.StatusMethodNotAllowed, pkgerrors.ErrMethodNotAllowed)
		}

		req, err := decodeUserLoginRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		if err := req.Validate(); err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		res, err := s.Login(r.Context(), req)
		if err != nil {
			if errors.Is(err, pkgerrors.ErrUserNotFound) {
				middleware.ErrorResponse(w, http.StatusNotFound, err)
			} else if errors.Is(err, pkgerrors.ErrWrongPassword) {
				middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			} else {
				middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			}
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, res)
	}
}

func Logout(s userServiceInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := helpers.GetClaimsFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}

		err = s.Logout(r.Context(), claims)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	}
}
