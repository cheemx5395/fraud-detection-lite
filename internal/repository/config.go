package repository

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines configuration of pgxpool
func Config() *pgxpool.Config {

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DB_URI"))
	if err != nil {
		log.Printf("error: failed while configuring: %v", err)
		return nil
	}

	dbConfig.PrepareConn = func(ctx context.Context, conn *pgx.Conn) (bool, error) {
		log.Println("Acquiring connection")
		return true, nil
	}

	dbConfig.AfterRelease = func(conn *pgx.Conn) bool {
		log.Println("Releasing connection")
		return true
	}

	dbConfig.BeforeClose = func(conn *pgx.Conn) {
		log.Println("Closing connection")
	}
	return dbConfig
}
