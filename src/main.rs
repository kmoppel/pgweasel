use core::str;

use clap::{Parser, Subcommand};
use log::{debug, error};

mod convert_args;
mod errors;
// Comented out to not get warnings on dead code
// mod files;
// mod logparser;
// mod logreader;
// mod postgres;
mod util;

/// A PostgreSQL log parser
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
#[command(subcommand_precedence_over_arg = false)] // Allow also subcommands before args as well
pub struct Cli {
    /// Input logfile paths (files or directories)
    #[arg(global = true, required = false)]
    input_files: Vec<String>,

    /// Verbose. Show debug information
    #[arg(
        global = true,
        short = 'v',
        long = "debug",
        required = false,
        default_value_t = false
    )]
    verbose: bool,

    /// Postgres log timestamp mask (e.g. "2025-05-21 12:57" - will show all events at 12:57)
    #[arg(global = true, short = 't', long = "timestamp-mask", required = false)]
    timestamp_mask: Option<String>,

    #[arg(
        global = true,
        short = 'b',
        long = "begin",
        required = false,
        value_name = "[ 10min | '2025-09-01 12:00' ]"
    )]
    begin: Option<String>,

    #[arg(
        global = true,
        short = 'e',
        long = "end",
        required = false,
        value_name = "[ 10min | '2025-09-02 13:00' ]"
    )]
    end: Option<String>,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Clone, Debug)]
enum Commands {
    /// Show or summarize error messages
    #[command(visible_alias = "err")]
    #[command(visible_alias = "errs")]
    #[command(visible_alias = "error")]
    Errors {
        /// Postgres log levels are DEBUG[5-1], INFO, NOTICE, WARNING, ERROR, LOG, FATAL, PANIC
        #[arg(short = 'l', long = "min-severity", default_value = "WARNING")]
        min_severity: String,

        #[command(subcommand)]
        subcommand: Option<ErrorsSubcommands>,
    },
}

#[derive(Subcommand, Clone, Debug)]
enum ErrorsSubcommands {
    /// Show the most frequent error messages with counts
    Top,
}

fn main() {
    let cli = Cli::parse();

    // Initialize logger based on verbose flag
    env_logger::Builder::from_default_env()
        .filter_level(if cli.verbose {
            log::LevelFilter::Debug
        } else {
            log::LevelFilter::Info
        })
        .init();

    debug!("Running in debug mode. Cmdline input: {cli:?}");

    match &cli.command.clone() {
        Commands::Errors {
            min_severity,
            subcommand,
        } => {
            // Validate min_severity early
            if let Err(e) = errors::validate_severity(min_severity) {
                error!("{}", e);
                std::process::exit(1);
            }

            match subcommand {
                Some(ErrorsSubcommands::Top) => {
                    println!("hello top errors. min_severity = {}", min_severity);
                }
                None => {
                    errors::process_errors(&cli.into(), min_severity);
                }
            }
        }
    }
}
