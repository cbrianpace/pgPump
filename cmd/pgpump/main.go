package main

import (
	//"database/sql"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	//_ "github.com/lib/pq"
)

func main() {
	// Define command-line flags
	actionCmd := flag.NewFlagSet("export", flag.ExitOnError)
	user := actionCmd.String("user", "postgres", "Username for database connection")
	password := actionCmd.String("password","","Database password")
	host := actionCmd.String("host", "localhost", "Host of the database")
	port := actionCmd.Int("port", 5432, "Port number for database connection")
	database := actionCmd.String("database", "postgres", "Database name")
	sslmode := actionCmd.String("sslmode","disable","Postgres SSL Mode")
	file := actionCmd.String("file", "output.bin", "File to process")
	table := actionCmd.String("table", "", "Table to load or export")
	format := actionCmd.String("format", "binary","Format of dump file.  Options are binary or csv.")
	columns := actionCmd.String("columns","all","Comma seperated list of columns to export.")

	// Parse the command-line flags
	actionCmd.Parse(os.Args[2:])

	// Get password from environment variable if not set
	var passwd string = os.Getenv("PGPASSWORD")
	if  len(*password) > 0 {
		passwd = *password
	}

	var columnList string = ""
	if *columns != "all" {
		columnList = fmt.Sprintf("(%s)",*columns)
	}

	// PostgreSQL connection information
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=%s user=%s password=%s",*host,*port,*database,*sslmode,*user, passwd)

	pool := pgConnection.createPool(connStr)

	// Initialize connection pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()
	
	switch os.Args[1] {

		case "export":
			// Open a file to write the binary data
			outputFile, err := os.Create(*file)
			if err != nil {
				log.Fatalf("Failed to create file: %v\n", err)
			}
			defer outputFile.Close()
			
			sqlQuery := fmt.Sprintf("COPY %s %s TO stdout WITH %s",*table,columnList,*format)

			// Execute the COPY TO STDOUT command
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				log.Fatalf("Failed to acquire connection: %v\n", err)
			}
			defer conn.Release()

			_, err = conn.Conn().PgConn().CopyTo(context.Background(), outputFile, sqlQuery)
			if err != nil {
				log.Fatalf("Failed to execute COPY TO STDOUT: %v\n", err)
			}

			fmt.Println("Binary data has been saved to ",*file)
		case "version":
			fmt.Println("Version: 0.1.0")
		default:
			fmt.Println("Expected export or import for command")
			os.Exit(1)
	}
}
