package logparser

import (
	"bufio"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/pglog"
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
		layout = "2006-01-02 15:04:05 MST"
		t, err = time.Parse(layout, s)
		if err != nil {
			log.Fatalf("Failed to parse time string '%s': %v", s, err)
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
		log.Println("Error opening file:", err)
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
					// log.Println("Found multi-line entry:", final)
				} else {
					final = lines[0]
				}
				if r == nil {
					r = CompileRegexForLogLinePrefix(logLinePrefix)
				}
				// log.Printf("Parsing prefix '%s', line: %s", logLinePrefix, line)
				// Convert the string builder to a string and parse it
				e, err := ParseEntryFromLogline(final, r)
				if err != nil {
					log.Println("Error in ParseEntryFromLogline:", err)
				} else {
					if e.SeverityNum() >= pglog.SeverityToNum(minLvl) {
						log.Println("Found line with minLvl:", e.ErrorSeverity, final)
						// log.Println("Found line with minLvl:", e.ErrorSeverity, final)
						// Here you can do something with the parsed entry, like storing it in a database or printing it
						// log.Printf("Parsed entry: %+v\n", e)
						log.Printf("Parsed entry: %+v\n", e)
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

	// // Loop through all lines in the log file from filePath, or logLines if logLines is not empty
	// for _, line := range logLines {
	// 	if strings.Contains(line, minLvl) {
	// 		log.Println("Found line with minLvl:", line)
	// 		e, err := ParseEntryFromLogline(line, logLinePrefix)
	// 		if err != nil {
	// 			log.Println("Error parsing log line:", err)
	// 			continue
	// 		}
	// 		log.Printf("Parsed entry: %+v\n", e)
	// 		// Here you can do something with the parsed entry, like storing it in a database or printing it
	// 		// log.Printf("Parsed entry: %+v\n", e)
	// 		// For example, print the entry
	// 	}
	// }
	return nil
}

func HasTimestampPrefix(line string) bool {
	r := regexp.MustCompile(`^(?P<time>[\d\-:\. ]+ [A-Z]+)`)
	return r.MatchString(line)
}
