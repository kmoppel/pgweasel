package cmd

import (
	"errors"
	"log"
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
	// Run: func(cmd *cobra.Command, args []string) { },
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one arg")
		}
		log.Println("args", args)
		// TODO look for logfiles ...
		return nil
	},
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
var Prefix string
var Users string
var NoUsers string
var From string
var To string
var Oneline bool
var Matches string
var Connstr string
var Tail bool
var Db string
var NoDb string
var GlobPath string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More chat")
	rootCmd.PersistentFlags().StringVarP(&Prefix, "prefix", "p", "", "Postgres log_line_prefix")
	rootCmd.PersistentFlags().StringVarP(&Users, "users", "u", "", "Only look at entries by certain users (regex)")
	rootCmd.PersistentFlags().StringVarP(&NoUsers, "no-users", "", "", "Ignore log lines from certain users (regex)")
	rootCmd.PersistentFlags().StringVarP(&From, "from", "", "", "Log entries from $time")
	rootCmd.PersistentFlags().StringVarP(&To, "to", "", "", "Log entries up to $time")
	rootCmd.PersistentFlags().BoolVarP(&Oneline, "oneline", "", false, "Compact multiline entries")
	rootCmd.PersistentFlags().StringVarP(&Matches, "matches", "", "", "Only look at entries matching certain patterns (regex)")
	rootCmd.PersistentFlags().StringVarP(&Connstr, "connstr", "", "", "Connect to specified instance and determine log location / settings")
	rootCmd.PersistentFlags().BoolVarP(&Tail, "tail", "t", false, "Keep watching the log file for new entries")
	rootCmd.PersistentFlags().StringVarP(&Db, "db", "", "", "Only look at certain databases (regex)")
	rootCmd.PersistentFlags().StringVarP(&NoDb, "no-db", "", "", "Ignore certain databases (regex)")
	rootCmd.PersistentFlags().StringVarP(&GlobPath, "glob", "", "", "Glob pattern to look for log file locations")
}
