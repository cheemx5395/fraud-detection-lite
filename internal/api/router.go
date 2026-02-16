package api

import (
	"net/http"

	"github.com/cheemx5395/fraud-detection-lite/internal/api/handler"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/cheemx5395/fraud-detection-lite/internal/service"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRouter(DB *repository.Queries, RD *redis.Client, txnService *service.TransactionService, userService *service.UserService, logger *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	// user registration/login routes
	router.Use(middleware.LoggerMiddleware)
	router.HandleFunc("/signup", handler.Signup(userService)).Methods(http.MethodPost).Name("signup")
	router.HandleFunc("/login", handler.Login(userService)).Methods(http.MethodPost).Name("login")

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(RD))

	// Transaction routes
	protected.HandleFunc("/transactions", handler.PostTransaction(txnService)).Methods(http.MethodPost)
	protected.HandleFunc("/transactions", handler.GetTransactions(DB)).Methods(http.MethodGet)
	protected.HandleFunc("/transactions/{id}", handler.GetTransaction(DB)).Methods(http.MethodGet)

	// bulk ingestion handlers
	protected.HandleFunc("/transactions/upload", handler.ProcessBulkTransactions(txnService)).Methods(http.MethodPost)

	// logout handler
	protected.HandleFunc("/logout", handler.Logout(userService)).Methods(http.MethodPost)

	return router
}
