package cmd

import (
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var GrepString string
var GrepRegex *regexp.Regexp

// statsCmd represents the stats command
var grepCmd = &cobra.Command{
	Use:   "grep $REGEX [$LOG_FILE_OR_FOLDER]",
	Short: "Show matching log entries only, e.g.: pgweasel grep 'Seq Scan.*tblX' mylogfile.log",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		GrepString = args[0]
		GrepRegex, err = regexp.Compile(GrepString)
		if err != nil {
			log.Fatal().Msg("Error parsing --grep regex input")
		}

		args = args[1:]
		MinErrLvl = "DEBUG5" // Override default WARNING+ output level
		showErrors(cmd, args)
	},
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"search", "regex", "find"},
}

func init() {
	rootCmd.AddCommand(grepCmd)
}
