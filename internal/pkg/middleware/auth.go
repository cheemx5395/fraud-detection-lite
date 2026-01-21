package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(RD *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ErrorResponse(w, http.StatusUnauthorized, errors.ErrEmptyToken)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				ErrorResponse(w, http.StatusUnauthorized, errors.ErrInvalidToken)
				return
			}

			claims, err := helpers.GetClaimsFromToken(tokenString)
			if err != nil {
				ErrorResponse(w, http.StatusUnauthorized, errors.ErrInvalidToken)
				return
			}

			// Check blacklist
			jti := claims.ID
			exists, err := RD.Exists(r.Context(), "blacklist:"+jti).Result()
			if err != nil {
				// Fail closed: if Redis is down, deny access
				ErrorResponse(w, http.StatusInternalServerError, errors.ErrAuthServiceUnavailable)
				return
			}

			if exists == 1 {
				ErrorResponse(w, http.StatusUnauthorized, errors.ErrExpiredToken)
				return
			}

			// Attach user identity to request context
			ctx := context.WithValue(r.Context(), ContextUserIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
