package util

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
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
	var fileData []struct {
		info os.FileInfo
		path string
	}

	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		log.Debug().Msgf("Found file: %s", path)
		if !info.IsDir() {
			fileData = append(fileData, struct {
				info os.FileInfo
				path string
			}{info: info, path: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(fileData, func(i, j int) bool {
		return fileData[i].info.ModTime().After(fileData[j].info.ModTime())
	})

	var logFiles []string
	for _, data := range fileData {
		logFiles = append(logFiles, data.path)
	}

	return logFiles, nil
}

func HumanTimedeltaToTime(htd string) (time.Time, error) {
	dur, err := time.ParseDuration(htd)
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().Add(dur), nil
}
