/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/postgres"
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

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-lvl", "", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	MinErrLvl = strings.ToUpper(MinErrLvl)
	log.Debug().Msgf("Running in debug mode. MinErrLvl = %s", MinErrLvl)

	logFile, logFolder, err := detector.GetLatestLogFileAndFolder(args, Connstr)

	if err != nil {
		log.Fatal().Err(err).Msg("Error determining any log files")
	}
	if logFile == "" {
		log.Error().Msg("No log files found")
		return
	}
	if Connstr != "" && Prefix == "" {
		Prefix, err = postgres.GetLogLinePrefix(Connstr)
		if err != nil {
			log.Fatal().Err(err).Msg("Error determining log_line_prefix from Connstr")
		}
	}

	if Prefix == "" {
		Prefix = detector.DEFAULT_LOG_LINE_PREFIX
		log.Warn().Msgf("Using default log_line_prefix: %s (use -p/--prefix to set)", detector.DEFAULT_LOG_LINE_PREFIX)
	}

	log.Debug().Msgf("Detected logFolder: %s, logFile: %s, Prefix: %s", logFolder, logFile, Prefix)

	logparser.ParseLogFile(cmd, logFile, nil, Prefix, MinErrLvl)
}
