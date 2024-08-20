package pgPackage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

func GetTablesForSchema(pool *pgxpool.Pool, schemaName string, table string) []string {
	log.Infof("Getting tables for schema %s",schemaName)

    // SQL query to retrieve the list of tables for a specific schema
    query := `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = $1 AND table_type = 'BASE TABLE'
    `

	// Execute the query
	var tables []string
	
	if table != "" {
		query += " AND table_name = '" + table + "'"
	}

    rows, err := pool.Query(context.Background(), query, schemaName)

	if err != nil {
		log.Fatalf("Error retreiving list of tables %s",err)
	}

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatalf("Error parsing list of tables %s",err)
		}

		log.Infof("Discovered table %s", tableName)

		tables = append(tables, tableName)
	}

    return tables
}