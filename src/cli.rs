use std::path::PathBuf;

use clap::{Arg, ArgAction, Command, arg, value_parser};

use crate::Severity;

pub fn cli() -> Command {
    Command::new("pgweasel")
        .about("A PostgreSQL log parser")
        .version("0.1")
        .arg(arg!(--debug <DEBUG>).short('d').help("Verbose. Show debug information").action(ArgAction::SetTrue))
        .arg(arg!(--mask <MASK>).short('m').help("Postgres log timestamp mask (e.g. \"2025-05-21 12:57\" - will show all events at 12:57)"))
        .arg(arg!(--begin <BEGIN>).short('b'))
        .arg(arg!(--end <END>).short('e'))
        .subcommand_required(true)
        .subcommand(
            Command::new("errors")
                .about("Show or summarize error messages")
                .alias("error")
                .alias("err")
                .args_conflicts_with_subcommands(true)
                .args(level_args())
                .args(filelist_args())
                .subcommand(Command::new("list").args(level_args()).args(filelist_args()))
                .subcommand(Command::new("top").args(level_args()).args(filelist_args()))
                .subcommand(Command::new("hist")
                    .about("Show histogram of error occurrences over time")
                    .alias("histogram")
                    .args(level_args())
                    .arg(arg!(--bucket <INTERVAL>)
                        .short('b')
                        .help("Interval for histogram buckets, e.g. 10s, 1m, 1h. Defaults to 1h")
                        .value_parser(value_parser!(String))
                        .default_value("1h"))
                    .args(filelist_args()))
        )
        .subcommand(
            Command::new("locks")
                .alias("loc")
                .alias("lock")
                .alias("deadlock")
                .alias("deadlocks")
                .about("Only show locking (incl. deadlocks, recovery conflicts) entries")
                .args(filelist_args())
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("peaks")
                .about("Show the \"busiest\" time periods with most log events")
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("slow")
                .subcommand(Command::new("top").args(filelist_args()))
                .args_conflicts_with_subcommands(true)
                .about("Show queries taking longer than give threshold")
                .arg(arg!(<TRESHOLD>).help("Treshold in format like 10s, 10ms to consider slow query."))
                .args(filelist_args())
        )
        .subcommand(
            Command::new("system")
                .args_conflicts_with_subcommands(true)
                .about("Show lifecycle / Postgres internal events, i.e. autovacuum, replication, extensions, config changes etc")
                .alias("sys")
                .alias("pg")
                .alias("postgre")
                .alias("postgres")
                .args(filelist_args())
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("connections")
                .args_conflicts_with_subcommands(true)
                .about("Show connections counts by total, db, user, application name. Assumes log_connections enabled")
                .alias("conns")
                .alias("conn")
                .args(filelist_args())
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("stats")
                .about("Summary of log events - counts / frequency of errors, connections, checkpoints, autovacuums")
                .args_conflicts_with_subcommands(true)
                .args(filelist_args())
        )
}

fn level_args() -> Vec<Arg> {
    vec![
        arg!(--level <SEVERITY>)
            .short('l')
            .value_parser(value_parser!(Severity)),
    ]
}

fn filelist_args() -> Vec<Arg> {
    vec![arg!(<PATH> ..."Log files to analyze").value_parser(clap::value_parser!(PathBuf))]
}
