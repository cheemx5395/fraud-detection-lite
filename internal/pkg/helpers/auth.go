package helpers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashedPass), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID int32, userName, email, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := specs.UserTokenClaims{
		UserID: userID,
		Name:   userName,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func GetClaimsFromRequest(r *http.Request) (*specs.UserTokenClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.ErrEmptyToken
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, errors.ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&specs.UserTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.ErrInvalidToken
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*specs.UserTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	return claims, nil
}

func GetIDFromRequest(r *http.Request) (int32, error) {
	// Try to get from context first (for testing/middleware)
	if uid, ok := r.Context().Value("user_id").(int32); ok {
		return uid, nil
	}

	claims, err := GetClaimsFromRequest(r)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
