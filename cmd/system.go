package cmd

import (
	"github.com/spf13/cobra"
)

var SystemOnly bool
var SystemIncludeCheckpointer bool

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show messages by Postgres internal processes",
	Run: func(cmd *cobra.Command, args []string) {
		SystemOnly = true
		showErrors(cmd, args)
	},
	Aliases: []string{"sys", "pg", "postgre", "postgres", "postmaster"},
}

func init() {
	systemCmd.Flags().BoolVarP(&SystemIncludeCheckpointer, "checkpointer", "c", false, "Include checkpointer events as well")

	rootCmd.AddCommand(systemCmd)
}
