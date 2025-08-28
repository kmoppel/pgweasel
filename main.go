package main

import (
	"os"

	"github.com/kmoppel/pgweasel/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// // Start CPU profiling
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("could not create CPU profile")
	// }
	// defer f.Close()

	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatal().Err(err).Msg("could not start CPU profile")
	// }
	// defer pprof.StopCPUProfile()

	cmd.Execute()
}
