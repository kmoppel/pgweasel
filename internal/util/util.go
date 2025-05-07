package util

import (
	"os"
	"path/filepath"
	"sort"
)

func IsPathExistsAndFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if !fileInfo.IsDir() {
		return true
	}
	return false
}

// Returns all text files recursively in the given folder, sorted by modification time
func GetPostgresLogFilesTimeSorted(filePath string) ([]string, error) {
	var files []os.FileInfo
	var logFiles []string

	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, info)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, file := range files {
		logFiles = append(logFiles, filepath.Join(filePath, file.Name()))
	}

	return logFiles, nil
}
