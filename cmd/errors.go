/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		showErrors(cmd, args)
	},
	Args: cobra.MaximumNArgs(1), // empty means outdetect or use hardcoded defaults
}

func init() {
	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	MinErrLvl = strings.ToUpper(MinErrLvl)
	log.Debug().Msgf("Running in debug mode. MinErrLvl = %s", MinErrLvl)

	var logFiles = make([]string, 0)
	var logFile string
	var logFolder string
	var err error
	var fromTime time.Time
	var toTime time.Time

	if From != "" {
		fromTime, err = util.HumanTimedeltaToTime(From)
		if err != nil {
			log.Warn().Msg("Error parsing --from timedelta input, supported units are 's', 'm', 'h'. Ignoring --from")
		}
	}
	if To != "" {
		toTime, err = util.HumanTimedeltaToTime(To)
		if err != nil {
			log.Warn().Msg("Error parsing --to timedelta input, supported units are 's', 'm', 'h'. Ignoring --to")
		}
	}

	if len(args) == 0 {
		log.Debug().Msg("No files / folders provided, looking for latest file from default locations ...")
		logFile, logFolder, err = detector.DiscoverLatestLogFileAndFolder(args, Connstr)
		logFiles = append(logFiles, logFile)
	} else {
		_, err := os.Stat(args[0])
		if err != nil {
			log.Fatal().Err(err).Msgf("Error accessing path: %s", args[0])
		}

		if util.IsPathExistsAndFile(args[0]) {
			logFile = args[0]
			logFolder = filepath.Base(args[0])
			logFiles = append(logFiles, logFile)
		} else {
			log.Debug().Msgf("Looking for log files from folder: %s ..", args[0])
			logFiles, err = util.GetPostgresLogFilesTimeSorted(args[0])
			log.Debug().Msgf("Found: %d", len(logFiles))
		}
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Error determining any log files")
	}
	if len(logFiles) == 0 {
		log.Error().Msg("No log files found")
		return
	}

	log.Debug().Msgf("Detected logFolder: %s, logFile: %s, MinErrLvl: %s, Filters: %v", logFolder, logFile, MinErrLvl, Filters)

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)
		if strings.HasSuffix(logFile, ".csv") {
			logparser.ShowErrorsCsv(logFile, MinErrLvl, Filters)
		} else {
			logparser.ShowErrors(logFile, MinErrLvl, Filters, fromTime, toTime)
		}
	}
}
