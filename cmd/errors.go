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
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var MinErrLvl string
var logger *zap.Logger
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
		logger = zap.Must(zap.NewDevelopment())
		defer logger.Sync() // flushes buffer, if any
	} else {
		logger = zap.NewNop()
	}

	MinErrLvl = strings.ToUpper(MinErrLvl)
	logger.Debug("Running in debug mode", zap.String("MinErrLvl", MinErrLvl))

	defaultLogFolder := "/var/log/postgresql"
	var logFolder, logFile, logDest string
	var err error

	log.Println("len(args)", len(args))
	if len(args) == 1 {
		logFile, logFolder, err = detector.DetectLatestPostgresLogFileAndFolder(args[0])
	} else {
		if Connstr != "" {
			logger.Debug("Using --connstr for log location / prefix ...")
			logDest, logFolder, Prefix, err = postgres.GetLogDestAndDirectoryAndPrefix()
			if err != nil {
				logger.Error("Error getting log directory and prefix from DB", zap.Error(err))
			}
			logger.Sugar().Debug("logDest", logDest, "logFolder", logFolder, "Prefix", Prefix)
		}
		if logFolder == "" {
			logFolder = defaultLogFolder
		}
		logFile, logFolder, err = detector.DetectLatestPostgresLogFileAndFolder(logFolder)
	}
	if err != nil {
		logger.Error("Error determining log files", zap.Error(err))
		return
	}
	logger.Sugar().Debugf("logFile: %s, logFolder: %s", logFile, logFolder)
	if logFile == "" {
		logger.Error("No log files found")
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
