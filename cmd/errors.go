/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
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

	logFile, logFolder, err := detector.GetLatestLogFileAndFolder(args, Connstr)

	if err != nil {
		log.Fatal().Err(err).Msg("Error determining any log files")
	}
	if logFile == "" {
		log.Error().Msg("No log files found")
		return
	}

	log.Debug().Msgf("Detected logFolder: %s, logFile: %s, MinErrLvl: %s, Filters: %v", logFolder, logFile, MinErrLvl, Filters)

	logparser.ShowErrors(logFile, MinErrLvl, Filters)
}
