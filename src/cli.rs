use std::path::PathBuf;

use clap::{Arg, ArgAction, Command, arg};

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
            Command::new("error")
                .about("Show or summarize error messages")
                .args_conflicts_with_subcommands(true)
                .flatten_help(true)
                .args(error_args())
                .args(filelist_args())
                .subcommand(Command::new("list").args(error_args()).args(filelist_args()))
                .subcommand(Command::new("top").args(error_args()).args(filelist_args()))
        )
        .subcommand(
            Command::new("locks")
                .about("Only show locking (incl. deadlocks, recovery conflicts) entries")
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("peaks")
                .about("Show the \"busiest\" time periods with most log events")
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("slow")
                .about("Show queries taking longer than give threshold")
                .args_conflicts_with_subcommands(true)
        )
        .subcommand(
            Command::new("stats")
                .about("Summary of log events - counts / frequency of errors, connections, checkpoints, autovacuums")
                .args_conflicts_with_subcommands(true)
        )
}

fn error_args() -> Vec<Arg> {
    vec![arg!(--level <LEVEL>).short('l')]
}

fn filelist_args() -> Vec<Arg> {
    vec![arg!(<PATH> ..."Log files to analyze").value_parser(clap::value_parser!(PathBuf))]
}
