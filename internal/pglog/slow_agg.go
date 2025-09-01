package pglog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/kmoppel/pgweasel/internal/util"
)

type SlowLogDurEntry struct {
	Duration    float64
	DurationStr string
}

type SlowLogStatsAggregator struct {
	CommandTagDurations map[string][]SlowLogDurEntry
}

func NewSlowLogAggregator() *SlowLogStatsAggregator {
	return &SlowLogStatsAggregator{
		CommandTagDurations: make(map[string][]SlowLogDurEntry),
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

	if _, ok := sa.CommandTagDurations[cmdTag]; !ok {
		sa.CommandTagDurations[cmdTag] = []SlowLogDurEntry{}
	}

	entry := SlowLogDurEntry{
		Duration:    duration,
		DurationStr: durationStr,
	}
	// log.Debug().Msgf("Adding duration entry: %+v for command tag: %s", entry, cmdTag)
	sa.CommandTagDurations[cmdTag] = append(sa.CommandTagDurations[cmdTag], entry)
}

func (sa *SlowLogStatsAggregator) ShowStats() {
	var totalDuration float64
	var totalCount int
	var allDurations []float64

	// First calculate stats per statement type
	for stmtTag, durations := range sa.CommandTagDurations {
		var tagTotal float64
		var tagDurations []float64

		for _, entry := range durations {
			tagTotal += entry.Duration
			tagDurations = append(tagDurations, entry.Duration)
			allDurations = append(allDurations, entry.Duration)
		}

		// Sort durations for percentile calculation
		sort.Float64s(tagDurations)

		p25 := util.CalculatePercentile(tagDurations, 25)
		p50 := util.CalculatePercentile(tagDurations, 50)
		p75 := util.CalculatePercentile(tagDurations, 75)
		p95 := util.CalculatePercentile(tagDurations, 95)

		fmt.Printf("Command Tag %s:\t\tp25: %.2f, p50: %.2f, p75: %.2f, p95: %.2f, SAMPLES: %d\n",
			stmtTag, p25, p50, p75, p95, len(durations))

		totalDuration += tagTotal
		totalCount += len(durations)
	}

	// Print overall statistics
	if totalCount > 0 {
		// Sort all durations for overall percentiles
		sort.Float64s(allDurations)
		overallP25 := util.CalculatePercentile(allDurations, 25)
		overallP50 := util.CalculatePercentile(allDurations, 50)
		overallP75 := util.CalculatePercentile(allDurations, 75)
		overallP95 := util.CalculatePercentile(allDurations, 95)

		fmt.Printf("TOTAL (ms):\t\t\tmin: %.2f, p25: %.2f, p50: %.2f, p75: %.2f, p95: %.2f, max: %.2f, SAMPLES: %d\n",
			allDurations[0], overallP25, overallP50, overallP75, overallP95, allDurations[len(allDurations)-1], totalCount)
	} else {
		fmt.Println("No statement statistics available")
	}
}
