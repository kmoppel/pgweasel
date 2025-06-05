package cmd

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var PeakBucketIntervalStr string = "10m" // Default bucket interval as a string
var PeakBucketDuration time.Duration

var peaksCmd = &cobra.Command{
	Use:   "peaks [--bucket $BUCKET_INTERVAL] [$LOG_FILE_OR_FOLDER]",
	Short: "Identify periods where most log entries are emitted, per severity level",
	Run: func(cmd *cobra.Command, args []string) {
		Peaks = true
		var err error
		PeakBucketDuration, err = time.ParseDuration(PeakBucketIntervalStr)
		if err != nil {
			log.Fatal().Err(err).Msgf("Invalid --bucket input: %s", PeakBucketIntervalStr)
		}
		showErrors(cmd, args)
	},
	Aliases: []string{"peak", "busy"},
}

var TopCmd = &cobra.Command{
	Use:   "top",
	Short: "Top N errors",
	Run: func(cmd *cobra.Command, args []string) {
		TopNErrorsOnly = true
		showErrors(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(peaksCmd)

	peaksCmd.Flags().StringVarP(&PeakBucketIntervalStr, "bucket", "b", "10m", "Bucket interval")

	peaksCmd.ad
}
