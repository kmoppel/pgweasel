package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/spf13/cobra"
)

var MinSlowDurationMs int

var slowCmd = &cobra.Command{
	Use:   "slow $MIN_DURATION [$LOG_FILE_OR_FOLDER]",
	Short: "Show queries above user set threshold, e.g.: pgweasel slow 1s mylogfile.log",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		MinSlowDurationMs, err = util.IntervalToMillis(args[0])
		if err != nil {
			log.Fatal("Failed to convert $MIN_DURATION input to milliseconds")
		}
		args = args[1:]
		MinErrLvl = "DEBUG5" // Override default WARNING+ output level
		showErrors(cmd, args)
	},
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"slo"},
}

var SlowTopN int
var SlowTopNOnly bool

var slowTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Show Top N slowest queries",
	Run: func(cmd *cobra.Command, args []string) {
		SlowTopNOnly = true
		showErrors(cmd, args)
	},
}

func init() {
	slowTopCmd.Flags().IntVarP(&SlowTopN, "top", "", 10, "Top slowest queries to show")

	slowCmd.AddCommand(slowTopCmd)

	rootCmd.AddCommand(slowCmd)
}
