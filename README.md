# pgweasel

A simple CLI / interactive use oriented PostgreSQL log parser, to complement [pgBadger](https://github.com/darold/pgbadger).

pgweasel tries to:

* be way faster than pgBadger (~10x)
* way simpler, with less flags and operation modes - concentrating on the Pareto DBA flow
* focus on CLI interactions - no html / json
* more cloud-friendly - no deps, a single binary
* zero config - not dependent on Postgres `log_line_prefix`
* be more user-friendly - handle relative time inputs, auto-detect the latest log file location if not specified, subcommand aliases

## Project status

BETA. Command / subcommands "API" might change.

# Operating modes

`pgweasel errors $LOGFILE(S)_OR_FOLDER` - Show WARN+ log entries "as is"

`pgweasel errors --from 10m $LOG` - Show WARN+ log entries from last 10min

`pgweasel errors top $LOG` - Show the most frequent error messages with counts

`pgweasel errors --hist $LOG` - Show a basic vertical histogram of error counts. Default --bucket=1h

`pgweasel errors --hist -l debug5 $LOG` - Show a histogram for all events, not only errors

`pgweasel locks $LOG` - Only show locking (incl. deadlocks, recovery conflicts) entries

`pgweasel peaks $LOG` - Show the "busiest" time periods with most log events, using a 10min bucket by default

`pgweasel slow 1s $LOG` - Show queries taking longer than give threshold

`pgweasel slow top $LOG` - Show top 10 (by default) slowest queries

`pgweasel stats $LOG` - Summary of log events - counts / frequency of errors, connections, checkpoints, autovacuums

`pgweasel system $LOG` - Show instance lifecycle events only, i.e. Postgres internal processes, replication, extensions

`pgweasel grep $REGEX $LOG` - Show full log entries of matching search fragments

`pgweasel connections $LOG` - Show connections counts by total, db, user, application name. Assumes log_connections enabled


# Quickstart

```
git clone https://github.com/kmoppel/pgweasel.git

cd pgweasel && go build -ldflags "-s -w"

./pgweasel -h

A simplistic PostgreSQL log parser for the console

Usage:
  pgweasel [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  connections Show connections summary
  errors      Shows WARNING and higher entries by default
  grep        Show matching log entries only, e.g.: pgweasel grep 'Seq Scan.*tblX' mylogfile.log
  help        Help about any command
  locks       Only show locking related entries
  peaks       Identify periods where most log entries are emitted, per severity level
  slow        Show queries above user set threshold, e.g.: pgweasel slow 1s mylogfile.log
  stats       Summary of log events
  system      Show messages by Postgres internal processes

Flags:
      --csv                  Specify that input file or stdin is actually CSV regardless of file extension
  -f, --filter stringArray   Add extra line match conditions (regex)
      --from string          Log entries from $time, e.g.: -1h
  -h, --help                 help for pgweasel
  -1, --oneline              Compact multiline entries
      --peaks                Show only event counts per log level for peak load periods
      --to string            Log entries up to $time
  -v, --verbose              More chat

Use "pgweasel [command] --help" for more information about a command.
```


# Contributing

All kinds of feedback and help (PR-s, co-maintainer) would be much appreciated - especially as I'm not a developer per se. Hopefully pgweasel will grow into a community project with rock solid quality.

## Have sample log files ?

I've scraped the Postgres mailing archives for *.log attachements (in [testdata](https://github.com/kmoppel/pgweasel/tree/main/testdata) folder), but they are not much sadly...so if you have some real-life logs from busy or somehow "troublesome" instances, not containing secrets - please add one one via PR or proide some S3 etc link under issues. Thank you!

# Perf difference indication

Let's say the goal here is the very common task of finding out the slowest query runtimes from logs (330M uncompressed, not large by any means)...which firstly is very unconvenient with pgbadger as it does a lot of things and doesn't have feature flags for all of the common things. One can speed things up though with disabling some features / summaries.

## pgbadger

```
time pgbadger testdata/pgbench_large.log.gz -j 2 --disable-error --disable-hourly --disable-type --disable-session --disable-connection --disable-lock --disable-temporary --disable-checkpoint --disable-autovacuum --no-progressbar
...
real	2m4.786s
user	2m2.201s
sys	  0m0.656s
```

Note the added `--jobs=2` as pgweasel by default uses an additional thread as well

## pgweasel

```
time ./pgweasel slow top testdata/pgbench_large.log.gz
...
real	0m9.168s
user	0m11.517s
sys	  0m0.683s
```
