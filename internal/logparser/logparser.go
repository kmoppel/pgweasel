package logparser

import (
	"bufio"
	"errors"
	stdlog "log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func ParseEntryFromLogline(line string, r *regexp.Regexp) (pglog.LogEntry, error) {
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

	timestamp, err := TimestringToTime(result["time"])
	if err != nil {
		return e, err
	}
	e.LogTime = timestamp

	e.ProcessID = result["pid"]

	e.UserName = result["user"]

	e.DatabaseName = result["db"]

	e.ErrorSeverity = result["level"]

	e.Message = result["message"]

	e.ConnectionFrom = result["remote"]

	return e, nil
}

// Convert a time string like "2025-04-28 00:20:02.274 EEST" to a time.Time object
func TimestringToTime(s string) (time.Time, error) {
	layout := "2006-01-02 15:04:05.000 MST"

	t, err := time.Parse(layout, s)
	if err != nil {
		if err != nil {
			log.Error().Msgf("Failed to parse time string '%s' with layout: %s", s, layout)
		}
		layout = "2006-01-02 15:04:05 MST"
		t, err = time.Parse(layout, s)
		if err != nil {
			log.Error().Msgf("Failed to parse time string '%s' with layout: %s", s, layout)
		}
	}
	return t, err
}

// 2025-04-28 00:20:02.274 EEST [2635] LOG:  checkpoint starting: time
func CompileRegexForLogLinePrefix(logLinePrefix string) *regexp.Regexp {
	// log.Printf("CompileRegexForLogLinePrefix for logLinePrefix: '%s'\n", logLinePrefix)
	var r = "^" + logLinePrefix
	r = strings.Replace(r, "[", "\\[", -1)
	r = strings.Replace(r, "]", "\\]", -1)
	r = strings.Replace(r, "%m", `(?P<time>[\d\-:\. ]+ [A-Z]+)`, -1) // 2025-05-02 18:25:05.617 EEST
	r = strings.Replace(r, "%t", `(?P<time>[\d\-:\. ]+ [A-Z]+)`, -1) // 2025-05-05 06:00:51 UTC
	r = strings.Replace(r, "%r", `(?P<remote>[\w\-\.]+\(\d+\))`, -1) // 127.0.0.1(32890)
	r = strings.Replace(r, "%p", `(?P<pid>\d+)`, -1)
	r = strings.Replace(r, "%q%u@%d", `(?:(?P<user>\w+)@(?P<db>\w+))?`, -1)
	r = strings.TrimRight(r, " ")
	r = strings.Replace(r, "%q", "", -1)
	r = strings.Replace(r, "%u", `(?P<user>\w+)`, -1)
	r = strings.Replace(r, "%d", `(?P<db>\w+)`, -1)
	r = r + `:?\s*(?P<level>[A-Z]+):\s*(?P<message>(?s:.*))$`
	// `^(?P<time>[\d\-:\. ]+ [A-Z]+) \[(?P<pid>\d+)\] (?:(?P<session>[\w\.\[\]]+)\s)?(?P<user>\w+)?@(?P<db>\w+)?`
	// log.Println("Final regex str:", r)
	// os.Exit(0)
	return regexp.MustCompile(r)
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func ParseLogFile(cmd *cobra.Command, filePath string, logLines []string, logLinePrefix string) error {
	minLvl, _ := cmd.Flags().GetString("min-lvl")

	// Open file from filePath and loop line by line
	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Msgf("Error opening file %s", filePath)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines = make([]string, 0)
	var r *regexp.Regexp

	gathering := false
	for scanner.Scan() {
		line := scanner.Text()

		// If the line does not have a timestamp, it is a continuation of the previous entry
		if HasTimestampPrefix(line) {
			if gathering && len(lines) > 0 {
				var final string
				if len(lines) > 1 {
					final = strings.Join(lines, "\n")
				} else {
					final = lines[0]
				}
				if r == nil {
					r = CompileRegexForLogLinePrefix(logLinePrefix)
				}

				e, err := ParseEntryFromLogline(final, r)
				if err != nil {
					log.Fatal().Err(err).Msg("Error in ParseEntryFromLogline")
				} else {
					if e.SeverityNum() >= pglog.SeverityToNum(minLvl) {
						stdlog.Printf("Parsed entry: %+v\n", e)
					}
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

func HasTimestampPrefix(line string) bool {
	r := regexp.MustCompile(`^(?P<time>[\d\-:\. ]+ [A-Z]+)`)
	return r.MatchString(line)
}
