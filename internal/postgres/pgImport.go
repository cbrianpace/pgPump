package pgPackage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	DataFile "pgPump/internal/datafile"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

func Import (dir string, pool *pgxpool.Pool, schema string, table string, columnList string, threads int) {
	start := time.Now()

	fileFilter := "DATA_"

	if table != "" {
		fileFilter += table + "."
	}
	
	fileList := DataFile.GetFilesInDir(dir, fileFilter)

	log.Infof("Running with a maximum of %d threads to import %d files", threads, len(fileList))

	// Semaphore channel to limit concurrent goroutines
	semaphore := make(chan struct{}, threads)

	// Channel to collect results
	results := make(chan string, len(fileList))

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	i := 0
	failedCnt := 0
	successCnt := 0
	
    // Loop through the files in the directory
    for _, file := range fileList {
		var tableName string

		parts := strings.Split(file, "_")
		if len(parts) > 1 {
			tableName = strings.Join(parts[1:], "_")
			tableName = tableName[:len(tableName)-4]
		} else {
			message := fmt.Sprintf("Failed to parse file %s", file)
			log.Error(message)
			failedCnt++
			continue
		}
		
		extension := filepath.Ext(file)
		fileType := "binary"

		switch extension {
		case ".csv":
			fileType = "csv"
		case ".bin":
			fileType = "binary"
		}

		i++
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a slot in the semaphore

		go func(taskID int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot in the semaphore

			err := ImportFile(dir, file, pool, fileType, schema, tableName, columnList)
			if err != nil {
				failedCnt++
				results <- fmt.Sprintf("File %s failed: %v", file, err)
			} else {
				successCnt++
				results <- fmt.Sprintf("File %s succeeded", file)
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

	log.Infof("Imports complete (%d total, %d successful, %d failed) in %.3f seconds",len(fileList),successCnt,failedCnt, time.Since(start).Seconds())

}


func ImportFile (dir string, fileName string, pool *pgxpool.Pool, fileType string, schema string, tableName string, columnList string) error {
	start := time.Now()
	fullFileName := dir + "/" + fileName
	
	log.Infof("Importing %s table %s.%s",fullFileName,schema,tableName)

	// Open a file to read the data
	inputFile, err := os.Open(fullFileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v\n", err)
		return err
	}
	defer inputFile.Close()

	sqlQuery := fmt.Sprintf("COPY %s.%s %s  FROM STDIN WITH %s",schema,tableName,columnList,fileType)

	println(sqlQuery)

	// Execute the COPY FROM STDIN command
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to acquire connection: %v\n", err)
		return err
	}
	defer conn.Release()

	_, err = conn.Conn().PgConn().CopyFrom(context.Background(), inputFile, sqlQuery)
	if err != nil {
		log.Fatalf("Failed to execute COPY FROM STDIN: %v\n", err)
		return err
	}

	log.Infof("Imported table %s.%s from %s in %.3f seconds",schema,tableName, fileName, time.Since(start).Seconds())				

	return nil

}