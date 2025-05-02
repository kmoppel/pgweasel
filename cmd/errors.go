/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/detector"
	"github.com/kmoppel/pgweasel/internal/logparser"
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
	log.Println("showErrors called")
	log.Println("MinErrLvl", MinErrLvl)
	lastLog, _ := detector.DetectLatestPostgresLogFile()
	log.Println("lastLog", lastLog)
	gp, _ := cmd.Flags().GetString("glob")
	log.Println("GlobPath", gp)

	var log1 = `2025-05-02 12:27:52.634 EEST [2380404] krl@pgwatch2_metrics ERROR:  column "asdasd" does not exist at character 8`
	e, _ := logparser.ParseEntryFromLogline(log1, Prefix)
	log.Println("e", e)
}
