package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/api"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/cheemx5395/fraud-detection-lite/internal/service"
	"github.com/cheemx5395/fraud-detection-lite/internal/worker"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// setup env
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading environment")
		return
	}

	logger.Info("Starting Server...")
	defer logger.Info("Shutting Down Server...")

	// Initialize DB
	DB, db, err := repository.InitializeDatabase(ctx)
	if err != nil {
		logger.Error("Database Error", zap.Error(err))
		return
	}
	logger.Info("Connected to Database!")
	defer db.Close()

	// Initialize Redis
	RD := worker.InitializeRedis()
	logger.Info("Connected to Redis")
	defer RD.Close()

	// Initialize Services
	txnService := service.NewTransactionService(DB, db, logger)
	userService := service.NewUserService(DB, RD, logger)

	// Initializing Router
	router := api.NewRouter(DB, RD, txnService, userService, logger)

	// CORS middleware
	corsOptions := cors.New(constants.CorsOptions)

	// Setup the server
	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: corsOptions.Handler(router),
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting Server", zap.Error(err))
		}
	}()

	// cron-job to update profile behavior added
	updater := worker.NewProfileUpdater(DB)
	cronInstance := updater.Start(ctx)

	// Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	sig := <-signalChan
	logger.Info("Received terminate, gracefully shutting down", zap.Any("signal", sig))

	ctx = cronInstance.Stop()
	<-ctx.Done()

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(tc)
	cancel()
}
