/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var MinErrLvl string
var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Shows WARNING+ log entries by default",
	Run: func(cmd *cobra.Command, args []string) {
		showErrors(cmd, args)
	},
	Aliases: []string{"err", "errs", "error"},
}

func init() {
	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	var logFiles = make([]string, 0)
	var logFile string
	var logFolder string
	var err error
	var fromTime time.Time
	var toTime time.Time
	var logLineRegex *regexp.Regexp

	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	MinErrLvl = strings.ToUpper(MinErrLvl)

	if LogLineRegex != "" {
		log.Debug().Msgf("Using regex to parse plain text entries: %s", LogLineRegex)
		if !strings.Contains(LogLineRegex, "<log_time>") || !strings.Contains(LogLineRegex, "<error_severity>") || !strings.Contains(LogLineRegex, "<message>") {
			log.Fatal().Msgf("Custom regex needs to have groups: log_time, error_severity, message. Default regex: %s", logparser.DEFAULT_REGEX_STR)
		}
		logLineRegex = regexp.MustCompile(LogLineRegex)
	}

	if From != "" {
		fromTime, err = util.HumanTimeOrDeltaStringToTime(From, time.Time{})
		if err != nil {
			log.Warn().Msg("Error parsing --from timedelta input, supported units are 's', 'm', 'h'. Ignoring --from")
		}
	}
	if To != "" {
		toTime, err = util.HumanTimeOrDeltaStringToTime(To, time.Time{})
		if err != nil {
			log.Warn().Msg("Error parsing --to timedelta input, supported units are 's', 'm', 'h'. Ignoring --to")
		}
	}

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, From=%s, To=%s, SystemOnly=%v", MinErrLvl, MinSlowDurationMs, fromTime, toTime, SystemOnly)

	if len(args) == 0 {
		log.Debug().Msg("No files / folders provided, looking for latest file from default locations ...")
		logFile, logFolder, err = detector.DiscoverLatestLogFileAndFolder(nil, Connstr)
		if err != nil {
			log.Fatal().Msgf("Failed to detect any log files from default locations")
		}
		logFiles = append(logFiles, logFile)
	} else {
		for _, arg := range args {
			log.Debug().Msgf("Checking input path: %s ...", arg)
			_, err = os.Stat(arg)
			if err != nil {
				log.Error().Err(err).Msgf("Error accessing path: %s", arg)
				continue
			}

			if util.IsPathExistsAndFile(arg) {
				logFile = arg
				logFolder = filepath.Base(arg)
				logFiles = append(logFiles, logFile)
			} else {
				log.Debug().Msgf("Looking for log files from folder: %s ..", arg)
				logFiles, err = util.GetPostgresLogFilesTimeSorted(arg)
				if err != nil {
					log.Error().Err(err).Msgf("Error determining any log files from folder: %s", arg)
					continue
				}
				log.Debug().Msgf("Found %d log files", len(logFiles))
			}
		}
	}

	if len(logFiles) == 0 {
		log.Error().Msg("No log files found")
		return
	}

	log.Debug().Msgf("Detected logFolder: %s, logFile: %s, MinErrLvl: %s, Filters: %v", logFolder, logFile, MinErrLvl, Filters)

	// Create a buffered writer for better performance
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)
		continue

		for rec := range logparser.GetLogRecordsFromFile(logFile, logLineRegex) {
			log.Debug().Msgf("Processing log entry: %+v", rec)
			if rec.ErrorSeverity != "" {
				if logparser.DoesLogRecordSatisfyUserFilters(rec, MinErrLvl, Filters, fromTime, toTime, MinSlowDurationMs, SystemOnly) {
					if rec.CsvColumns != nil {
						w.WriteString(rec.CsvColumns.String())
					} else {
						w.WriteString(strings.Join(rec.Lines, "\n"))
					}
					w.WriteByte('\n')
				}
			}
		}
		log.Debug().Msgf("Finished processing log file: %s", logFile)
	}
	w.Flush()
}
