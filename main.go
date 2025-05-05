package main

import (
	"log"

	"github.com/kmoppel/pgweasel/cmd"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
