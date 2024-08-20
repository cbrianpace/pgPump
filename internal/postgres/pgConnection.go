package pgPackage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

func CreatePool (connStr string) *pgxpool.Pool {

	// Initialize connection pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	//defer pool.Close()
	
	return pool
}