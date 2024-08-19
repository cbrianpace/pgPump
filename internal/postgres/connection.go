package pgConnection

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createPool (connStr string) *pgxpool.Pool {

	// Initialize connection pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()
	
	return pool
}