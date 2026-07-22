package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// GetTablesForSchema returns the base tables in the given schema. When table is
// non-empty, only that table is returned (if it exists).
func GetTablesForSchema(ctx context.Context, pool *pgxpool.Pool, schemaName string, table string) ([]string, error) {
	log.Infof("Getting tables for schema %s", schemaName)

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'`
	args := []any{schemaName}

	if table != "" {
		query += " AND table_name = $2"
		args = append(args, table)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("retrieving list of tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("parsing list of tables: %w", err)
		}
		log.Infof("Discovered table %s", tableName)
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating list of tables: %w", err)
	}

	return tables, nil
}
