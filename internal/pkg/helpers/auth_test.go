package helpers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHelpers(t *testing.T) {
	password := "mypassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = CheckPasswordHash(password, hash)
	assert.NoError(t, err)

	err = CheckPasswordHash("wrongpassword", hash)
	assert.Error(t, err)
}

func TestJWTHelpers(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	userID := int32(1)
	userName := "Test User"
	email := "test@example.com"

	token, err := MakeJWT(userID, userName, email, "testsecret", time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	claims, err := GetClaimsFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, userName, claims.Name)
}

func TestGetIDFromRequest(t *testing.T) {
	t.Run("From context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), "user_id", int32(100))
		req = req.WithContext(ctx)

		id, err := GetIDFromRequest(req)
		assert.NoError(t, err)
		assert.Equal(t, int32(100), id)
	})

	t.Run("From token", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "testsecret")
		token, _ := MakeJWT(1, "User", "u@e.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		id, err := GetIDFromRequest(req)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), id)
	})
}
