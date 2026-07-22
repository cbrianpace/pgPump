package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	postgres "pgPump/internal/postgres"

	log "github.com/sirupsen/logrus"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

// CustomFormatter renders log entries as "[timestamp] level: message".
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, entry.Level.String(), entry.Message)
	return []byte(logLine), nil
}

type commandArgs struct {
	user, password, host, database, sslmode, dir, table, schema, format, columns string
	port, threads                                                                int
	verbose                                                                      bool
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("expected 'export', 'import', or 'version' command")
	}

	action := os.Args[1]
	switch action {
	case "version", "--version", "-version":
		fmt.Printf("pgPump %s\n", Version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	case "export", "import":
		// handled below
	default:
		printUsage()
		return fmt.Errorf("unknown command %q: expected 'export', 'import', or 'version'", action)
	}

	args, connStr, err := parseFlags()
	if err != nil {
		return err
	}

	setupLogging(args.verbose)

	// Root context cancelled on SIGINT/SIGTERM so in-flight COPY operations
	// can be aborted cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.CreatePool(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()

	switch action {
	case "export":
		log.Infof("Exporting tables from schema %s to directory %s using format %s", args.schema, args.dir, args.format)
		return postgres.Export(ctx, args.dir, args.format, pool, args.schema, args.table, args.columns, args.threads)
	case "import":
		log.Infof("Importing files from %s into schema %s", args.dir, args.schema)
		return postgres.Import(ctx, args.dir, pool, args.schema, args.table, args.columns, args.threads)
	}

	return nil
}

func setupLogging(verbose bool) {
	level := log.InfoLevel
	if verbose {
		level = log.DebugLevel
	}
	log.SetLevel(level)
	log.SetReportCaller(true)
	log.SetFormatter(&CustomFormatter{})
}

func parseFlags() (commandArgs, string, error) {
	const (
		defaultUser     = "postgres"
		defaultHost     = "localhost"
		defaultDatabase = "postgres"
		defaultSSLMode  = "disable"
		defaultDir      = "."
		defaultFormat   = "binary"
		defaultPort     = 5432
		defaultSchema   = "public"
		defaultThread   = 1
	)

	fs := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	fs.Usage = printUsage

	user := fs.String("user", envOr("PGUSER", defaultUser), "Username for database connection")
	password := fs.String("password", "", "Database password (falls back to PGPASSWORD)")
	host := fs.String("host", envOr("PGHOST", defaultHost), "Host of the database")
	database := fs.String("database", envOr("PGDATABASE", defaultDatabase), "Database name")
	sslmode := fs.String("sslmode", envOr("PGSSLMODE", defaultSSLMode), "Postgres SSL mode")
	dir := fs.String("dir", defaultDir, "Directory to process")
	table := fs.String("table", "", "Table to load or export")
	schemaName := fs.String("schema", defaultSchema, "Schema of table(s) to load or export")
	format := fs.String("format", defaultFormat, "Format of dump file (binary or csv)")
	columns := fs.String("columns", "all", "Comma-separated list of columns to export")
	port := fs.Int("port", defaultPort, "Port number for database connection")
	threadLimit := fs.Int("parallel", defaultThread, "Number of concurrent threads to perform exports/imports")
	verbose := fs.Bool("verbose", false, "Enable debug logging")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return commandArgs{}, "", err
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
		verbose:  *verbose,
	}

	if args.password == "" {
		args.password = os.Getenv("PGPASSWORD")
	}

	if args.format != "binary" && args.format != "csv" {
		return commandArgs{}, "", fmt.Errorf("invalid format %q: expected 'binary' or 'csv'", args.format)
	}

	if args.threads < 1 {
		return commandArgs{}, "", fmt.Errorf("--parallel must be at least 1")
	}

	if args.columns != "all" {
		args.columns = fmt.Sprintf("(%s)", args.columns)
	} else {
		args.columns = ""
	}

	connStr := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=%s user=%s password=%s",
		args.host, args.port, args.database, args.sslmode, args.user, args.password)

	return args, connStr, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func printUsage() {
	fmt.Fprint(os.Stderr, `pgPump - PostgreSQL data export and import

Usage:
  pgpump export|import [flags]
  pgpump version
  pgpump help

Flags:
  --user      Postgres user (default "postgres", env PGUSER)
  --password  Postgres password (env PGPASSWORD)
  --host      Postgres host (default "localhost", env PGHOST)
  --port      Postgres port (default 5432)
  --database  Database name (default "postgres", env PGDATABASE)
  --schema    Schema of table(s) (default "public")
  --table     Restrict to a single table
  --columns   Comma-separated column list, used with --table (default "all")
  --dir       Directory for dump files (default ".")
  --format    Dump format: binary or csv (default "binary")
  --parallel  Number of concurrent workers (default 1)
  --sslmode   Postgres SSL mode (default "disable", env PGSSLMODE)
  --verbose   Enable debug logging
`)
}
