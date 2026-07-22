package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	datafile "pgPump/internal/datafile"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const dumpFilePrefix = "DATA_"

// Import loads every dump file in dir into the target schema, running up to
// `threads` imports concurrently. When table is non-empty, only that table's
// dump file is imported.
func Import(ctx context.Context, dir string, pool *pgxpool.Pool, schema string, table string, columnList string, threads int) error {
	start := time.Now()

	fileFilter := dumpFilePrefix
	if table != "" {
		fileFilter += table + "."
	}

	fileList, err := datafile.GetFilesInDir(dir, fileFilter)
	if err != nil {
		return err
	}
	if len(fileList) == 0 {
		return fmt.Errorf("no dump files matching %q found in %s", fileFilter, dir)
	}

	log.Infof("Running with a maximum of %d threads to import %d files", threads, len(fileList))

	success, failed := runParallel(ctx, fileList, threads, func(file string) task {
		return func(ctx context.Context) error {
			tableName, fileType, err := parseDumpFileName(file)
			if err != nil {
				return err
			}
			return ImportFile(ctx, dir, file, pool, fileType, schema, tableName, columnList)
		}
	})

	log.Infof("Imports complete (%d total, %d successful, %d failed) in %.3f seconds",
		len(fileList), success, failed, time.Since(start).Seconds())

	if failed > 0 {
		return fmt.Errorf("%d of %d file imports failed", failed, len(fileList))
	}
	return nil
}

// parseDumpFileName derives the table name and COPY format from a dump file
// name of the form DATA_<table>.<bin|csv>. The table name may itself contain
// underscores.
func parseDumpFileName(file string) (tableName string, fileType string, err error) {
	base := filepath.Base(file)
	ext := filepath.Ext(base)

	switch ext {
	case ".csv":
		fileType = "csv"
	case ".bin":
		fileType = "binary"
	default:
		return "", "", fmt.Errorf("unrecognized dump file extension %q for %s", ext, file)
	}

	name := strings.TrimSuffix(base, ext)
	if !strings.HasPrefix(name, dumpFilePrefix) {
		return "", "", fmt.Errorf("file %s missing %q prefix", file, dumpFilePrefix)
	}
	tableName = strings.TrimPrefix(name, dumpFilePrefix)
	if tableName == "" {
		return "", "", fmt.Errorf("could not parse table name from file %s", file)
	}

	return tableName, fileType, nil
}

// ImportFile loads a single dump file into schema.table via COPY FROM STDIN.
func ImportFile(ctx context.Context, dir string, fileName string, pool *pgxpool.Pool, fileType string, schema string, tableName string, columnList string) error {
	start := time.Now()
	fullFileName := filepath.Join(dir, fileName)

	log.Infof("Importing %s into table %s.%s", fullFileName, schema, tableName)

	inputFile, err := os.Open(fullFileName)
	if err != nil {
		return fmt.Errorf("opening file %s: %w", fullFileName, err)
	}
	defer inputFile.Close()

	relation := pgx.Identifier{schema, tableName}.Sanitize()
	sqlQuery := fmt.Sprintf("COPY %s %s FROM STDIN WITH (FORMAT %s)", relation, columnList, fileType)
	log.Debugf("executing: %s", sqlQuery)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Conn().PgConn().CopyFrom(ctx, inputFile, sqlQuery); err != nil {
		return fmt.Errorf("COPY FROM STDIN for %s: %w", relation, err)
	}

	log.Infof("Imported table %s.%s from %s in %.3f seconds", schema, tableName, fileName, time.Since(start).Seconds())
	return nil
}
