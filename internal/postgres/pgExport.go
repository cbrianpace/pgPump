package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// Export writes one dump file per table in the schema to dir, running up to
// `threads` exports concurrently. It returns an error only if the set of tables
// cannot be determined or if any individual table export fails.
func Export(ctx context.Context, dir string, format string, pool *pgxpool.Pool, schema string, table string, columnList string, threads int) error {
	start := time.Now()

	tables, err := GetTablesForSchema(ctx, pool, schema, table)
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		return fmt.Errorf("no tables found in schema %q", schema)
	}

	log.Infof("Running with a maximum of %d threads to export %d tables", threads, len(tables))

	success, failed := runParallel(ctx, tables, threads, func(name string) task {
		return func(ctx context.Context) error {
			return ExportTable(ctx, dir, format, pool, schema, name, columnList)
		}
	})

	log.Infof("Exports complete (%d total, %d successful, %d failed) in %.3f seconds",
		len(tables), success, failed, time.Since(start).Seconds())

	if failed > 0 {
		return fmt.Errorf("%d of %d table exports failed", failed, len(tables))
	}
	return nil
}

// ExportTable copies a single table to a dump file named DATA_<table>.<ext>.
func ExportTable(ctx context.Context, dir string, format string, pool *pgxpool.Pool, schema string, table string, columnList string) error {
	start := time.Now()
	log.Infof("Exporting table %s.%s", schema, table)

	fileName := filepath.Join(dir, fmt.Sprintf("DATA_%s.%s", table, format[:3]))
	outputFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", fileName, err)
	}
	defer outputFile.Close()

	relation := pgx.Identifier{schema, table}.Sanitize()
	sqlQuery := fmt.Sprintf("COPY %s %s TO STDOUT WITH (FORMAT %s)", relation, columnList, format)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Conn().PgConn().CopyTo(ctx, outputFile, sqlQuery); err != nil {
		return fmt.Errorf("COPY TO STDOUT for %s: %w", relation, err)
	}

	log.Infof("Exported table %s.%s to %s in %.3f seconds", schema, table, fileName, time.Since(start).Seconds())
	return nil
}
