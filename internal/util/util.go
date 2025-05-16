package util

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/errors"
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

func HumanTimeOrDeltaStringToTime(hti string) (time.Time, error) {
	// Try parsing simple deltas first
	dur, err := time.ParseDuration(hti)
	if err == nil {
		return time.Now().Add(dur), nil
	}

	// Fallback to parsing short dates and full timestamps
	// Add support for parsing date-only timestamps in the format "2006-01-02"
	timestampFormats := []string{
		"2006-01-02 15:04:05.000 MST",
		"2006-01-02 15:04:05 MST",
	}
	for _, format := range timestampFormats {
		if t, err := time.Parse(format, hti); err == nil {
			return t, nil
		}
	}

	// Handle the "2006-01-02" format separately to use the local time zone
	if len(hti) == len("2006-01-02") {
		current_time := time.Now()
		currentTimeZone, _ := current_time.Zone()
		if t, err := time.Parse("2006-01-02 MST", hti+" "+currentTimeZone); err == nil {
			return t.In(time.Now().Location()), nil
		}
	}

	return time.Time{}, errors.New("unsupported time delta / timestamp format")
}
