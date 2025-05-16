package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgweasel",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run:  showErrors,
	Args: cobra.MaximumNArgs(1), // empty means outdetect or use hardcoded defaults
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More chat")
	rootCmd.PersistentFlags().StringVarP(&From, "from", "", "", "Log entries from $time")
	rootCmd.PersistentFlags().StringVarP(&To, "to", "", "", "Log entries up to $time")
	rootCmd.PersistentFlags().BoolVarP(&Oneline, "oneline", "1", false, "Compact multiline entries")
	rootCmd.PersistentFlags().StringArrayVarP(&Filters, "filter", "f", nil, "Add extra line match conditions (regex)")
	rootCmd.PersistentFlags().StringVarP(&Connstr, "connstr", "", "", "Connect to specified instance and determine log location / settings")
	rootCmd.PersistentFlags().BoolVarP(&Tail, "tail", "t", false, "Keep watching the log file for new entries")
}
