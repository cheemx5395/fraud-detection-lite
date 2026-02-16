package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRedis struct {
	mock.Mock
}

func (m *mockRedis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	// Variadic arguments handling for testify/mock
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func TestAuthMiddleware(t *testing.T) {
	mockRD := new(mockRedis)
	mockNext := new(mockHandler)
	os.Setenv("JWT_SECRET", "testsecret")

	middleware := AuthMiddleware(mockRD)(mockNext)

	t.Run("Success - valid token", func(t *testing.T) {
		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Mock redis.Exists returning 0 (not blacklisted)
		cmd := redis.NewIntCmd(context.Background())
		cmd.SetVal(0)
		mockRD.On("Exists", mock.Anything, mock.Anything).Return(cmd).Once()

		mockNext.On("ServeHTTP", w, mock.Anything).Return().Once()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRD.AssertExpectations(t)
		mockNext.AssertExpectations(t)
	})

	t.Run("Failure - blacklisted token", func(t *testing.T) {
		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Mock redis.Exists returning 1 (blacklisted)
		cmd := redis.NewIntCmd(context.Background())
		cmd.SetVal(1)
		mockRD.On("Exists", mock.Anything, mock.Anything).Return(cmd).Once()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockRD.AssertExpectations(t)
	})

	t.Run("Failure - missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Failure - invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
