package logparser

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
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

	logTime, err := TimestringToTime(result["log_time"])
	if err != nil {
		return e, err
	}
	e.LogTime = logTime

	e.ErrorSeverity = result["error_severity"]

	e.Message = result["message"]

	e.Lines = lines

	return e, nil
}

// Convert a time string like "2025-04-28 00:20:02.274 EEST" to a time.Time object
func TimestringToTime(s string) (time.Time, error) {
	layout := "2006-01-02 15:04:05.000 MST"

	t, err := time.Parse(layout, s)
	if err != nil {
		layout = "2006-01-02 15:04:05 MST" // Try without milliseconds (RDS)
		t, err = time.Parse(layout, s)
		if err != nil {
			log.Error().Msgf("Failed to parse time string '%s' with layout: %s", s, layout)
		}
	}
	return t, err
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
func GetRecordsFromLogFile(filePath string, logLineParsingRegex *regexp.Regexp) <-chan pglog.CsvEntry {
	ch := make(chan pglog.CsvEntry)
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
		parseErrors := 0
		for scanner.Scan() {
			line := scanner.Text()

			// If the line does not have a timestamp, it is a continuation of the previous entry
			var e pglog.LogEntry
			if HasTimestampPrefix(line) {
				if gathering && len(lines) > 0 {
					e, err = EventLinesToPgLogEntry(lines, logLineParsingRegex)
					if err != nil {
						parseErrors += 1
						if parseErrors > 10 {
							log.Fatal().Err(err).Msg("10 parse errors reached, bailing")
						} else {
							log.Error().Err(err).Msgf("Regex parse failure for line: %s", line)
						}
						lines = make([]string, 0)
						continue
					}

				}
				ch <- pglog.CsvEntry{
					LogTime:       e.LogTime.String(),
					ErrorSeverity: e.ErrorSeverity,
					Message:       e.Message,
					Lines:         lines,
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
func ShowErrors(filePath string, minLvl string, extraFilters []string, fromTime time.Time, toTime time.Time, logLineParsingRegex *regexp.Regexp, minSlowDurationMs int) error {
	// Open file from filePath and loop line by line
	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Msgf("Error opening file %s", filePath)
		return err
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			log.Error().Err(err).Msgf("Error creating gzip reader for file %s", filePath)
			return err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var lines = make([]string, 0)

	gathering := false
	parseErrors := 0
	for scanner.Scan() {
		line := scanner.Text()

		// If the line does not have a timestamp, it is a continuation of the previous entry
		if HasTimestampPrefix(line) {
			if gathering && len(lines) > 0 {
				eventLines := strings.Join(lines, "\n")
				userFiltersSatisfied := 0

				e, err := EventLinesToPgLogEntry(lines, logLineParsingRegex)
				if err != nil {
					parseErrors += 1
					if parseErrors > 10 {
						log.Fatal().Err(err).Msg("10 parse errors reached, bailing")
					} else {
						log.Error().Err(err).Msgf("Default regex failure for line: %s", line)
					}
					continue
				}

				// log.Debug().Msgf("Processing entry with severity %s: %v", e.ErrorSeverity, strings.Join(lines, " "))

				if len(extraFilters) > 0 {
					for _, userFilter := range extraFilters {
						m, err := regexp.MatchString(userFilter, eventLines) // compile and cache the regex
						if err != nil {
							log.Fatal().Err(err).Msgf("Error matching user provided filter %s on line: %s", userFilter, eventLines)
							continue
						}
						if m {
							userFiltersSatisfied += 1
							break
						}
					}
				}

				if !TimestampFitsFromTo(e.LogTime, fromTime, toTime) {
					// log.Debug().Msgf("Skipping entry outside of time range: %s", e.LogTime)
					goto next_entry
				}

				if minSlowDurationMs > 0 {
					duration := ExtractDurationMillisFromLogMessage(e.Message)
					log.Debug().Msgf("Extracted duration: %f, message: %s", duration, e.Message)
					if duration < float64(minSlowDurationMs) {
						goto next_entry
					}
					fmt.Println(e.Lines)
				} else if e.SeverityNum() >= pglog.SeverityToNum(minLvl) && userFiltersSatisfied == len(extraFilters) {
					fmt.Println(e.Lines)
				}
			}
		next_entry:
			gathering = true
			lines = make([]string, 0)
		} else if !gathering { // Skip over very first non-full lines (is even possible?)
			continue
		}
		lines = append(lines, line)
	}

	return nil
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

func GetRecordsFromFileGeneric(filePath string, regex string) <-chan pglog.CsvEntry {
	ch := make(chan pglog.CsvEntry)
	go func() {
		defer close(ch)
		if strings.Contains(filePath, ".csv") {
			for rec := range GetRecordsFromCsvFile(filePath) {
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
			for rec := range GetRecordsFromLogFile(filePath, r) {
				ch <- rec
			}
		}
	}()
	return ch
}
