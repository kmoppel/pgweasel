package util

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/detector"
	dps "github.com/markusmobius/go-dateparser"
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

func IsPathExistsAndFolder(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
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

	// Special case for "today"
	if strings.ToLower(humanInput) == "today" {
		year, month, day := referenceTime.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, referenceTime.Location()), nil
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

// IntervalToMillis converts an interval string to milliseconds.
// If no unit is specified, assumes the value is in milliseconds.
func IntervalToMillis(interval string) (int, error) {
	// Remove spaces to handle formats like "5 min"
	interval = strings.TrimSpace(strings.ReplaceAll(interval, " ", ""))

	// Check if it's just a number (no units)
	if val, err := strconv.Atoi(interval); err == nil {
		return val, nil // Return as is, assuming milliseconds
	}

	// Try standard ParseDuration
	dur, err := time.ParseDuration(interval)
	if err != nil {
		// Try to handle common aliases like "min" for "m"
		replacer := strings.NewReplacer(
			"mins", "m",
			"min", "m",
			"minutes", "m",
			"minute", "m",
			"secs", "s",
			"sec", "s",
			"seconds", "s",
			"second", "s",
			"hrs", "h",
			"hr", "h",
			"hours", "h",
			"hour", "h",
			"days", "24h",
			"day", "24h",
		)
		normalized := replacer.Replace(interval)
		dur, err = time.ParseDuration(normalized)
		if err != nil {
			return 0, errors.Wrap(err, "invalid interval format")
		}
	}
	return int(dur.Milliseconds()), nil
}

// Convert a time string like "2025-04-28 00:20:02.274 EEST" to a time.Time object
func TimestringToTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	// Check if it's an epoch timestamp (log_line_prefix %n) a la 1748867052.006
	if epoch, err := strconv.ParseFloat(s, 64); err == nil {
		sec := int64(epoch)
		nsec := int64((epoch - float64(sec)) * 1e9)
		return time.Unix(sec, nsec)
	}

	layout := "2006-01-02 15:04:05.000 MST"

	t, err := time.Parse(layout, s)
	if err != nil {
		layout = "2006-01-02 15:04:05 MST" // Try without milliseconds (RDS)
		t, err = time.Parse(layout, s)
		if err != nil {
			log.Error().Msgf("Failed to parse time string '%s' with layout: %s", s, layout)
			return time.Time{}
		}
	}
	return t
}

func CheckStdinAvailable() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Error().Err(err).Msg("Error Stat-ing stdin")
		return false
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		log.Debug().Msg("Stdin is not available")
		return false
	}
	return true
}

func GetLogFilesFromUserArgs(args []string) []string {
	var logFiles []string

	if len(args) == 0 {
		log.Debug().Msgf("No files / folders provided, looking for latest file from default locations: %v ...", detector.DEFAULT_LOG_LOCATIONS)
		logFile, _, err := detector.DiscoverLatestLogFileAndFolder(nil)
		if err != nil {
			log.Warn().Msgf("Failed to detect any log files from default locations: %v", detector.DEFAULT_LOG_LOCATIONS)
			return nil
		}
		logFiles = append(logFiles, logFile)
	} else {
		for _, arg := range args {
			log.Debug().Msgf("Checking input path: %s ...", arg)
			_, err := os.Stat(arg)
			if err != nil {
				log.Warn().Err(err).Msgf("Error accessing path: %s", arg)
				continue
			}

			if IsPathExistsAndFile(arg) {
				logFiles = append(logFiles, arg)
				continue
			}
			if IsPathExistsAndFolder(arg) {
				log.Debug().Msgf("Looking for log files from folder: %s ..", arg)
				logFiles, err = GetPostgresLogFilesTimeSorted(arg)
				if err != nil {
					log.Warn().Err(err).Msgf("Error scanning for log files from folder: %s", arg)
					continue
				}
				log.Debug().Msgf("Found %d log files", len(logFiles))
				logFiles = append(logFiles, logFiles...)
			}
		}
	}
	return logFiles
}

func NormalizeErrorMessage(msg string) string { // TODO
	// Remove leading and trailing whitespace
	msg = strings.TrimSpace(msg)

	// Replace multiple spaces with a single space
	msg = strings.Join(strings.Fields(msg), " ")

	// Convert to lowercase
	msg = strings.ToLower(msg)

	// Remove special characters (except for alphanumeric and spaces)
	msg = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == ' ' {
			return r
		}
		return -1
	}, msg)

	return msg

}

func TruncateString(s string, maxChars int) string {
	runes := []rune(s)
	if len(runes) > maxChars {
		return string(runes[:maxChars])
	}
	return s
}
