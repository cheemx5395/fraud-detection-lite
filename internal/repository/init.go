package repository

import (
	"context"
	"log"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitializeDatabase function used to initialize the database and returns the database object
func InitializeDatabase(ctx context.Context) (*Queries, *pgxpool.Pool, error) {
	connPool, err := pgxpool.NewWithConfig(ctx, Config())
	if err != nil {
		log.Printf("Failed to create connection pool: %v", err)
		return nil, nil, errors.ErrConnectionFailed
	}

	err = connPool.Ping(ctx)
	if err != nil {
		log.Printf("Failed to ping the database: %v", err)
		return nil, nil, errors.ErrConnectionFailed
	}

	queries := New(connPool)

	return queries, connPool, nil
}
