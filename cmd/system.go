package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show messages by Postgres internal processes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("system called")
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// systemCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// systemCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
