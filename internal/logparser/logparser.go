package logparser

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog/log"
)

const LOG_ENTRIES_PER_BATCH = 10

var REGEX_HAS_TIMESTAMP_PREFIX = regexp.MustCompile(`^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \[[0-9\-]+\] )?(?P<time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})`)
var REGEX_LOG_TIME = regexp.MustCompile(`^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \[[0-9\-]+\] )?(?P<log_time>[\d\-:\. ]{19,23} [A-Z0-9\-\+]{2,5}|[0-9\.]{14})[\s:\-]`)
var REGEX_LOG_LEVEL_MESSAGE = regexp.MustCompile(`^.*?[\s:\-](?P<log_level>[A-Z12345]{3,12}):  (?P<message>.*)$`)

func EventLinesToPgLogEntry(lines []string, filename string) (pglog.LogEntry, error) {
	e := pglog.LogEntry{}
	if len(lines) == 0 {
		return e, errors.New("empty log line")
	}

	match := REGEX_LOG_TIME.FindStringSubmatch(lines[0])
	if match == nil {
		return e, errors.New("failed to parse REGEX_LOG_TIME regex for file " + filename + " line: " + lines[0])
	}
	// log.Debug().Msgf("Regex match: %+v", match)
	result := make(map[string]string)
	for i, name := range REGEX_LOG_TIME.SubexpNames() {
		if i > 0 && name != "" {
			result[name] = match[i]
		}
	}
	e.LogTime = result["log_time"]

	match = REGEX_LOG_LEVEL_MESSAGE.FindStringSubmatch(lines[0])
	if match == nil {
		return e, errors.New("failed to parse REGEX_LOG_LEVEL_MESSAGE regex for file " + filename + " line: " + lines[0])
	}
	// log.Debug().Msgf("Regex match: %+v", match)
	result = make(map[string]string)
	for i, name := range REGEX_LOG_LEVEL_MESSAGE.SubexpNames() {
		if i > 0 && name != "" {
			result[name] = match[i]
		}
	}
	e.ErrorSeverity = result["log_level"]

	e.Message = result["message"]
	if len(lines) > 1 {
		e.Message += "\n" + strings.Join(lines[1:], "\n")
	}

	e.Lines = lines

	return e, nil
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func GetLogRecordsFromLogFile(filePath string, logLineParsingRegex *regexp.Regexp) <-chan []pglog.LogEntry {
	log.Debug().Msgf("Looking for log entries from plain text log file: %s", filePath)
	ch := make(chan []pglog.LogEntry)
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

		batch := make([]pglog.LogEntry, 0, LOG_ENTRIES_PER_BATCH)
		firstCompleteEntryFound := false
		for scanner.Scan() {
			line := scanner.Text()
			log.Debug().Msgf("Inspecting line: %s", line)

			// If the line does not have a timestamp, it is a continuation of the previous entry
			if HasTimestampPrefix(line) {
				if firstCompleteEntryFound {
					// e, err := EventLinesToPgLogEntry(lines, filePath)
					e, err := EventLinesToPgLogEntry(lines, filePath)
					if err != nil {
						log.Fatal().Err(err).Msgf("Log line regex parse error. Line: %s", strings.Join(lines, "\n"))
					}
					log.Debug().Msgf("Capture OK. severity:%s lines: %d len(lines): %d", e.ErrorSeverity, len(lines), len(strings.Join(lines, " ")))
					lines = make([]string, 0)
					lines = append(lines, line)

					batch = append(batch, e)

					// Send batch when it reaches newConst entries
					if len(batch) == LOG_ENTRIES_PER_BATCH {
						ch <- batch
						batch = make([]pglog.LogEntry, 0, LOG_ENTRIES_PER_BATCH)
					}
					continue
				}
				firstCompleteEntryFound = true
			}
			lines = append(lines, line)
		}
		// Special handling for the last line
		if firstCompleteEntryFound && len(lines) > 0 {
			// e, err := EventLinesToPgLogEntry(lines, filePath)
			e, err := EventLinesToPgLogEntry(lines, filePath)
			if err != nil {
				log.Fatal().Err(err).Msgf("Log line regex parse error. Line: %s", strings.Join(lines, "\n"))
			}
			batch = append(batch, e)
		}

		// Send remaining entries in batch if any
		if len(batch) > 0 {
			ch <- batch
		}
	}()
	return ch
}

