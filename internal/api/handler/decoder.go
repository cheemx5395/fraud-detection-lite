package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
)

// decode the signup request
func decodeUserSignupRequest(r *http.Request) (specs.UserSignupRequest, error) {
	var req specs.UserSignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error decoding signup request: %v\n", err)
		return specs.UserSignupRequest{}, errors.ErrInvalidBody
	}

	return req, nil
}

// decode the login request
func decodeUserLoginRequest(r *http.Request) (specs.UserLoginRequest, error) {
	var req specs.UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error decoding login request: %v\n", err)
		return specs.UserLoginRequest{}, err
	}
	return req, nil
}
