# Currently implemented operating modes / subcommands

## errors [ err | errs ]

`pgweasel errors $LOG(s)` - Show WARN+ (by default) log entries "as is"

`pgweasel errors --begin 10m $LOG(s)` - Show entries from last 10min

`pgweasel errors -l error $LOG(s)` - Show ERROR+ entries
