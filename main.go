package main

import (
	"log"
	"os"

	"github.com/kmoppel/pgweasel/cmd"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
)

func main() {
	log.SetFlags(0) // Disable default output of: 2025/05/08 12:34:56 hello
	log.SetOutput(os.Stdout)
	zl.Logger = zl.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd.Execute()
}
