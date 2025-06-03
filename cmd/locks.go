package cmd

import (
	"github.com/spf13/cobra"
)

var LocksOnly bool

var locksCmd = &cobra.Command{
	Use:   "locks [$LOG_FILE_OR_FOLDER]",
	Short: "Only show locking related entries",
	Run: func(cmd *cobra.Command, args []string) {
		LocksOnly = true
		showErrors(cmd, args)
	},
	Aliases: []string{"loc", "lock", "deadlock", "deadlocks"},
}

func init() {
	rootCmd.AddCommand(locksCmd)
}
