package cmd

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pgweasel",
	Short: "A simplistic PostgreSQL log parser for the console",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var Verbose bool
var From string
var To string
var Oneline bool
var Filters []string
var Connstr string
var Tail bool
var LogLineRegex string
var Csv bool
var Peaks bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More chat")
	rootCmd.PersistentFlags().StringVarP(&From, "from", "", "", "Log entries from $time, e.g.: -1h")
	rootCmd.PersistentFlags().StringVarP(&To, "to", "", "", "Log entries up to $time")
	rootCmd.PersistentFlags().BoolVarP(&Oneline, "oneline", "1", false, "Compact multiline entries")
	rootCmd.PersistentFlags().StringArrayVarP(&Filters, "filter", "f", nil, "Add extra line match conditions (regex)")
	// rootCmd.PersistentFlags().BoolVarP(&Tail, "tail", "t", false, "Keep watching the log file for new entries")
	rootCmd.PersistentFlags().StringVarP(&LogLineRegex, "regex", "", logparser.DEFAULT_REGEX_STR, "Use a custom regex instead of:")
	rootCmd.PersistentFlags().BoolVarP(&Csv, "csv", "", false, "Specify that input file or stdin is actually CSV regardless of file extension")
	rootCmd.PersistentFlags().BoolVarP(&Peaks, "peaks", "", false, "Show only event counts per log level for peak load periods")
}

type WeaselConfig struct {
	FromTime                time.Time
	ToTime                  time.Time
	LogLineRegex            *regexp.Regexp
	MinErrLvl               string
	MinErrLvlNum            int
	SystemOnly              bool
	MinSlowDurationMs       int
	SlowTopN                int
	SlowTopNOnly            bool
	ForceCsvInput           bool
	Oneline                 bool
	PeaksOnly               bool
	LocksOnly               bool
	StatsOnly               bool
	ErrorsHistogram         bool
	HistogramBucketDuration time.Duration
}

func PreProcessArgs(cmd *cobra.Command, args []string) WeaselConfig {
	var err error
	var fromTime time.Time
	var toTime time.Time
	var logLineRegex *regexp.Regexp

	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	minErrLvl := strings.ToUpper(MinErrLvl)

	if LogLineRegex != "" {
		log.Debug().Msgf("Using regex to parse plain text entries: %s", LogLineRegex)
		if !strings.Contains(LogLineRegex, "<log_time>") || !strings.Contains(LogLineRegex, "<error_severity>") || !strings.Contains(LogLineRegex, "<message>") {
			log.Fatal().Msgf("Custom regex needs to have groups: log_time, error_severity, message. Default regex: %s", logparser.DEFAULT_REGEX_STR)
		}
		logLineRegex, err = regexp.Compile(LogLineRegex)
		if err != nil {
			log.Fatal().Msgf("Failed to compile provided regex: %s", LogLineRegex)
		}
	}

	if From != "" {
		fromTime, err = util.HumanTimeOrDeltaStringToTime(From, time.Time{})
		if err != nil {
			log.Warn().Msg("Error parsing --from timedelta input, supported units are 's', 'm', 'h', 'd'. Ignoring --from")
		}
	}
	if To != "" {
		toTime, err = util.HumanTimeOrDeltaStringToTime(To, time.Time{})
		if err != nil {
			log.Warn().Msg("Error parsing --to timedelta input, supported units are 's', 'm', 'h'. Ignoring --to")
		}
	}
	if ErrorsShowHistogram {
		log.Debug().Msgf("Histogram bucket duration: %s", ErrorsHistogramBucketIntervalStr)
		var err error
		PeakBucketDuration, err = time.ParseDuration(ErrorsHistogramBucketIntervalStr)
		if err != nil {
			log.Fatal().Err(err).Msgf("Invalid histogram --bucket input: %s", ErrorsHistogramBucketIntervalStr)
		}
	}
	return WeaselConfig{
		FromTime:                fromTime,
		ToTime:                  toTime,
		LogLineRegex:            logLineRegex,
		MinErrLvl:               minErrLvl,
		MinErrLvlNum:            pglog.SeverityToNum(minErrLvl),
		MinSlowDurationMs:       MinSlowDurationMs,
		SystemOnly:              SystemOnly,
		ForceCsvInput:           Csv,
		Oneline:                 Oneline,
		PeaksOnly:               Peaks,
		LocksOnly:               LocksOnly,
		StatsOnly:               Stats,
		SlowTopNOnly:            SlowTopNOnly,
		SlowTopN:                SlowTopN,
		ErrorsHistogram:         ErrorsShowHistogram,
		HistogramBucketDuration: PeakBucketDuration,
	}
}
