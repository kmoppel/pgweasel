package main

import (
	"log"
	"os"

	"github.com/kmoppel/pgweasel/cmd"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
)

func main() {
	log.SetFlags(0)
	zl.Logger = zl.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd.Execute()
}
