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
var ErrorsShowHistogram bool
var ErrorsHistogramBucketIntervalStr string

var errorsTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Top N errors",
	Run: func(cmd *cobra.Command, args []string) {
		TopNErrorsOnly = true
		showErrors(cmd, args)
	},
}

const MAX_HISTO_WIDTH_CHARS = 80 // Maximum width of histogram output in characters in addition to timestamp + count

func init() {
	errorsCmd.AddCommand(errorsTopCmd)
	errorsCmd.Flags().BoolVarP(&ErrorsShowHistogram, "histo", "", false, "Show error counts histogram")
	errorsCmd.Flags().StringVarP(&ErrorsHistogramBucketIntervalStr, "bucket", "b", "1h", "Bucket size for histogram, e.g. 10min, 1h, 1d")

	errorsTopCmd.Flags().IntVarP(&TopNErrors, "top", "", 10, "Nr of top errors to show")

	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().StringVarP(&MinErrLvl, "min-level", "l", "WARNING", "The minimum Postgres error level to show")
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

func showErrors(cmd *cobra.Command, args []string) {
	var logFiles []string
	var topErrors *TopErrorsCollector = &TopErrorsCollector{}
	topErrors.Initialize()

	cfg := PreProcessArgs(cmd, args)

	log.Debug().Msgf("Running in debug mode. MinErrLvl=%s, MinSlowDurationMs=%d, SlowTopNOnly=%t, SlowTopN=%d, SlowStatsOnly=%t, From=%s, To=%s, SystemOnly=%v, SystemIncludeCheckpointer=%v, TopNErrorsOnly=%t, PeaksOnly=%t, StatsOnly=%t, GrepString=%s",
		cfg.MinErrLvl, cfg.MinSlowDurationMs, cfg.SlowTopNOnly, cfg.SlowTopN, cfg.SlowStatsOnly, cfg.FromTime, cfg.ToTime, cfg.SystemOnly, cfg.SystemIncludeCheckpointer, TopNErrorsOnly, cfg.PeaksOnly, cfg.StatsOnly, cfg.GrepString)

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

	slowTopNCollector := pglog.NewTopN(cfg.SlowTopN)

	slowStmtStatsCollector := pglog.NewSlowLogAggregator()

	histoBuckets := pglog.HistogramBucket{}
	if cfg.ErrorsHistogram {
		histoBuckets.Init(cfg.HistogramBucketDuration)
	}

	connectionsAggregator := pglog.ConnsAggregator{}
	if cfg.ConnectionsSummary {
		connectionsAggregator.Init()
	}

	extraErrContextTimestamp := ""

	for _, logFile := range logFiles {
		log.Debug().Msgf("Processing log file: %s", logFile)
		if !cfg.ToTime.IsZero() && logFile != "stdin" {
			// Peek at the first record and skip file if past the --to time
			firstRecord, err := logparser.PeekRecordFromFile(logFile, cfg.LogLineRegex, cfg.ForceCsvInput)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to peek at first record in %s", logFile)
			} else if firstRecord != nil {
				if firstRecord.GetTime().After(cfg.ToTime) {
					log.Warn().Msgf("Skipping file %s as first line is past the --to time", logFile)
					continue
				}
			}
		}

		for batch := range logparser.GetLogRecordsBatchFromFile(logFile, cfg.LogLineRegex, cfg.ForceCsvInput) {
			for _, rec := range batch {
				if rec.ErrorSeverity == "" {
					log.Error().Msgf("Got invalid entry: %+v", rec)
					continue
				}
				log.Debug().Msgf("Processing log entry: %+v", rec)

				if cfg.ConnectionsSummary {
					connectionsAggregator.AddEvent(rec)
					continue
				}

				if cfg.ErrorsHistogram {
					histoBuckets.Add(rec, cfg.HistogramBucketDuration, cfg.MinErrLvlNum)
					continue
				}

				if cfg.SlowStatsOnly {
					if rec.ErrorSeverity != "LOG" {
						continue
					}
					slowStmtStatsCollector.Add(rec)
					continue
				}

				if cfg.SlowTopNOnly {
					if rec.ErrorSeverity != "LOG" {
						continue
					}
					duration, _ := util.ExtractDurationMillisFromLogMessage(rec.Message)
					if duration == 0.0 {
						continue
					}
					slowTopNCollector.Add(pglog.TopNSlowLogEntry{Rec: &rec, DurationMs: duration})
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
					if rec.SeverityNum() == -1 && rec.LogTime == extraErrContextTimestamp {
						// Output extra context continuation records (HINT / STATEMENT) for an ERROR
						OutputLogRecord(rec, w, cfg.Oneline)
						w.WriteByte('\n')
					} else if logparser.DoesLogRecordSatisfyUserFilters(rec, cfg.MinErrLvlNum, Filters, cfg.FromTime, cfg.ToTime, cfg.MinSlowDurationMs, cfg.SystemOnly, cfg.SystemIncludeCheckpointer, cfg.GrepRegex) { // TODO pass cfg
						OutputLogRecord(rec, w, cfg.Oneline)
						if rec.ErrorSeverity == "ERROR" {
							extraErrContextTimestamp = rec.LogTime
						}
						w.WriteByte('\n')
					}
					if Verbose {
						w.Flush()
					}
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

	if cfg.SlowStatsOnly {
		slowStmtStatsCollector.ShowStats()
	}

	if cfg.ConnectionsSummary {
		connectionsAggregator.ShowStats()
	}

	if cfg.SlowTopNOnly {
		for _, slowTopNEntry := range slowTopNCollector.Values() {
			message := slowTopNEntry.Rec.LogTime + " " + slowTopNEntry.Rec.Message
			if cfg.Oneline {
				messageRunes := []rune(message)
				if len(messageRunes) > 200 {
					message = string(messageRunes[:197]) + "..."
				}
				w.WriteString(strings.ReplaceAll(message, "\n", " ") + "\n")
			} else {
				w.WriteString(message + "\n\n")
			}
		}
	}

	if cfg.ErrorsHistogram {
		OutputHistogramAsVertical(histoBuckets.GetSortedBuckets(), w, MAX_HISTO_WIDTH_CHARS)
		w.WriteString("\nTotal errors: " + strconv.Itoa(histoBuckets.TotalEvents) + "\n")
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

func OutputHistogramAsVertical(buckets []pglog.TimeBucket, w *bufio.Writer, maxWidth int) {
	// Find maximum count to normalize
	maxCount := 0
	for _, hb := range buckets {
		if hb.Count > maxCount {
			maxCount = hb.Count
		}
	}

	if maxCount == 0 {
		return // No data to show
	}

	// Output histogram bars
	for _, hb := range buckets {
		// Calculate normalized width
		normalizedWidth := 1
		if maxCount > 1 {
			normalizedWidth = int(float64(hb.Count) / float64(maxCount) * float64(maxWidth))
			if normalizedWidth < 1 {
				normalizedWidth = 1 // Ensure at least one asterisk
			}
		}

		// Format: timestamp: count [asterisks representing count]
		bar := strings.Repeat("*", normalizedWidth)
		w.WriteString(fmt.Sprintf("%-20s: %-5d %s\n", hb.Time, hb.Count, bar))
	}
}
