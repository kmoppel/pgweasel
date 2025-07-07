package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/spf13/cobra"
)

var ShowPlans bool

// To show only auto-explained plans, use the `pgweasel plans` command.
var plansCmd = &cobra.Command{
	Use:   "plans $MIN_DURATION_MS [$LOG_FILE_OR_FOLDER]",
	Short: "Show matching log entries only, e.g.: pgweasel grep 'Seq Scan.*tblX' mylogfile.log",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		MinSlowDurationMs, err = util.IntervalToMillis(args[0])
		if err != nil {
			log.Fatal("Failed to convert $MIN_DURATION_MS input to milliseconds")
		}
		args = args[1:]
		MinErrLvl = "DEBUG5" // Override default WARNING+ output level
		ShowPlans = true
		showErrors(cmd, args)
	},
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"plan", "explain"},
}

func init() {
	rootCmd.AddCommand(plansCmd)
}