func DoesLogRecordSatisfyUserFilters(rec pglog.LogEntry, minLvlNum int, extraRegexFilters []string, fromTime time.Time, toTime time.Time, minSlowDurationMs int, systemOnly bool, systemIncludeCheckpointer bool, grepRegex *regexp.Regexp) bool {
	if grepRegex != nil {
		return grepRegex.MatchString(rec.Message)
	}

	if systemOnly {
		return rec.IsSystemEntry(systemIncludeCheckpointer)
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
		duration, _ := util.ExtractDurationMillisFromLogMessage(rec.Message)
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

func GetLogRecordsBatchFromFile(filePath string, r *regexp.Regexp, useCsvFormat bool) <-chan []pglog.LogEntry {
	ch := make(chan []pglog.LogEntry, 3)
	go func() {
		defer close(ch)
		if strings.Contains(filePath, ".csv") || useCsvFormat {
			for batch := range GetLogRecordsFromCsvFile(filePath) {
				ch <- batch
			}
		} else {
			for batch := range GetLogRecordsFromLogFile(filePath, r) {
				ch <- batch
			}
		}
	}()
	return ch
}

// PeekRecordFromFile reads and returns just the first log record from a file without processing the entire file.
// This is useful for getting a sample of the log format or checking if the file is valid.
// Returns nil if no valid record is found or if the file is empty.
func PeekRecordFromFile(filePath string, logLineParsingRegex *regexp.Regexp, useCsvFormat bool) (*pglog.LogEntry, error) {
	log.Debug().Msgf("Peeking at first record from file: %s", filePath)

	if strings.Contains(filePath, ".csv") || useCsvFormat {
		return peekRecordFromCsvFile(filePath)
	}
	return peekRecordFromLogFile(filePath, logLineParsingRegex)
}

// peekRecordFromLogFile reads the first valid log entry from a plain text log file
func peekRecordFromLogFile(filePath string, logLineParsingRegex *regexp.Regexp) (*pglog.LogEntry, error) {
	var scanner *bufio.Scanner

	if filePath == "stdin" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
		}
		defer file.Close()

		var reader io.Reader = file
		if strings.HasSuffix(filePath, ".gz") {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				return nil, fmt.Errorf("error creating gzip reader for file %s: %w", filePath, err)
			}
			defer gzReader.Close()
			reader = gzReader
		}
		scanner = bufio.NewScanner(reader)
	}

	scanner.Split(bufio.ScanLines)

	var lines []string
	firstCompleteEntryFound := false

	for scanner.Scan() {
		line := scanner.Text()

		// If the line has a timestamp, it's potentially a new log entry
		if HasTimestampPrefix(line) {
			if firstCompleteEntryFound {
				// We have collected a complete entry, parse and return it
				entry, err := EventLinesToPgLogEntry(lines, filePath)
				if err != nil {
					return nil, fmt.Errorf("failed to parse first log entry: %w", err)
				}
				return &entry, nil
			}
			firstCompleteEntryFound = true
		}
		lines = append(lines, line)
	}

	// Handle case where there's only one entry in the file
	if firstCompleteEntryFound && len(lines) > 0 {
		entry, err := EventLinesToPgLogEntry(lines, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse first log entry: %w", err)
		}
		return &entry, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return nil, errors.New("no valid log entries found in file")
}

// peekRecordFromCsvFile reads the first valid log entry from a CSV log file
func peekRecordFromCsvFile(filePath string) (*pglog.LogEntry, error) {
	var reader io.Reader

	if filePath == "stdin" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
		}
		defer file.Close()

		reader = file
		if strings.HasSuffix(filePath, ".gz") {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				return nil, fmt.Errorf("error creating gzip reader for file %s: %w", filePath, err)
			}
			defer gzReader.Close()
			reader = gzReader
		}
	}

	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable fields

	// Read the first record
	record, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("CSV file is empty")
		}
		return nil, fmt.Errorf("error reading CSV record from %s: %w", filePath, err)
	}

	// Convert CSV record to LogEntry using the same logic as GetLogRecordsFromCsvFile
	if len(record) < 23 {
		return nil, errors.New("incomplete CSV record - expected at least 23 fields")
	}

	entry := pglog.LogEntry{
		LogTime:       record[0],
		ErrorSeverity: record[11],
		Message:       record[13],
		CsvColumns: &pglog.CsvEntry{ // Field order from https://www.postgresql.org/docs/current/file-fdw.html
			CsvColumnCount:       len(record),
			LogTime:              record[0],
			UserName:             record[1],
			DatabaseName:         record[2],
			ProcessID:            record[3],
			ConnectionFrom:       record[4],
			SessionID:            record[5],
			SessionLineNum:       record[6],
			CommandTag:           record[7],
			SessionStartTime:     record[8],
			VirtualTransactionID: record[9],
			TransactionID:        record[10],
			ErrorSeverity:        record[11],
			SQLStateCode:         record[12],
			Message:              record[13],
			Detail:               record[14],
			Hint:                 record[15],
			InternalQuery:        record[16],
			InternalQueryPos:     record[17],
			Context:              record[18],
			Query:                record[19],
			QueryPos:             record[20],
			Location:             record[21],
			ApplicationName:      record[22],
		},
	}
	//   v13 added backend_type
	//   v14 added leader_pid and query_id
	if len(record) >= 24 {
		entry.CsvColumns.BackendType = record[23]
	}
	if len(record) >= 26 {
		entry.CsvColumns.LeaderPid = record[24]
		entry.CsvColumns.QueryId = record[25]
	}

	return &entry, nil
}
