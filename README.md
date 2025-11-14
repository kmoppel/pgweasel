# Currently implemented operating modes / subcommands

## errors [ err | errs ]

`pgweasel errors $LOG(s)` - Show WARN+ (by default) log entries "as is"

`pgweasel errors --begin 10m $LOG(s)` - Show entries from last 10min

`pgweasel errors -l error $LOG(s)` - Show ERROR+ entries

`pgweasel errors -l LOG -t '2025-05-08 12:29' $LOG(s)` - Show LOG entries with timestamp begining with '2025-05-08 12:29'

`pgweasel errors -v $LOG(s)` - Show LOG entries with debug info including execution time.
