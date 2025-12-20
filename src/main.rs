//! # pgweasel
//!
//! A simple CLI usage oriented PostgreSQL log parser, to complement pgBadger.
//!
//! pgweasel tries to:
//!  - be an order of magnitude faster than pgBadger
//!  - way simpler, with less flags, operating rather via commands and sub-commands
//!  - focus on CLI interactions only - no html / json
//!  - more cloud-friendly - no deps, a single binary
//!  - zero config - not dependent on Postgres log_line_prefix
//!  - be more user-friendly - handle relative time inputs, auto-detect log files, subcommand aliases
//!
//! # Features
//!
//!  - errors
//!    - [x] list
//!    - [ ] top
//!  - [ ] locks
//!  - [ ] peaks
//!  - [ ] slow
//!  - [ ] stats
//!  - [ ] system
//!  - [ ] connections

use log::error;

use crate::{convert_args::ConvertedArgs, print_logs::print_logs, severity::Severity};

mod cli;
mod convert_args;
mod error;
mod parsers;
mod print_logs;
mod severity;
mod util;

pub use self::error::{Error, Result};

fn main() -> Result<()> {
    let cli = cli::cli();
    let matches = cli.clone().get_matches();

    let mut converted_args: ConvertedArgs = matches.clone().into();
    converted_args = converted_args.expand_dirs()?.expand_archives()?;

    match matches.subcommand() {
        Some(("errors", sub_matches)) => {
            let error_command = sub_matches.subcommand().unwrap_or(("list", sub_matches));
            match error_command {
                ("list", list_subcommand) => {
                    print_logs(
                        converted_args,
                        list_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error),
                    )?;
                }
                ("top", _) => {
                    println!("Analyzing for top errors");
                }
                (name, _) => {
                    error!("Unsupported subcommand `{name}`")
                }
            }
        }
        Some(("locks", _)) => {
            error!("Not implemented")
        }
        Some(("peaks", _)) => {
            error!("Not implemented")
        }
        Some(("slow", _)) => {
            error!("Not implemented")
        }
        Some(("stats", _)) => {
            error!("Not implemented")
        }
        _ => error!("command not found"),
    }

    Ok(())
}
