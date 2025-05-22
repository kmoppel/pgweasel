/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/util"
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
	cfg := PreProcessArgs(cmd, args)

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, From=%s, To=%s, SystemOnly=%v", cfg.MinErrLvl, cfg.MinSlowDurationMs, cfg.FromTime, cfg.ToTime, cfg.SystemOnly)

	logFiles := util.GetLogFilesFromUserArgs(args, Connstr)

	if len(logFiles) == 0 {
		log.Error().Msg("No log files found to process")
		return
	}

	// Create a buffered writer for better performance
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)
		continue

		for rec := range logparser.GetLogRecordsFromFile(logFile, cfg.LogLineRegex) {
			log.Debug().Msgf("Processing log entry: %+v", rec)
			if rec.ErrorSeverity != "" {
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
	w.Flush()
}
