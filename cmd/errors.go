/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/pgdetector"
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
	lastLog, _ := pgdetector.DetectLatestPostgresLogFile()
	log.Println("lastLog", lastLog)
}
