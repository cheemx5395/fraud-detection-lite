package api

import (
	"net/http"

	"github.com/cheemx5395/fraud-detection-lite/internal/api/handler"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewRouter(DB *repository.Queries, RD *redis.Client) *mux.Router {
	router := mux.NewRouter()

	// user registration/login routes
	router.HandleFunc("/signup", handler.Signup(DB)).Methods(http.MethodPost)
	router.HandleFunc("/login", handler.Login(DB)).Methods(http.MethodPost)

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(RD))

	// user routes
	protected.HandleFunc("/user", handler.GetUser(DB)).Methods(http.MethodGet)

	// Transaction routes
	protected.HandleFunc("/transactions", handler.PostTransaction(DB)).Methods(http.MethodPost)
	protected.HandleFunc("/transactions", handler.GetTransactions(DB)).Methods(http.MethodGet)
	protected.HandleFunc("/transactions/{id}", handler.GetTransaction(DB)).Methods(http.MethodGet)

	// logout handler
	protected.HandleFunc("/logout", handler.Logout(DB, RD)).Methods(http.MethodPost)

	return router
}
