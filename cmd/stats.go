package cmd

import (
	"github.com/spf13/cobra"
)

var Stats bool

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Summary of log events",
	Run: func(cmd *cobra.Command, args []string) {
		Stats = true
		showErrors(cmd, args)
	},
	Aliases: []string{"stat", "statistics"},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
