package api

import (
	"context"
	"net/http"

	"github.com/cheemx5395/fraud-detection-lite/internal/api/handler"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewRouter(ctx context.Context, DB *repository.Queries, RD *redis.Client) *mux.Router {
	router := mux.NewRouter()

	// user registration/login routes
	router.HandleFunc("/signup", handler.Signup(ctx, DB)).Methods(http.MethodPost)
	router.HandleFunc("/login", handler.Login(ctx, DB)).Methods(http.MethodPost)

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(RD))

	protected.HandleFunc("/logout", handler.Logout(ctx, DB, RD)).Methods(http.MethodPost)

	return router
}
