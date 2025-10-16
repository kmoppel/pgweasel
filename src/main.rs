use core::str;

use chrono::{DateTime, Local};
use clap::{Parser, Subcommand};
use log::{debug, error};

mod errors;
mod files;
mod logparser;
mod logreader;
mod postgres;
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

#[derive(Subcommand, Debug)]
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

#[derive(Subcommand, Debug)]
enum ErrorsSubcommands {
    /// Show the most frequent error messages with counts
    Top,
}

struct ConvertedArgs {
    begin: Option<DateTime<Local>>,
    end: Option<DateTime<Local>>,
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

    let converted_args = match util::convert_args(&cli) {
        Ok(args) => args,
        Err(e) => {
            error!("Error processing arguments: {}", e);
            std::process::exit(1);
        }
    };

    match &cli.command {
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
                    errors::process_errors(&cli, &converted_args, min_severity);
                }
            }
        }
    }
}
