package datafile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	csvExtension = ".csv"
	binExtension = ".bin"
)

// GetFilesInDir returns the names of dump files in dir whose names start with
// fileFilter and carry a recognized (.csv or .bin) extension.
func GetFilesInDir(dir string, fileFilter string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var fileList []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, fileFilter) {
			continue
		}

		switch filepath.Ext(name) {
		case csvExtension, binExtension:
			fileList = append(fileList, name)
		default:
			log.Warnf("Skipping file with unsupported extension: %s", name)
		}
	}

	return fileList, nil
}
