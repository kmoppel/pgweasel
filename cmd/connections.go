package cmd

import (
	"github.com/spf13/cobra"
)

var ConnectionsSummary bool

var connsCmd = &cobra.Command{
	Use:   "connections [$LOG_FILE_OR_FOLDER]",
	Short: "Show connections summary",
	Run: func(cmd *cobra.Command, args []string) {
		ConnectionsSummary = true
		showErrors(cmd, args)
	},
	Aliases: []string{"conns", "conn"},
}

func init() {
	rootCmd.AddCommand(connsCmd)
}
