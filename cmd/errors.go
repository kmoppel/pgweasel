/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var MinErrLvl string
var errorsCmd = &cobra.Command{
	Use:   "errors [$LOG_FILE_OR_FOLDER]",
	Short: "Shows WARNING and higher entries by default",
	Run: func(cmd *cobra.Command, args []string) {
		showErrors(cmd, args)
	},
	Aliases: []string{"err", "errs", "error"},
}
var TopNErrorsOnly int
var errorsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Top N errors",
	Run:   showErrors,
}

type TopErrorsCollector struct {
	NormalizedErrorCounts map[string]int
}

func (tec *TopErrorsCollector) AddError(level string, message string) {
	// uti
	tec.NormalizedErrorCounts[message]++
}

type TopError struct {
	Level   string
	Message string
	Count   int
}

func (tec *TopErrorsCollector) GetTopN(topN int) []TopError {
	// sort
	// print
	topErrors := make([]TopError, 0, topN)
	return topErrors
}

func init() {
	errorsCmd.AddCommand(errorsTopCmd)
	errorsTopCmd.Flags().IntVarP(&TopNErrorsOnly, "top", "", 0, "Nr of top errors to show")

	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	var logFiles []string
	var topErrors *TopErrorsCollector = &TopErrorsCollector{NormalizedErrorCounts: make(map[string]int)}

	cfg := PreProcessArgs(cmd, args)

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, From=%s, To=%s, SystemOnly=%v, TopNErrors=%d", cfg.MinErrLvl, cfg.MinSlowDurationMs, cfg.FromTime, cfg.ToTime, cfg.SystemOnly, TopNErrorsOnly)

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

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)

		for rec := range logparser.GetLogRecordsFromFile(logFile, cfg.LogLineRegex) {
			if rec.ErrorSeverity == "" {
				log.Error().Msgf("Got invalid entry: %+v", rec)
				continue
			}
			log.Debug().Msgf("Processing log entry: %+v", rec)

			if TopNErrorsOnly > 0 {
				topErrors.AddError(rec.ErrorSeverity, rec.Message)
			} else {
				if logparser.DoesLogRecordSatisfyUserFilters(rec, cfg.MinErrLvl, Filters, cfg.FromTime, cfg.ToTime, cfg.MinSlowDurationMs, cfg.SystemOnly) {
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

	if TopNErrorsOnly > 0 {
		topErrorsList := topErrors.GetTopN(TopNErrorsOnly)
		for _, topError := range topErrorsList {
			w.WriteString(topError.Level + ": " + topError.Message + " (" + strconv.Itoa(topError.Count) + ")\n")
		}
	}

	w.Flush()
}
