package util

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dps "github.com/markusmobius/go-dateparser"
	"github.com/pkg/errors"
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
		// log.Debug().Msgf("Found file: %s", path)
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

func HumanTimeOrDeltaStringToTime(humanInput string, referenceTime time.Time) (time.Time, error) {
	if humanInput == "" {
		return time.Time{}, nil
	}

	if referenceTime.IsZero() {
		referenceTime = time.Now()
	}

	// Try parsing simple deltas first as probably the most common case
	// ParseDuration valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	dur, err := time.ParseDuration(humanInput)
	if err == nil {
		if strings.HasPrefix(humanInput, "-") {
			return referenceTime.Add(dur), nil
		}
		// Add leading '-' for positive durations, parse again
		dur, err = time.ParseDuration("-" + humanInput)
		if err == nil {
			return referenceTime.Add(dur), nil
		}
	}

	// Fallback to parsing short dates and full timestamps
	// Add support for parsing date-only timestamps in the format "2006-01-02"
	timestampFormats := []string{
		"2006-01-02 15:04:05.000 MST",
		"2006-01-02 15:04:05 MST",
	}
	for _, format := range timestampFormats {
		if t, err := time.Parse(format, humanInput); err == nil {
			return t, nil
		}
	}

	// Handle the "2006-01-02" format separately to use the local time zone
	if len(humanInput) == len("2006-01-02") {
		currentTimeZone, _ := referenceTime.Zone()
		if t, err := time.Parse("2006-01-02 MST", humanInput+" "+currentTimeZone); err == nil {
			return t.In(referenceTime.Location()), nil
		}
	}

	// Try a human-friendly date parser library for "1 hour ago" support plus non-standard dates
	// https://github.com/markusmobius/go-dateparser?tab=readme-ov-file#62-relative-date
	cfg := &dps.Configuration{
		CurrentTime:     time.Now(),
		DefaultTimezone: time.Local,
	}

	dt, err := dps.Parse(cfg, humanInput)
	if err == nil {
		return dt.Time, nil
	}

	return time.Time{}, errors.New("unsupported time delta / timestamp format")
}

func IntervalToMillis(interval string) (int, error) {
	return 0, nil
}
