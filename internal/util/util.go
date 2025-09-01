package util

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/detector"
	dps "github.com/markusmobius/go-dateparser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var REGEX_DURATION_MILLIS = regexp.MustCompile(`duration:\s*([\d\.]+)\s*ms`)
var REGEX_CHECKPOINT_DURATION_SECONDS = regexp.MustCompile(`total=([\d\.]+) s;`)
var REGEX_AUTOVACUUM_DURATION_SECONDS = regexp.MustCompile(`(?s)^automatic (analyze|vacuum) of table "(?P<table_name>[\w\.\-]+)".* elapsed: (?P<duration>[\d\.]+) s$`)
var REGEX_AUTOVACUUM_READ_WRITE_RATES = regexp.MustCompile(`avg read rate: ([\d\.]+) MB/s, avg write rate: ([\d\.]+) MB/s`)

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

	// Convert day unit input to hours as ParseDuration doesn't support days
	dayRegex := regexp.MustCompile(`^(-?\d+)(d|day|days)$`)
	if matches := dayRegex.FindStringSubmatch(humanInput); matches != nil {
		daysNum, err := strconv.Atoi(matches[1])
		if err == nil {
			humanInput = strconv.Itoa(daysNum*24) + "h" // Convert days to hours
		}
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

// Returns (0, "") if no match or error
func ExtractDurationMillisFromLogMessage(message string) (float64, string) {
	// Example message: "duration: 0.211 ms"
	match := REGEX_DURATION_MILLIS.FindStringSubmatch(message)
	if match == nil {
		return 0.0, ""
	}

	durationStr := match[1]
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0.0, ""
	}
	return duration, durationStr
}

// Returns 0 if no match or error
func ExtractCheckpointDurationSecondsFromLogMessage(message string) float64 {
	// checkpoint complete: wrote 66 buffers (0.4%); 0 WAL file(s) added, 0 removed, 0 recycled; write=6.468 s, sync=0.036 s, total=6.517 s; sync files=48, longest=0.009 s, average=0.001 s; distance=152 kB, estimate=152 kB
	match := REGEX_CHECKPOINT_DURATION_SECONDS.FindStringSubmatch(message)
	if match == nil {
		return 0.0
	}

	durationStr := match[1]
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0.0
	}
	return duration
}

// Returns duration in seconds and table name if matched, otherwise 0 and empty string
func ExtractAutovacuumOrAnalyzeDurationSecondsFromLogMessage(message string) (float64, string) {
	// Example: automatic vacuum of table "mytable"... elapsed: 2326.38 s
	match := REGEX_AUTOVACUUM_DURATION_SECONDS.FindStringSubmatch(message)
	if match == nil {
		return 0.0, ""
	}

	// Get named capture groups
	tableName := ""
	durationStr := ""

	for i, name := range REGEX_AUTOVACUUM_DURATION_SECONDS.SubexpNames() {
		if i > 0 && i <= len(match) {
			if name == "table_name" {
				tableName = match[i]
			} else if name == "duration" {
				durationStr = match[i]
			}
		}
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0.0, ""
	}
	return duration, tableName
}

// Returns read rate and write rate in MB/s if matched, otherwise 0.0, 0.0
func ExtractAutovacuumReadWriteRatesFromLogMessage(message string) (float64, float64) {
	// Example: avg read rate: 5.492 MB/s, avg write rate: 4.932 MB/s
	match := REGEX_AUTOVACUUM_READ_WRITE_RATES.FindStringSubmatch(message)
	if match == nil {
		return 0.0, 0.0
	}

	readRate, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0.0, 0.0
	}

	writeRate, err := strconv.ParseFloat(match[2], 64)
	if err != nil {
		return 0.0, 0.0
	}

	return readRate, writeRate
}

func ExtractConnectHostFromLogMessage(message string) string {
	// Example: "connection received: host=127.0.0.1 port=44410"
	// Example: "connection received: host=[local]"
	host := ""
	if hostStart := strings.Index(message, "host="); hostStart != -1 {
		hostStart += 5 // Skip "host="
		hostEnd := strings.Index(message[hostStart:], " ")
		if hostEnd == -1 {
			host = message[hostStart:]
		} else {
			host = message[hostStart : hostStart+hostEnd]
		}

		// Handle [local] case
		if host == "[local]" {
			host = "local"
		}
	}
	return host
}

func ExtractConnectUserDbAppnameSslFromLogMessage(message string) (string, string, string, bool) {
	// connection authorized: user=krl database=postgres application_name=psql
	// connection authorized: user=monitor database=bench SSL enabled (protocol=TLSv1.3, cipher=TLS_AES_256_GCM_SHA384, bits=256)
	user := ""
	db := ""
	appname := ""
	ssl := false

	if userStart := strings.Index(message, "user="); userStart != -1 {
		userStart += 5 // Skip "user="
		userEnd := strings.Index(message[userStart:], " ")
		if userEnd == -1 {
			user = message[userStart:]
		} else {
			user = message[userStart : userStart+userEnd]
		}
	}

	if dbStart := strings.Index(message, "database="); dbStart != -1 {
		dbStart += 9 // Skip "database="
		dbEnd := strings.Index(message[dbStart:], " ")
		if dbEnd == -1 {
			db = message[dbStart:]
		} else {
			db = message[dbStart : dbStart+dbEnd]
		}
	}

	if appnameStart := strings.Index(message, "application_name="); appnameStart != -1 {
		appnameStart += 17 // Skip "application_name="
		appnameEnd := strings.Index(message[appnameStart:], "\n")

		if appnameEnd == -1 {
			appname = message[appnameStart:]
		} else {
			appname = message[appnameStart : appnameStart+appnameEnd]
		}
	}

	if strings.Contains(message, "SSL enabled") {
		ssl = true
	}

	return user, db, appname, ssl
}

// CalculatePercentile calculates the percentile value from a sorted slice of float64 values
func CalculatePercentile(sortedData []float64, percentile float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}
	if len(sortedData) == 1 {
		return sortedData[0]
	}

	n := float64(len(sortedData))
	rank := (percentile / 100.0) * (n - 1)
	lowerIndex := int(rank)
	upperIndex := lowerIndex + 1

	if upperIndex >= len(sortedData) {
		return sortedData[len(sortedData)-1]
	}

	weight := rank - float64(lowerIndex)
	return sortedData[lowerIndex]*(1-weight) + sortedData[upperIndex]*weight
}
