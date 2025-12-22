# Currently implemented operating modes / subcommands

## errors [ err | errs ]

`pgweasel errors $LOG(s)` - Show WARN+ (by default) log entries "as is"

`pgweasel errors --begin 10m $LOG(s)` - Show entries from last 10min

`pgweasel errors -l ERROR $LOG(s)` - Show ERROR+ entries

`pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog_pg14.csv` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

`pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog1.csv.gz` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

`pgweasel errors --debug $LOG(s)` - Show LOG entries with debug info including execution time.

`pgweasel slow 1s ./testdata/csvlog_pg14.csv` - Show LOG entries that took longer than 1second.
