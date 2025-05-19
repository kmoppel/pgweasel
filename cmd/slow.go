package cmd

import (
	"log"

	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/spf13/cobra"
)

var MinSlowDurationMs int

var slowCmd = &cobra.Command{
	Use:   "slow $MIN_DURATION [$LOG_FILE_OR_FOLDER]",
	Short: "Show slow queries only",
	Long:  "Show queries taking longer than input $MIN_DURATION",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		MinSlowDurationMs, err = util.IntervalToMillis(args[0])
		if err != nil {
			log.Fatal("Failed to convert $MIN_DURATION input to milliseconds")
		}
		args = args[1:]
		MinErrLvl = "DEBUG5"
		showErrors(cmd, args)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(slowCmd)
}
