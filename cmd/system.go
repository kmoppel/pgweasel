package cmd

import (
	"github.com/spf13/cobra"
)

var SystemOnly bool

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show messages by Postgres internal processes",
	Run: func(cmd *cobra.Command, args []string) {
		SystemOnly = true
		showErrors(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
