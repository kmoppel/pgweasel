use core::str;
use std::path::PathBuf;

use clap::{Parser, Subcommand};
use log::{debug, error};

// use crate::convert_args::ConvertedArgs;

mod cli;
// mod convert_args;
// mod errors;
// Comented out to not get warnings on dead code
// mod files;
// mod logparser;
// mod logreader;
// mod postgres;
mod util;

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

fn main() -> Result<()> {
    let cli = cli::cli();
    let matches = cli.clone().get_matches();

    // Initialize logger based on verbose flag
    env_logger::Builder::from_default_env()
        .filter_level(if matches.get_flag("debug") {
            log::LevelFilter::Debug
        } else {
            log::LevelFilter::Info
        })
        .init();

    debug!("Running in debug mode.");

    match matches.subcommand() {
        Some(("error", sub_matches)) => {
            let paths = sub_matches
                .get_many::<PathBuf>("PATH")
                .into_iter()
                .flatten()
                .collect::<Vec<_>>();
            println!("Analyzing for error {paths:?}");
        }
        _ => unreachable!(),
    }

    // let mut processed_cli: ConvertedArgs = cli.clone().into();
    // processed_cli = processed_cli.expand_dirs()?.expand_archives()?;

    // match &cli.command {
    //     Commands::Errors {
    //         min_severity,
    //         subcommand,
    //         #[allow(unused)]
    //         input_files,
    //     } => {
    //         // Validate min_severity early
    //         if let Err(e) = errors::validate_severity(min_severity) {
    //             error!("{}", e);
    //             std::process::exit(1);
    //         }

    //         match subcommand {
    //             Some(ErrorsSubcommands::Top) => {
    //                 println!("hello top errors. min_severity = {}", min_severity);
    //             }
    //             None => {
    //                 errors::process_errors(&processed_cli, min_severity);
    //             }
    //         }
    //     }
    // };
    Ok(())
}
