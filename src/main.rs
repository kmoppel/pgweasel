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

use std::time::Duration;

use humantime::parse_duration;
use log::error;

use crate::{
    aggregators::{Aggregator, TopSlowQueryAggregator},
    convert_args::ConvertedArgs,
    filters::{Filter, FilterSlow},
    output_results::output_results,
    severity::Severity,
};

mod aggregators;
mod cli;
mod convert_args;
mod duration;
mod error;
mod filters;
mod output_results;
mod severity;
mod util;

pub use self::error::{Error, Result};

fn main() -> Result<()> {
    let cli = cli::cli();
    let matches = cli.clone().get_matches();

    let mut converted_args: ConvertedArgs = ConvertedArgs::parse_from_matches(matches.clone())?;
    converted_args = converted_args.expand_dirs()?.expand_archives()?;

    let mut aggregators: Vec<Box<dyn Aggregator>> = Vec::new();
    let mut filters: Vec<Box<dyn Filter>> = Vec::new();

    match matches.subcommand() {
        Some(("errors", sub_matches)) => {
            let error_command = sub_matches.subcommand().unwrap_or(("list", sub_matches));
            match error_command {
                ("list", list_subcommand) => {
                    output_results(
                        converted_args,
                        list_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error),
                        &mut aggregators,
                        &filters,
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
        Some(("slow", matches)) => {
            let mut treshold = Duration::from_secs(3);
            if let Some(treshold_str) = matches.get_one::<String>("treshold") {
                treshold = parse_duration(&treshold_str)?;
            };
            filters.push(Box::new(FilterSlow::new(treshold)));
            output_results(
                converted_args,
                &Severity::Log,
                &mut aggregators,
                &filters,
            )?;
        }
        Some(("stats", _)) => {
            error!("Not implemented")
        }
        _ => error!("command not found"),
    }

    Ok(())
}
