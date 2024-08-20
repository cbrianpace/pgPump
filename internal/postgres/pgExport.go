package pgPackage

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

func Export(dir string, dumpFormat string, pool *pgxpool.Pool, schema string, table string, columnList string, threads int) {
	start := time.Now()

	// Get Tables to Export
	tables := GetTablesForSchema(pool, schema, table)

	log.Infof("Running with a maximum of %d threads to export %d tables", threads, len(tables))

	// Semaphore channel to limit concurrent goroutines
	semaphore := make(chan struct{}, threads)

	// Channel to collect results
	results := make(chan string, len(tables))

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	i := 0
	failedCnt := 0
	successCnt := 0

	for _, tableName := range tables {
		i++
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a slot in the semaphore

		go func(taskID int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot in the semaphore

			err := ExportTable(dir, dumpFormat, pool, schema, tableName, columnList)
			if err != nil {
				failedCnt++
				results <- fmt.Sprintf("Table %s failed: %v", tableName, err)
			} else {
				successCnt++
				results <- fmt.Sprintf("Table %s succeeded", tableName)
			}
		}(i)
	}

	// Close results channel once all tasks are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and print results
	for result := range results {
		log.Info(result)
	}

	log.Infof("Exports complete (%d total, %d successful, %d failed) in %.3f seconds",len(tables),successCnt,failedCnt, time.Since(start).Seconds())
							
}

func ExportTable(dir string, dumpFormat string, pool *pgxpool.Pool, schema string, table string, columnList string) error {

	start := time.Now()
	
	log.Infof("Exporting table %s.%s",schema,table)

	// Open a file to write the data
	fileName := fmt.Sprintf("%s/DATA_%s.%s",dir,table,(dumpFormat)[:3])
	outputFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v\n", err)
		return err
	}
	defer outputFile.Close()

	sqlQuery := fmt.Sprintf("COPY %s.%s %s TO stdout WITH %s",schema,table,columnList,dumpFormat)

	// Execute the COPY TO STDOUT command
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to acquire connection: %v\n", err)
		return err
	}
	defer conn.Release()

	_, err = conn.Conn().PgConn().CopyTo(context.Background(), outputFile, sqlQuery)
	if err != nil {
		log.Fatalf("Failed to execute COPY TO STDOUT: %v\n", err)
		return err
	}

	log.Infof("Exported table %s.%s to %s in %.3f seconds",schema,table, fileName, time.Since(start).Seconds())				

	return nil
}