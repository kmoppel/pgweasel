# Rust rewrite of pgweasel

This is a fork of [pgweasel](https://github.com/kmoppel/pgweasel/) as rewrite in RUST. 

A simple CLI usage oriented PostgreSQL log parser, to complement pgBadger.

pgweasel tries to:

* be an order of magnitude faster than pgBadger
* way simpler, with less flags, operating rather via commands and sub-commands
* focus on CLI interactions only - no html / json
* more cloud-friendly - no deps, a single binary
* zero config - not dependent on Postgres log_line_prefix
* be more user-friendly - handle relative time inputs, auto-detect log files, subcommand aliases

## Table of Contents

1. [Status](#status)
2. [Installation](#installation)
3. [Usage](#usage)
4. [Contributing](#contributing)
5. [License](#license)

## Status

This project is in BETA. Command / subcommands "API" might change.

MAC & Linux targts passes tests. Windows - needs testers.

## Installation

### Download latest binaries from GitHub

You can download binaries from [GitHub Releases](https://github.com/gintsgints/pgweasel/releases/tag/latest).

### Install from source

Make sure, you have RUST compiler installed. Then:

```sh
git clone https://github.com/gintsgints/pgweasel.git
cd pgweasel
cargo build --release
The built binary will be in target/release/pgweasel.
```


## Usage

Here is a list of currently implemented commands

### Global options

These flags work with any subcommand:

| Flag | Short | Description |
|------|-------|-------------|
| `--begin` | `-b` | Show entries from a given time or relative interval (e.g. `10m`, `2025-05-21 12:00:00`) |
| `--end` | `-e` | Show entries up to a given time |
| `--mask` | `-m` | Timestamp prefix mask — show all events matching the prefix |
| `--context` | `-C` | Show NUM records before **and** after each matching record |
| `--before-context` | `-B` | Show NUM records before each matching record |
| `--after-context` | `-A` | Show NUM records after each matching record |

### errors [ err | errs ]

- [x] `pgweasel errors $LOG(s)` - Show WARN+ (by default) log entries "as is"

- [x] `pgweasel errors --begin 10m $LOG(s)` - Show entries from last 10min

- [x] `pgweasel errors -l error $LOG(s)` - Show ERROR+ entries

- [x] `pgweasel error top ./tests/files/debian_default2.log` - Show top LOG message count

- [x] `pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog_pg14.csv` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

- [x] `pgweasel -t "2025-05-21 13:00:00" errors -l LOG testdata/csvlog1.csv.gz` - Show LOG entries with timestamp begining with '2025-05-21 13:00:00'

- [x] `pgweasel errors --debug $LOG(s)` - Show LOG entries with debug info including execution time.

- [x] `pgweasel errors hist $LOG` - Show a basic vertical histogram of error counts. Default --bucket=1h

- [x] `pgweasel errors hist -b 3m -l debug5 $LOG` - Show a histogram for all events, not only errors, using --bucket=3m

### locks

- [x] `pgweasel locks ./tests/files/locking.log` Only show locking entries (incl. deadlocks, recovery conflicts)

### peaks

- [ ] `pgweasel peaks $LOG` Show the "busiest" time periods with most log events, using a 10min bucket by default

### slow

- [x] `pgweasel slow 1s ./testdata/csvlog_pg14.csv` - Show LOG entries that took longer than 1second.

- [x] `pgweasel slow top ./testdata/cloudsql.log` - Show top 10 slowest queries.

- [ ] `pgweasel slow stat $LOG` Show avg slow log exec times per query type

### stats

- [ ] `pgweasel stats $LOG` Summary of log events - counts / frequency of errors, connections, checkpoints, autovacuums

### system

- [x] `pgweasel system testdata/debian_default.log` Show lifecycle / Postgres internal events, i.e. autovacuum, replication, extensions, config changes etc

### connections

- [x] `pgweasel connections ./tests/files/azure_connections.log` Show connections counts by total, db, user, application name. Assumes log_connections enabled. Output is sorted by count descending.

### grep

- [x] `pgweasel grep "deadlock" $LOG(s)` Show log records containing the search term (case-insensitive)

- [x] `pgweasel -A 2 grep "deadlock" $LOG(s)` Show matching records plus 2 records after each match

- [x] `pgweasel -B 1 grep "FATAL" $LOG(s)` Show matching records plus 1 record before each match

- [x] `pgweasel -C 2 grep "checkpoint" $LOG(s)` Show matching records plus 2 records on both sides

## Contributing

All kinds of feedback and help (PR-s, co-maintainer) would be much appreciated. Hopefully pgweasel will grow into a community project with rock solid quality.

When creating MR, make sure `cargo test` and `cargo fmt --all -- --check` pass. 

Have sample log files ?
I've scraped the Postgres mailing archives for *.log attachements (in testdata folder), but they are not much sadly...so if you have some real-life logs from busy or somehow "troublesome" instances, not containing secrets - please add one one via PR or proide some S3 etc link under issues. Thank you!

## License

pgweasel is free software distributed under the [Apache Licence](./LICENSE).
