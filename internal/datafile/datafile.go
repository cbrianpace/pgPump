package DataFile

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	csvExtension = ".csv"
	binExtension = ".bin"
)

func GetFilesInDir (dir string, fileFilter string) []string {
	var fileList []string

	// Open the directory
    files, err := os.ReadDir(dir)
    if err != nil {
        log.Fatalf("Failed to read directory: %v", err)
    }

	for _, file := range files {

		if file.IsDir() {
            continue // Skip directories
        }

		if strings.HasPrefix(file.Name(), fileFilter) {
			// Get the file extension
			extension := filepath.Ext(file.Name())

			if strings.Contains(extension, csvExtension) || strings.Contains(extension, binExtension) {
				fileList = append(fileList, file.Name())
			} else {
				log.Warnf("Invalid file format %s, skipping files",extension)
				continue
			}
		}
	}

	return fileList

}