package logparser

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/rs/zerolog/log"
)

const DEFAULT_REGEX_STR = `^(?P<log_time>[\d\-:\. ]+ [A-Z]+).*[\s:]+(?P<error_severity>[A-Z0-9]+):\s*(?P<message>(?s:.*))$`

var DEFAULT_REGEX = regexp.MustCompile(DEFAULT_REGEX_STR)
var REGEX_DURATION_MILLIS = regexp.MustCompile(`duration:\s*([\d\.]+)\s*ms`)

func EventLinesToPgLogEntry(lines []string, r *regexp.Regexp) (pglog.LogEntry, error) {
	e := pglog.LogEntry{}
	if len(lines) == 0 {
		return e, errors.New("empty log line")
	}
	line := strings.Join(lines, "\n")
	match := r.FindStringSubmatch(line)
	if match == nil {
		return e, errors.New("failed to parse log line regex")
	}

	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i > 0 && name != "" {
			result[name] = match[i]
		}
	}

	e.LogTime = result["log_time"]

	e.ErrorSeverity = result["error_severity"]

	e.Message = result["message"]

	e.Lines = lines

	return e, nil
}

// Returns 0 if no match or error
func ExtractDurationMillisFromLogMessage(message string) float64 {
	// Example message: "duration: 0.211 ms"
	match := REGEX_DURATION_MILLIS.FindStringSubmatch(message)
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

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func GetLogRecordsFromLogFile(filePath string, logLineParsingRegex *regexp.Regexp) <-chan pglog.LogEntry {
	log.Debug().Msgf("Looking for log entries from plain text log file: %s", filePath)
	ch := make(chan pglog.LogEntry)
	go func() {
		defer close(ch)
		// Open file from filePath and loop line by line
		file, err := os.Open(filePath)
		if err != nil {
			log.Error().Err(err).Msgf("Error opening file: %s", filePath)
			return
		}
		defer file.Close()

		var reader io.Reader = file
		if strings.HasSuffix(filePath, ".gz") {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				log.Error().Err(err).Msgf("Error creating gzip reader for file: %s", filePath)
				return
			}
			defer gzReader.Close()
			reader = gzReader
		}

		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanLines)

		var lines = make([]string, 0)
		if logLineParsingRegex == nil {
			logLineParsingRegex = DEFAULT_REGEX
		}

		gathering := false
		for scanner.Scan() {
			line := scanner.Text()
			// log.Debug().Msgf("Got line: %s", line)

			// If the line does not have a timestamp, it is a continuation of the previous entry
			if HasTimestampPrefix(line) {
				if gathering && len(lines) > 0 {
					e, err := EventLinesToPgLogEntry(lines, logLineParsingRegex)
					if err != nil {
						log.Fatal().Err(err).Msgf("Log line regex parse error. Line: %s", strings.Join(lines, "\n"))
					}
					lines = make([]string, 0)
					ch <- e
					continue
				}
				gathering = true
				lines = make([]string, 0)
			} else if !gathering { // Skip over very first non-full lines (is even possible?)
				continue
			}
			lines = append(lines, line)
		}
	}()
	return ch
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func DoesLogRecordSatisfyUserFilters(rec pglog.LogEntry, minLvl string, extraRegexFilters []string, fromTime time.Time, toTime time.Time, logLineParsingRegex *regexp.Regexp, minSlowDurationMs int) bool {
	if rec.SeverityNum() < pglog.SeverityToNum(minLvl) {
		return false
	}

	if len(extraRegexFilters) > 0 {
		eventLines := strings.Join(rec.Lines, "\n")
		for _, userFilter := range extraRegexFilters {
			m, err := regexp.MatchString(userFilter, eventLines) // TODO compile and cache the regex
			if err != nil {
				log.Fatal().Err(err).Msgf("Error matching user provided filter %s on line: %s", userFilter, eventLines) // Fail early for beta period at least
			}
			if !m {
				return false
			}
		}
	}

	if !TimestampFitsFromTo(rec.GetTime(), fromTime, toTime) {
		// log.Debug().Msgf("Skipping entry outside of time range: %s", e.LogTime)
		return false
	}

	if minSlowDurationMs > 0 {
		duration := ExtractDurationMillisFromLogMessage(rec.Message)
		log.Debug().Msgf("Extracted duration: %f, message: %s", duration, rec.Message)
		if duration < float64(minSlowDurationMs) {
			return false
		}

	}

	return true
}

func TimestampFitsFromTo(time, fromTime, toTime time.Time) bool {
	if !fromTime.IsZero() && time.Before(fromTime) {
		return false
	}
	if !toTime.IsZero() && time.After(toTime) {
		return false
	}
	return true
}

func HasTimestampPrefix(line string) bool {
	r := regexp.MustCompile(`^(?P<time>[\d\-:\. ]+ [A-Z]+)`)
	return r.MatchString(line)
}

func GetLogRecordsFromFile(filePath string, regex string) <-chan pglog.LogEntry {
	ch := make(chan pglog.LogEntry)
	go func() {
		defer close(ch)
		if strings.Contains(filePath, ".csv") {
			for rec := range GetLogRecordsFromCsvFile(filePath) {
				ch <- rec
			}
		} else {
			var r *regexp.Regexp
			var err error

			if regex != "" {
				r, err = regexp.Compile(regex)
				if err != nil {
					log.Fatal().Err(err).Msgf("Error compiling regex: %s", regex)
				}
			}
			for rec := range GetLogRecordsFromLogFile(filePath, r) {
				ch <- rec
			}
		}
	}()
	return ch
}
