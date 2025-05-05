/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"strings"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/postgres"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
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
}

func init() {
	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-lvl", "", "ERROR", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	MinErrLvl = strings.ToUpper(MinErrLvl)
	zl.Debug().Msgf("Running in debug mode. MinErrLvl = %s", MinErrLvl)

	defaultLogFolder := "/var/log/postgresql"
	var logFolder, logFile, logDest string
	var err error

	log.Println("len(args)", len(args))
	if len(args) == 1 {
		logFile, logFolder, err = detector.DetectLatestPostgresLogFileAndFolder(args[0])
	} else {
		if Connstr != "" {
			zl.Debug().Msg("Using --connstr for log location / prefix ...")
			logDest, logFolder, Prefix, err = postgres.GetLogDestAndDirectoryAndPrefix()
			if err != nil {
				zl.Error().Msgf("Error getting log directory and prefix from DB: %v", err)
			}
			zl.Debug().Msgf("logDest: %s, logFolder: %s, Prefix: %s", logDest, logFile, Prefix)
		}
		if logFolder == "" {
			logFolder = defaultLogFolder
		}
		logFile, logFolder, err = detector.DetectLatestPostgresLogFileAndFolder(logFolder)
	}
	if err != nil {
		zl.Error().Msgf("Error determining log files: %v", err)
		return
	}
	zl.Debug().Msgf("logDest: %s, logFolder: %s, Prefix: %s", logDest, logFile, Prefix)
	if logFile == "" {
		zl.Error().Msg("No log files found")
		return
	}

	logparser.ParseLogFile(cmd, "testdata/debian_default.log", nil, Prefix)
	// logparser.ParseLogFile(cmd, "testdata/rds_default.log", nil, "%t:%r:%u@%d:[%p]")

	// var log1 = `2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`
	// e, err := logparser.ParseEntryFromLogline(log1, Prefix)
	// if err != nil {
	// 	log.Println("Error in ParseEntryFromLogline:", err)
	// 	return
	// }
	// log.Printf("e: %+v\n", e)

	// var log2 = `2025-05-02 12:26:27.649 EEST [2308351] LOG:  background worker "TimescaleDB Background Worker Scheduler" (PID 2380175) exited with exit code 1`
	// e, err = logparser.ParseEntryFromLogline(log2, Prefix)
	// if err != nil {
	// 	log.Println("Error in ParseEntryFromLogline:", err)
	// 	return
	// }
	// log.Printf("e: %+v\n", e)
}
