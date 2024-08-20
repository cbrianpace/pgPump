package main

import (
	"flag"
	"fmt"
	"os"
	pgPackage "pgPump/internal/postgres"
	"time"

	log "github.com/sirupsen/logrus"
)

// CustomFormatter implements logrus.Formatter
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
    timestamp := entry.Time.Format(time.RFC3339)
    logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, entry.Level.String(), entry.Message)
    return []byte(logLine), nil
}

func main() {
	setupLogging()

	action, connStr, args := parseFlags()

	pool := pgPackage.CreatePool(connStr)
	
	switch action {

		case "export":
			log.Infof("Exporting tables from schema %s to directory %s using format %s",args.schema, args.dir, args.format)
			pgPackage.Export(args.dir, args.format, pool, args.schema, args.table, args.columns, args.threads)
		case "import":
			log.Infof("Importing from files in %s",args.dir)
			pgPackage.Import(args.dir, pool, args.schema, args.table, args.columns, args.threads)
		case "version":
			fmt.Println("Version: 0.1.0")
		default:
			log.Fatal("Expected export or import for command")
			os.Exit(1)
	}
}

func setupLogging() {
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(true)
	log.SetFormatter(&CustomFormatter{})
}

type commandArgs struct {
	user, password, host, database, sslmode, dir, table, schema, format, columns string
	port, threads                                                                int
}

func parseFlags() (string, string, commandArgs) {
	const (
		defaultUser     = "postgres"
		defaultPassword = ""
		defaultHost     = "localhost"
		defaultDatabase = "postgres"
		defaultSSLMode  = "disable"
		defaultDir      = "."
		defaultFormat   = "binary"
		defaultPort     = 5432
		defaultSchema   = "public"
		defaultThread   = 1
	)

	actionCmd := flag.NewFlagSet("action", flag.ExitOnError)

	user := actionCmd.String("user", defaultUser, "Username for database connection")
	password := actionCmd.String("password", defaultPassword, "Database password")
	host := actionCmd.String("host", defaultHost, "Host of the database")
	database := actionCmd.String("database", defaultDatabase, "Database name")
	sslmode := actionCmd.String("sslmode", defaultSSLMode, "Postgres SSL Mode")
	dir := actionCmd.String("dir", defaultDir, "Directory to process")
	table := actionCmd.String("table", "", "Table to load or export")
	schemaName := actionCmd.String("schema", defaultSchema, "Schema of table(s) to load or export")
	format := actionCmd.String("format", defaultFormat, "Format of dump file (binary or csv)")
	columns := actionCmd.String("columns", "all", "Comma-separated list of columns to export")
	port := actionCmd.Int("port", defaultPort, "Port number for database connection")
	threadLimit := actionCmd.Int("parallel", defaultThread, "Number of concurrent threads to perform exports")


	// Ensure at least one argument is passed for action
	if len(os.Args) < 2 {
		log.Fatal("Expected 'export', 'import', or 'version' command")
		os.Exit(1)
	}

	if err := actionCmd.Parse(os.Args[2:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	args := commandArgs{
		user:     *user,
		password: *password,
		host:     *host,
		database: *database,
		sslmode:  *sslmode,
		dir:      *dir,
		table:    *table,
		schema:   *schemaName,
		format:   *format,
		columns:  *columns,
		port:     *port,
		threads:  *threadLimit,
	}

	if args.password == "" {
		args.password = os.Getenv("PGPASSWORD")
	}

	if args.columns != "all" {
		args.columns = fmt.Sprintf("(%s)", args.columns)
	} else {
		args.columns = ""
	}

	action := os.Args[1]

	connStr := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=%s user=%s password=%s", args.host, args.port, args.database, args.sslmode, args.user, args.password)

	return action, connStr, args
}