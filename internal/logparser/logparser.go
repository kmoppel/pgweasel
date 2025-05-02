package logparser

import (
	"errors"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/spf13/cobra"
)

func ParseEntryFromLogline(line string, logLinePrefix string) (pglog.LogEntry, error) {
	e := pglog.LogEntry{}
	if line == "" {
		return e, errors.New("empty log line")
	}

	r := CompileRegexForLogLinePrefix(logLinePrefix)
	log.Printf("Parsing prefix '%s', line: %s", logLinePrefix, line)
	// log.Println("FindAllString", r.FindAllString(line, -1))

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

	return e, nil
}

// Convert a time string like "2025-04-28 00:20:02.274 EEST" to a time.Time object
func TimestringToTime(s string) (time.Time, error) {
	layout := "2006-01-02 15:04:05.000 MST"
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Fatalf("Failed to parse time string '%s': %v", s, err)
	}
	return t, err
}

// 2025-04-28 00:20:02.274 EEST [2635] LOG:  checkpoint starting: time
func CompileRegexForLogLinePrefix(logLinePrefix string) *regexp.Regexp {
	log.Printf("CompileRegexForLogLinePrefix for logLinePrefix: '%s'\n", logLinePrefix)
	var r = "^" + logLinePrefix
	r = strings.Replace(r, "[", "\\[", -1)
	r = strings.Replace(r, "]", "\\]", -1)
	r = strings.Replace(r, "%m", `(?P<time>[\d\-:\. ]+ [A-Z]+)`, -1)
	r = strings.Replace(r, "%p", `(?P<pid>\d+)`, -1)
	r = strings.Replace(r, "%q", "", -1)
	r = strings.Replace(r, "%u", `(?P<user>\w+)?`, -1)
	r = strings.Replace(r, "%d", `(?P<db>\w+)?`, -1)
	r = r + `(?P<level>[A-Z]+):(?P<message>.*)$`
	// `^(?P<time>[\d\-:\. ]+ [A-Z]+) \[(?P<pid>\d+)\] (?:(?P<session>[\w\.\[\]]+)\s)?(?P<user>\w+)?@(?P<db>\w+)?`
	// r = `^(?P<time>[\d\-:\. ]+ [A-Z]+) `
	log.Println("Final regex str:", r)
	return regexp.MustCompile(r)
}

// Handle multi-line entries, collect all lines until a new entry starts and then parse
func ParseLogFile(cmd *cobra.Command, filePath string, logLines []string, logLinePrefix string) error {
	minLvl, _ := cmd.Flags().GetString("min-lvl")
	log.Println("Showing all msgs with minLvl >=", minLvl)
	return nil
}
