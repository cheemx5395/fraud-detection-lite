package helpers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
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

func MakeJWT(userID int32, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		ID:        uuid.NewString(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((expiresIn))),
		Subject:   strconv.FormatInt(int64(userID), 10),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func GetClaimsFromRequest(r *http.Request) (*jwt.RegisteredClaims, error) {
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
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil || !token.Valid {
		if !token.Valid {
			return nil, errors.ErrInvalidToken
		}
		return nil, err
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	return claims, nil
}

func GetIDFromRequest(r *http.Request) (int32, error) {
	claims, err := GetClaimsFromRequest(r)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil {
		return 0, errors.ErrInternalService
	}
	return int32(id), nil
}
