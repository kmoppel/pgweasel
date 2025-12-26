# Rust rewrite of pgweasel

This is a fork of [pgweasel](https://github.com/kmoppel/pgweasel/) as rewrite in RUST. 

## Currently implemented operating modes / subcommands

### errors [ err | errs ]

[*] `pgweasel errors $LOG(s)` - Show WARN+ (by default) log entries "as is"

[*] `pgweasel errors --begin 10m $LOG(s)` - Show entries from last 10min

[*] `pgweasel errors -l ERROR $LOG(s)` - Show ERROR+ entries

[*] `pgweasel error top top ./tests/files/debian_default2.log` - Show top LOG message count

[*] `pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog_pg14.csv` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

[*] `pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog1.csv.gz` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

[*] `pgweasel errors --debug $LOG(s)` - Show LOG entries with debug info including execution time.

[ ] `pgweasel errors --histo $LOG` - Show a basic vertical histogram of error counts. Default --bucket=1h

[ ] `pgweasel errors --histo -l debug5 $LOG` - Show a histogram for all events, not only errors

### locks

[*] `pgweasel locks ./tests/files/locking.log` Only show locking entries (incl. deadlocks, recovery conflicts)

### peaks

[ ] `pgweasel peaks $LOG` Show the "busiest" time periods with most log events, using a 10min bucket by default

### slow

[*] `pgweasel slow 1s ./testdata/csvlog_pg14.csv` - Show LOG entries that took longer than 1second.

[*] `pgweasel slow top ./testdata/cloudsql.log` - Show top 10 slowest queries.

[ ] `pgweasel slow stat $LOG` Show avg slow log exec times per query type

### stats

[ ] `pgweasel stats $LOG` Summary of log events - counts / frequency of errors, connections, checkpoints, autovacuums

### system

[ ] `pgweasel system $LOG` Show lifecycle / Postgres internal events, i.e. autovacuum, replication, extensions, config changes etc

### connections

[ ] `pgweasel connections $LOG` Show connections counts by total, db, user, application name. Assumes log_connections enabled

### grep

For grep I would recommend using grep cli - ripgrep

