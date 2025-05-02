package logparser

import (
	"log"

	"github.com/spf13/cobra"
)

func ParseLogFile(filePath string, cmd *cobra.Command) error {
	minLvl, _ := cmd.Flags().GetString("min-lvl")
	log.Println("Showing all msgs with minLvl >=", minLvl)
	return nil
}
