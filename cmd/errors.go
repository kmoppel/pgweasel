/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var MinErrLvl = "WARNING"
var errorsCmd = &cobra.Command{
	Use:   "errors [$LOG_FILE_OR_FOLDER]",
	Short: "Shows WARNING and higher entries by default",
	Run: func(cmd *cobra.Command, args []string) {
		showErrors(cmd, args)
	},
	Aliases: []string{"err", "errs", "error"},
}
var TopNErrors int
var TopNErrorsOnly bool

var errorsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Top N errors",
	Run: func(cmd *cobra.Command, args []string) {
		TopNErrorsOnly = true
		showErrors(cmd, args)
	},
}

type ErrorLevelMessage struct {
	ErrorLevel string
	Message    string
}

type TopErrorsCollector struct {
	ErrorCounts map[string]map[string]int // [ERROR][MESSAGE] = COUNT
	TotalCount  int
}

func (tec *TopErrorsCollector) Initialize() {
	tec.ErrorCounts = make(map[string]map[string]int)
	for _, level := range pglog.ERROR_SEVERITIES {
		tec.ErrorCounts[level] = make(map[string]int)
	}
}

func (tec *TopErrorsCollector) AddError(level string, message string) {
	tec.ErrorCounts[level][message]++
}

type TopError struct {
	Level   string
	Message string
	Count   int
}

func (tec *TopErrorsCollector) GetTopN(topN int) []TopError {
	// flatten before sorting
	topErrors := make([]TopError, 0)
	for level, messages := range tec.ErrorCounts {
		for message, count := range messages {
			topErrors = append(topErrors, TopError{
				Level:   level,
				Message: message,
				Count:   count,
			})
		}
	}
	// sort by count
	sort.Slice(topErrors, func(i, j int) bool {
		return topErrors[i].Count > topErrors[j].Count
	})
	return topErrors[:topN]
}

func init() {
	errorsCmd.AddCommand(errorsTopCmd)
	errorsTopCmd.Flags().IntVarP(&TopNErrors, "top", "", 10, "Nr of top errors to show")

	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	var logFiles []string
	var topErrors *TopErrorsCollector = &TopErrorsCollector{}
	topErrors.Initialize()

	cfg := PreProcessArgs(cmd, args)

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, From=%s, To=%s, SystemOnly=%v, TopNErrorsOnly=%t", cfg.MinErrLvl, cfg.MinSlowDurationMs, cfg.FromTime, cfg.ToTime, cfg.SystemOnly, TopNErrorsOnly)

	if len(args) == 0 && util.CheckStdinAvailable() {
		logFiles = []string{"stdin"}
	} else {
		logFiles = util.GetLogFilesFromUserArgs(args)
		if len(logFiles) == 0 {
			log.Error().Msg("No log files found to process")
			return
		}
	}

	// Create a buffered writer for better performance
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	minErrLvlSeverityNum := pglog.SeverityToNum(MinErrLvl)

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)

		for rec := range logparser.GetLogRecordsFromFile(logFile, cfg.LogLineRegex, cfg.ForceCsvInput) {
			if rec.ErrorSeverity == "" {
				log.Error().Msgf("Got invalid entry: %+v", rec)
				continue
			}
			log.Debug().Msgf("Processing log entry: %+v", rec)

			if TopNErrorsOnly && rec.SeverityNum() >= minErrLvlSeverityNum {
				topErrors.AddError(rec.ErrorSeverity, rec.Message)
			} else {
				if logparser.DoesLogRecordSatisfyUserFilters(rec, cfg.MinErrLvlNum, Filters, cfg.FromTime, cfg.ToTime, cfg.MinSlowDurationMs, cfg.SystemOnly) {
					if rec.CsvColumns != nil {
						w.WriteString(rec.CsvColumns.String())
					} else {
						w.WriteString(strings.Join(rec.Lines, "\n"))
					}
					w.WriteByte('\n')
				}
				if Verbose {
					w.Flush()
				}
			}
		}
		log.Debug().Msgf("Finished processing log file: %s", logFile)
	}

	if TopNErrorsOnly {
		for _, topError := range topErrors.GetTopN(TopNErrors) {
			w.WriteString(strconv.Itoa(topError.Count) + "x " + topError.Level + ": " + topError.Message + "\n")
		}
	}

	w.Flush()
}
