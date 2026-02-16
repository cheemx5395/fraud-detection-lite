package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
)

// decode the signup request
func decodeUserSignupRequest(r *http.Request) (specs.UserSignupRequest, error) {
	var req specs.UserSignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return specs.UserSignupRequest{}, errors.ErrInvalidBody
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	return req, nil
}

// decode the login request
func decodeUserLoginRequest(r *http.Request) (specs.UserLoginRequest, error) {
	var req specs.UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return specs.UserLoginRequest{}, err
	}
	req.Email = strings.TrimSpace(req.Email)
	return req, nil
}

// decode the transaction request
func decodeCreateTransaction(r *http.Request) (specs.CreateTransactionRequest, error) {
	var req specs.CreateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return specs.CreateTransactionRequest{}, err
	}
	return req, nil
}
