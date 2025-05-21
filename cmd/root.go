package cmd

import (
	"os"

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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More chat")
	rootCmd.PersistentFlags().StringVarP(&From, "from", "", "", "Log entries from $time, e.g.: -1h")
	rootCmd.PersistentFlags().StringVarP(&To, "to", "", "", "Log entries up to $time")
	// rootCmd.PersistentFlags().BoolVarP(&Oneline, "oneline", "1", false, "Compact multiline entries")
	rootCmd.PersistentFlags().StringArrayVarP(&Filters, "filter", "f", nil, "Add extra line match conditions (regex)")
	rootCmd.PersistentFlags().StringVarP(&Connstr, "connstr", "", "", "Connect to specified instance and determine log location / settings")
	// rootCmd.PersistentFlags().BoolVarP(&Tail, "tail", "t", false, "Keep watching the log file for new entries")
	rootCmd.PersistentFlags().StringVarP(&LogLineRegex, "regex", "", `(?s)^(?P<log_time>[\d\-:\. ]{19,23} [A-Z]{2,5})[\s:\-].*[\s:\-](?P<error_severity>[A-Z12345]+):\s*(?P<message>(?s:.*))$`, "Use a custom regex instead of:")
}
