package logparser

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog/log"
)

const DEFAULT_REGEX_STR = `(?s)^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \[[0-9\-]+\] )?(?P<log_time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})[\s:\-].*?[\s:\-]?(?P<error_severity>[A-Z12345]{3,12}):\s*(?P<message>(?s:.*))$`

var DEFAULT_REGEX = regexp.MustCompile(DEFAULT_REGEX_STR)
var REGEX_HAS_TIMESTAMP_PREFIX = regexp.MustCompile(`^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \[[0-9\-]+\] )?(?P<time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})`)
var REGEX_LOG_LEVEL = regexp.MustCompile(`^[A-Z12345]{3,12}$`)

func EventLinesToPgLogEntry(lines []string, r *regexp.Regexp, filename string) (pglog.LogEntry, error) {
	e := pglog.LogEntry{}
	if len(lines) == 0 {
		return e, errors.New("empty log line")
	}
	line := strings.Join(lines, "\n")
	match := r.FindStringSubmatch(line)
	if match == nil {
		return e, errors.New("failed to parse log line regex for file: " + filename)
	}
	// log.Debug().Msgf("Regex match: %+v", match)

	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i > 0 && name != "" {
			result[name] = match[i]
		}
	}

	e.LogTime = result["log_time"]

	e.ErrorSeverity = result["error_severity"]

	if !REGEX_LOG_LEVEL.MatchString(e.ErrorSeverity) {
		return e, errors.New("invalid log level: " + e.ErrorSeverity)
	}

	e.Message = result["message"]

	e.Lines = lines

	return e, nil
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func GetLogRecordsFromLogFile(filePath string, logLineParsingRegex *regexp.Regexp) <-chan pglog.LogEntry {
	log.Debug().Msgf("Looking for log entries from plain text log file: %s", filePath)
	ch := make(chan pglog.LogEntry)
	var scanner *bufio.Scanner

	go func() {
		defer close(ch)

		if filePath == "stdin" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {
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
			scanner = bufio.NewScanner(reader)
		}

		scanner.Split(bufio.ScanLines)

		var lines = make([]string, 0)
		if logLineParsingRegex == nil {
			logLineParsingRegex = DEFAULT_REGEX
		}

		firstCompleteEntryFound := false
		for scanner.Scan() {
			line := scanner.Text()
			log.Debug().Msgf("Inspecting line: %s", line)

			// If the line does not have a timestamp, it is a continuation of the previous entry
			if HasTimestampPrefix(line) {
				if firstCompleteEntryFound {
					e, err := EventLinesToPgLogEntry(lines, logLineParsingRegex, filePath)
					if err != nil {
						log.Fatal().Err(err).Msgf("Log line regex parse error. Line: %s", strings.Join(lines, "\n"))
					}
					log.Debug().Msgf("Capture OK. severity:%s lines: %d len(lines): %d", e.ErrorSeverity, len(lines), len(strings.Join(lines, " ")))
					lines = make([]string, 0)
					lines = append(lines, line)
					ch <- e
					continue
				}
				firstCompleteEntryFound = true
			}
			lines = append(lines, line)
		}
		// Special handling for the last line
		if firstCompleteEntryFound && len(lines) > 0 {
			e, err := EventLinesToPgLogEntry(lines, logLineParsingRegex, filePath)
			if err != nil {
				log.Fatal().Err(err).Msgf("Log line regex parse error. Line: %s", strings.Join(lines, "\n"))
			}
			ch <- e
		}
	}()
	return ch
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func DoesLogRecordSatisfyUserFilters(rec pglog.LogEntry, minLvlNum int, extraRegexFilters []string, fromTime time.Time, toTime time.Time, minSlowDurationMs int, systemOnly bool) bool {
	if systemOnly {
		return rec.IsSystemEntry()
	}

	if rec.SeverityNum() < minLvlNum {
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
		duration := util.ExtractDurationMillisFromLogMessage(rec.Message)
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
	return REGEX_HAS_TIMESTAMP_PREFIX.MatchString(line)
}

func GetLogRecordsFromFile(filePath string, r *regexp.Regexp, useCsvFormat bool) <-chan pglog.LogEntry {
	ch := make(chan pglog.LogEntry)
	go func() {
		defer close(ch)
		if strings.Contains(filePath, ".csv") || useCsvFormat {
			for rec := range GetLogRecordsFromCsvFile(filePath) {
				ch <- rec
			}
		} else {
			for rec := range GetLogRecordsFromLogFile(filePath, r) {
				ch <- rec
			}
		}
	}()
	return ch
}
