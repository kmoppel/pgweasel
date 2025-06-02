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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More chat")
	rootCmd.PersistentFlags().StringVarP(&From, "from", "", "", "Log entries from $time, e.g.: -1h")
	rootCmd.PersistentFlags().StringVarP(&To, "to", "", "", "Log entries up to $time")
	rootCmd.PersistentFlags().BoolVarP(&Oneline, "oneline", "1", false, "Compact multiline entries")
	rootCmd.PersistentFlags().StringArrayVarP(&Filters, "filter", "f", nil, "Add extra line match conditions (regex)")
	// rootCmd.PersistentFlags().BoolVarP(&Tail, "tail", "t", false, "Keep watching the log file for new entries")
	rootCmd.PersistentFlags().StringVarP(&LogLineRegex, "regex", "", logparser.DEFAULT_REGEX_STR, "Use a custom regex instead of:")
	rootCmd.PersistentFlags().BoolVarP(&Csv, "csv", "", false, "Specify that input file or stdin is actually CSV regardless of file extension")
}

type WeaselConfig struct {
	FromTime          time.Time
	ToTime            time.Time
	LogLineRegex      *regexp.Regexp
	MinErrLvl         string
	MinErrLvlNum      int
	SystemOnly        bool
	MinSlowDurationMs int
	ForceCsvInput     bool
	Oneline           bool
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
			log.Warn().Msg("Error parsing --from timedelta input, supported units are 's', 'm', 'h'. Ignoring --from")
		}
	}
	if To != "" {
		toTime, err = util.HumanTimeOrDeltaStringToTime(To, time.Time{})
		if err != nil {
			log.Warn().Msg("Error parsing --to timedelta input, supported units are 's', 'm', 'h'. Ignoring --to")
		}
	}
	return WeaselConfig{
		FromTime:          fromTime,
		ToTime:            toTime,
		LogLineRegex:      logLineRegex,
		MinErrLvl:         minErrLvl,
		MinErrLvlNum:      pglog.SeverityToNum(minErrLvl),
		MinSlowDurationMs: MinSlowDurationMs,
		SystemOnly:        SystemOnly,
		ForceCsvInput:     Csv,
		Oneline:           Oneline,
	}
}
