/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kmoppel/pgweasel/internal/logparser"
	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/kmoppel/pgweasel/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var MinErrLvl = "WARNING"
var errorsCmd = &cobra.Command{
	Use:   "errors [$LOG_FILE_OR_FOLDER]",
	Short: "Shows WARNING and higher entries by default",
	Run: func(cmd *cobra.Command, args []string) {
		showErrors(cmd, args)
	},
	Aliases: []string{"err", "errs", "error"},
}
var TopNErrors int
var TopNErrorsOnly bool

var errorsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Top N errors",
	Run: func(cmd *cobra.Command, args []string) {
		TopNErrorsOnly = true
		showErrors(cmd, args)
	},
}

type ErrorLevelMessage struct {
	ErrorLevel string
	Message    string
}

type TopErrorsCollector struct {
	ErrorCounts map[string]map[string]int // [ERROR][MESSAGE] = COUNT
	TotalCount  int
}

func (tec *TopErrorsCollector) Initialize() {
	tec.ErrorCounts = make(map[string]map[string]int)
	for _, level := range pglog.ERROR_SEVERITIES {
		tec.ErrorCounts[level] = make(map[string]int)
	}
}

func (tec *TopErrorsCollector) AddError(level string, message string) {
	tec.ErrorCounts[level][message]++
}

type TopError struct {
	Level   string
	Message string
	Count   int
}

func (tec *TopErrorsCollector) GetTopN(topN int) []TopError {
	// flatten before sorting
	topErrors := make([]TopError, 0)
	for level, messages := range tec.ErrorCounts {
		for message, count := range messages {
			topErrors = append(topErrors, TopError{
				Level:   level,
				Message: message,
				Count:   count,
			})
		}
	}
	// sort by count
	sort.Slice(topErrors, func(i, j int) bool {
		return topErrors[i].Count > topErrors[j].Count
	})
	if len(topErrors) <= topN {
		return topErrors
	}
	return topErrors[:topN]
}

func init() {
	errorsCmd.AddCommand(errorsTopCmd)
	errorsTopCmd.Flags().IntVarP(&TopNErrors, "top", "", 10, "Nr of top errors to show")

	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
}

func showErrors(cmd *cobra.Command, args []string) {
	var logFiles []string
	var topErrors *TopErrorsCollector = &TopErrorsCollector{}
	topErrors.Initialize()

	cfg := PreProcessArgs(cmd, args)

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, SlowTopNOnly=%t, SlowTopN=%d, From=%s, To=%s, SystemOnly=%v, TopNErrorsOnly=%t, PeaksOnly=%t, StatsOnly=%t",
		cfg.MinErrLvl, cfg.MinSlowDurationMs, cfg.SlowTopNOnly, cfg.SlowTopN, cfg.FromTime, cfg.ToTime, cfg.SystemOnly, TopNErrorsOnly, cfg.PeaksOnly, cfg.StatsOnly)

	if len(args) == 0 && util.CheckStdinAvailable() {
		logFiles = []string{"stdin"}
	} else {
		logFiles = util.GetLogFilesFromUserArgs(args)
		if len(logFiles) == 0 {
			log.Error().Msg("No log files found to process")
			return
		}
	}

	// Create a buffered writer for better performance
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	minErrLvlSeverityNum := pglog.SeverityToNum(MinErrLvl)

	peaksBucket := pglog.EventBucket{}
	if cfg.PeaksOnly {
		peaksBucket.Init()
		log.Debug().Msgf("In peaks mode with bucket interval: %v", PeakBucketDuration)
	}

	statsAggregator := pglog.StatsAggregator{}
	if cfg.StatsOnly {
		statsAggregator.Init()
	}

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)

		for rec := range logparser.GetLogRecordsFromFile(logFile, cfg.LogLineRegex, cfg.ForceCsvInput) {
			if rec.ErrorSeverity == "" {
				log.Error().Msgf("Got invalid entry: %+v", rec)
				continue
			}
			log.Debug().Msgf("Processing log entry: %+v", rec)

			if cfg.SlowTopNOnly {
				if rec.ErrorSeverity != "LOG" {
					continue
				}
				duration := util.ExtractDurationMillisFromLogMessage(rec.Message)
				if duration == 0.0 {
					continue
				}
				continue
			}

			if cfg.StatsOnly {
				statsAggregator.AddEvent(rec)
				continue
			}

			if cfg.LocksOnly {
				if rec.IsLockingRelatedEntry() {
					OutputLogRecord(rec, w, cfg.Oneline)
					w.WriteByte('\n')
					if Verbose {
						w.Flush()
					}
				}
				continue
			}

			if cfg.PeaksOnly {
				peaksBucket.AddEvent(rec, PeakBucketDuration)
				continue
			}

			if TopNErrorsOnly && rec.SeverityNum() >= minErrLvlSeverityNum {
				topErrors.AddError(rec.ErrorSeverity, rec.Message)
			} else {
				if logparser.DoesLogRecordSatisfyUserFilters(rec, cfg.MinErrLvlNum, Filters, cfg.FromTime, cfg.ToTime, cfg.MinSlowDurationMs, cfg.SystemOnly) { // TODO pass cfg
					OutputLogRecord(rec, w, cfg.Oneline)
					w.WriteByte('\n')
				}
				if Verbose {
					w.Flush()
				}
			}
		}
		log.Debug().Msgf("Finished processing log file: %s", logFile)
	}

	if TopNErrorsOnly {
		for _, topError := range topErrors.GetTopN(TopNErrors) {
			w.WriteString(strconv.Itoa(topError.Count) + " " + topError.Level + ": " + topError.Message + "\n")
		}
	}

	if cfg.PeaksOnly {
		w.WriteString("Most events per " + PeakBucketIntervalStr + ":\n\n")
		for lvl, bucketWithCount := range peaksBucket.GetTopBucketsBySeverity() {
			for timeBucket, count := range bucketWithCount {
				w.WriteString(fmt.Sprintf("%-12s: %-6d (%s, e.g.: %s)\n", lvl, count, timeBucket, peaksBucket.GetFirstRealTimeStringForBucket(timeBucket)))
			}
		}

		topLockingTime, lockCount, realTimeString := peaksBucket.GetTopLockingPeriod()
		w.WriteString(fmt.Sprintf("\n%-12s: %-6d (%s, e.g.: %s)\n", "LOCKS", lockCount, topLockingTime, realTimeString))

		topConnsTime, connsCount, realTimeString := peaksBucket.GetTopConnectPeriod()
		w.WriteString(fmt.Sprintf("\n%-12s: %-6d (%s, e.g: %s)\n", "CONNECTS", connsCount, topConnsTime, realTimeString))
	}

	if cfg.StatsOnly {
		statsAggregator.ShowStats()
	}

	w.Flush()
}

func OutputLogRecord(rec pglog.LogEntry, w *bufio.Writer, oneline bool) {
	if rec.CsvColumns != nil {
		if oneline {
			w.WriteString(strings.ReplaceAll(rec.CsvColumns.String(), "\n", " "))
		} else {
			w.WriteString(rec.CsvColumns.String())
		}
	} else {
		if oneline {
			w.WriteString(strings.Join(rec.Lines, " "))
		} else {
			w.WriteString(strings.Join(rec.Lines, "\n"))
		}
	}
}
