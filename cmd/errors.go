/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
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
	logger.Debug("DEV Hello from Zap!")
	logger.Sugar().Debugf("DEV Hello from Zap! %s", "Hello from Zap!")
	log.Println("showErrors called")
	log.Println("MinErrLvl", MinErrLvl)
	lastLog, _ := detector.DetectLatestPostgresLogFile() //TODO glob
	log.Println("lastLog", lastLog)
	gp, _ := cmd.Flags().GetString("glob")
	log.Println("GlobPath", gp)
	// logparser.ParseLogFile(cmd, lastLog, nil, Prefix)
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
