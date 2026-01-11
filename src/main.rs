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
//!    - [x] top
//!    - [ ] histogram
//!  - [x] locks
//!  - [ ] peaks
//!  - [x] slow
//!   - [x] filter by threshold
//!   - [x] top slow queries
//!   - [ ] stat
//!  - [ ] stats
//!  - [x] system
//!  - [ ] connections

use std::time::Duration;

use humantime::parse_duration;
use log::{debug, error};

use crate::{
    aggregators::{
        Aggregator, ConnectionsAggregator, ErrorFrequencyAggregator, ErrorHistogramAggregator,
        TopSlowQueries,
    },
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
mod format;
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
                ("top", top_subcommand) => {
                    aggregators.push(Box::new(ErrorFrequencyAggregator::new()));
                    converted_args.print_details = false;
                    output_results(
                        converted_args,
                        top_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error),
                        &mut aggregators,
                        &filters,
                    )?;
                }
                ("hist", hist_subcommand) => {
                    // aggregators.push(Box::new(ErrorFrequencyAggregator::new()));
                    converted_args.print_details = false;
                    let mut interval = Duration::from_hours(1);
                    if let Some(interval_str) = hist_subcommand.get_one::<String>("bucket") {
                        interval = parse_duration(interval_str)?;
                    }
                    aggregators.push(Box::new(ErrorHistogramAggregator::new(interval)));
                    debug!(
                        "Histogram severity: {:?}",
                        hist_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error)
                    );
                    debug!("Histogram interval: {:?}", interval);
                    output_results(
                        converted_args,
                        hist_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error),
                        &mut aggregators,
                        &filters,
                    )?;
                }
                (name, _) => {
                    error!("Unsupported subcommand `{name}`")
                }
            }
        }
        Some(("locks", _)) => {
            filters.push(Box::new(crate::filters::LockingFilter::new()));
            output_results(converted_args, &Severity::Log, &mut aggregators, &filters)?;
        }
        Some(("system", _)) => {
            filters.push(Box::new(crate::filters::SystemFilter::new()));
            output_results(converted_args, &Severity::Log, &mut aggregators, &filters)?;
        }
        Some(("connections", _)) => {
            aggregators.push(Box::new(ConnectionsAggregator::new()));
            converted_args.print_details = false;
            output_results(converted_args, &Severity::Log, &mut aggregators, &filters)?;
        }
        Some(("peaks", _)) => {
            error!("Not implemented")
        }
        Some(("slow", sub_matches)) => {
            if let Some(("top", _)) = sub_matches.subcommand() {
                debug!("Using TopSlowQueryAggregator");
                aggregators.push(Box::new(TopSlowQueries::new(10)));
                converted_args.print_details = false;
                output_results(converted_args, &Severity::Log, &mut aggregators, &filters)?;
            } else {
                let mut treshold = Duration::from_secs(3);
                if let Some(treshold_str) = sub_matches.get_one::<String>("TRESHOLD") {
                    treshold = parse_duration(treshold_str)?;
                }
                filters.push(Box::new(FilterSlow::new(treshold)));
                debug!("Using FilterSlow with treshold {:?}", treshold);
                output_results(converted_args, &Severity::Log, &mut aggregators, &filters)?;
            }
        }
        Some(("stats", _)) => {
            error!("Not implemented")
        }
        _ => error!("command not found"),
    }

    Ok(())
}
