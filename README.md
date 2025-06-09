# pgweasel

A simple CLI / interactive use oriented PostgreSQL log parser, to complement [pgBadger](https://github.com/darold/pgbadger).

pgweasel tries to:

* be way faster than pgBadger (~10x)
* way simpler, with less flags and operation modes - concentrating on the Pareto DBA flow
* focus on CLI interactions - no html / json
* more cloud-friendly - no deps, a single binary
* zero config - not dependent on Postgres `log_line_prefix`
* be more user-friendly - handle relative time inputs, auto-detect the latest log file location if not, subcommand short aliases


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


# Quickstart

```
git clone https://github.com/kmoppel/pgweasel.git

cd pgweasel && go build

./pgweasel -h

A simplistic PostgreSQL log parser for the console

Usage:
  pgweasel [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  errors      Shows WARNING and higher entries by default
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
      --regex string         Use a custom regex instead of: (default "(?s)^(?<syslog>[A-Za-z]{3} [0-9]{1,2} [0-9:]{6,} .*?: \\[[0-9\\-]+\\] )?(?P<log_time>[\\d\\-:\\. ]{19,23} [A-Z0-9\\-\\+]{2,5}|[0-9\\.]{14})[\\s:\\-].*?[\\s:\\-]?(?P<error_severity>[A-Z12345]{3,12}):\\s*(?P<message>(?s:.*))$")
      --to string            Log entries up to $time
  -v, --verbose              More chat

Use "pgweasel [command] --help" for more information about a command.
```


# Contributing

All kinds of feedback and help would be much appreciated - especially as I'm not a developer per se. Hopefully pgweasel will grow into a community project with rock solid quality.


# TODO

* jsonlog support
* goreleaser support
* more test & refactoring
* temp files mode
* perf optimizations, no effort so far
