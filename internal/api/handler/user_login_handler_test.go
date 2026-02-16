package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	pkgerrors "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSignup(t *testing.T) {
	t.Run("invalid request method", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Signup(mockService)
		req := httptest.NewRequest(http.MethodGet, "/signup", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Success signup", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Signup(mockService)
		reqBody := specs.UserSignupRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Signup", mock.Anything, reqBody).Return(specs.UserSignupResponse{
			Message: "Signup Success!",
			ID:      1,
			Name:    "Test User",
			Email:   "test@example.com",
		}, nil).Once()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Signup Success!", response["data"].(map[string]any)["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid body", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Signup(mockService)
		req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(http.StatusBadRequest), response["error_code"])
	})

	t.Run("leading and trailing spaces in name in signup request", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Signup(mockService)
		reqBody := specs.UserSignupRequest{
			Name:     "                        1                                ",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Signup", mock.Anything, mock.MatchedBy(func(r specs.UserSignupRequest) bool {
			return r.Name == "1"
		})).Return(specs.UserSignupResponse{
			Message: "Signup Success!",
			ID:      1,
			Name:    "1",
			Email:   "test@example.com",
		}, nil).Once()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Signup Success!", response["data"].(map[string]any)["message"])
		assert.Equal(t, "1", response["data"].(map[string]any)["name"])
		mockService.AssertExpectations(t)
	})

	t.Run("leading and trailing spaces in email in signup request", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Signup(mockService)
		reqBody := specs.UserSignupRequest{
			Name:     "test Name",
			Email:    "                                 test@example.com            ",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Signup", mock.Anything, mock.MatchedBy(func(r specs.UserSignupRequest) bool {
			return r.Email == "test@example.com"
		})).Return(specs.UserSignupResponse{
			Message: "Signup Success!",
			ID:      1,
			Name:    "test Name",
			Email:   "test@example.com",
		}, nil).Once()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Signup Success!", response["data"].(map[string]any)["message"])
		assert.Equal(t, "test@example.com", response["data"].(map[string]any)["email"])
		mockService.AssertExpectations(t)
	})

	testCases := []struct {
		Name    string
		Req     specs.UserSignupRequest
		ResErr  error
		ResCode int
	}{
		{
			Name: "missing name",
			Req: specs.UserSignupRequest{
				Email:    "test@gmail.com",
				Password: "password123",
			},
			ResErr:  pkgerrors.ErrMissingNameInRequest,
			ResCode: http.StatusBadRequest,
		},
		{
			Name: "missing email",
			Req: specs.UserSignupRequest{
				Name:     "Test User",
				Password: "password123",
			},
			ResErr:  pkgerrors.ErrMissingEmailInRequest,
			ResCode: http.StatusBadRequest,
		},
		{
			Name: "missing password",
			Req: specs.UserSignupRequest{
				Name:  "testing name",
				Email: "email@email.com",
			},
			ResErr:  pkgerrors.ErrMissingPasswordInRequest,
			ResCode: http.StatusBadRequest,
		},
		{
			Name:    "empty request body",
			Req:     specs.UserSignupRequest{},
			ResErr:  pkgerrors.ErrInvalidBody,
			ResCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := Signup(mockService)
			reqBody, _ := json.Marshal(tc.Req)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			handler(w, req)
			var response map[string]any
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tc.ResCode, w.Code)
			assert.Equal(t, tc.ResErr.Error(), response["error_message"])
		})
	}
}

func TestLogin(t *testing.T) {
	t.Run("invalid request method", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Login(mockService)
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Success login", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Login(mockService)
		reqBody := specs.UserLoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Login", mock.Anything, reqBody).Return(specs.UserLoginResponse{
			Message: "Logged in Successfully",
			Token:   "mock-token",
		}, nil).Once()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]any)
		assert.Equal(t, "Logged in Successfully", data["message"])
		assert.Equal(t, "mock-token", data["token"])
		mockService.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Login(mockService)
		reqBody := specs.UserLoginRequest{
			Email:    "notfound@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Login", mock.Anything, reqBody).Return(specs.UserLoginResponse{}, pkgerrors.ErrUserNotFound).Once()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Wrong password", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := Login(mockService)
		reqBody := specs.UserLoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		mockService.On("Login", mock.Anything, reqBody).Return(specs.UserLoginResponse{}, pkgerrors.ErrWrongPassword).Once()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestLogout(t *testing.T) {
	mockService := new(MockUserService)
	handler := Logout(mockService)
	os.Setenv("JWT_SECRET", "testsecret")

	t.Run("Success logout", func(t *testing.T) {
		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		mockService.On("Logout", mock.Anything, mock.Anything).Return(nil).Once()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized - no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, pkgerrors.ErrEmptyToken.Error(), response["error_message"])
	})

	t.Run("Unauthorized - invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
