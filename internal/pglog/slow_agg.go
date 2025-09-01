package pglog

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/kmoppel/pgweasel/internal/util"
)

type SlowLogDurEntry struct {
	Duration    float64
	DurationStr string
}

type SlowLogStatsAggregator struct {
	StmtTagDurations map[string][]SlowLogDurEntry
}

func NewSlowLogAggregator() *SlowLogStatsAggregator {
	return &SlowLogStatsAggregator{
		StmtTagDurations: make(map[string][]SlowLogDurEntry),
	}
}

func (sa *SlowLogStatsAggregator) Add(r LogEntry) {
	if r.ErrorSeverity != "LOG" {
		return
	}
	if !strings.HasPrefix(r.Message, "duration: ") {
		return
	}

	cmdTag := r.GetCommandTag()
	if cmdTag == "" {
		log.Fatal().Msgf("Failed to extract command tag from: %s", r.Message)
	}

	duration, durationStr := util.ExtractDurationMillisFromLogMessage(r.Message)
	if duration == 0 {
		log.Fatal().Msgf("Got zero duration from: %s", r.Message)
	}

	if _, ok := sa.StmtTagDurations[cmdTag]; !ok {
		sa.StmtTagDurations[cmdTag] = []SlowLogDurEntry{}
	}

	entry := SlowLogDurEntry{
		Duration:    duration,
		DurationStr: durationStr,
	}
	// log.Debug().Msgf("Adding duration entry: %+v for command tag: %s", entry, cmdTag)
	sa.StmtTagDurations[cmdTag] = append(sa.StmtTagDurations[cmdTag], entry)
}

func (sa *SlowLogStatsAggregator) ShowStats() {
	var totalDuration float64
	var totalCount int

	// First calculate stats per statement type
	for stmtTag, durations := range sa.StmtTagDurations {
		var tagTotal float64
		for _, entry := range durations {
			tagTotal += entry.Duration
		}
		tagAvg := tagTotal / float64(len(durations))
		fmt.Printf("Command Tag: %s, avg_ms: %.2f, count: %d\n", stmtTag, tagAvg, len(durations))

		totalDuration += tagTotal
		totalCount += len(durations)
	}

	// Print overall mean
	if totalCount > 0 {
		overallMean := totalDuration / float64(totalCount)
		fmt.Printf("TOTAL mean: %.2f ms, total statements: %d\n", overallMean, totalCount)
	} else {
		fmt.Println("No statement statistics available")
	}
}
