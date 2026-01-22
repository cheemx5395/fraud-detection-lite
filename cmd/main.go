package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/api"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/cheemx5395/fraud-detection-lite/internal/worker"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	ctx := context.Background()

	// setup env
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading environment")
		return
	}

	fmt.Println("Starting Server...")
	defer fmt.Println("Shutting Down Server...")

	// Initialize DB
	DB, db, err := repository.InitializeDatabase(ctx)
	if err != nil {
		log.Printf("Database Error: %v", err)
		return
	}
	fmt.Println("Connected to Database!")
	defer db.Close()

	// Initialize Redis
	RD := worker.InitializeRedis()
	fmt.Println("Connected to Redis")
	defer RD.Close()

	// Initializing Router
	router := api.NewRouter(DB, RD)

	// CORS middleware
	cors := cors.New(constants.CorsOptions)

	// Setup the server
	server := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: cors.Handler(router),
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting Server: ", err)
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
	fmt.Println("Received terminate, gracefully shutting down:", sig)

	ctx = cronInstance.Stop()
	<-ctx.Done()

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(tc)
	cancel()
}
