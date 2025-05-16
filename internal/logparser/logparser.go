package logparser

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/rs/zerolog/log"
)

var DEFAULT_REGEX = regexp.MustCompile(`^(?P<log_time>[\d\-:\. ]+ [A-Z]+).*(?P<error_severity>DEBUG5|DEBUG4|DEBUG3|DEBUG2|DEBUG1|LOG|INFO|NOTICE|WARNING|ERROR|FATAL|PANIC|STATEMENT|DETAIL):\s*(?P<message>(?s:.*))$`)

func EventLinesToPgLogEntry(line string, r *regexp.Regexp) (pglog.LogEntry, error) {
	e := pglog.LogEntry{}
	if line == "" {
		return e, errors.New("empty log line")
	}

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

	e.Line = line

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

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func ShowErrors(filePath string, minLvl string, extraFilters []string, fromTime time.Time, toTime time.Time) error {
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

				e, err := EventLinesToPgLogEntry(eventLines, DEFAULT_REGEX)
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

				if e.SeverityNum() >= pglog.SeverityToNum(minLvl) && userFiltersSatisfied == len(extraFilters) && TimestampFitsFromTo(e.LogTime, fromTime, toTime) {
					fmt.Println(e.Line)
				}
			}
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
