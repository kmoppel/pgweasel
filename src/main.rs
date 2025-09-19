use core::str;

use chrono::{DateTime, Local};
use clap::{Parser, Subcommand};

mod errors;
mod logreader;
mod util;

/// A PostgreSQL log parser
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
#[command(subcommand_precedence_over_arg = false)] // Allow also subcommands before args as well
pub struct Cli {
    /// Input logfile path
    #[arg(global = true, required = false)]
    filename: Option<String>,

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

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Error command for testing
    #[command(visible_alias = "err")]
    #[command(visible_alias = "errs")]
    #[command(visible_alias = "error")]
    Errors,
}

struct ConvertedArgs {
    begin: Option<DateTime<Local>>,
    //    end: Option<DateTime<Local>>,
}

fn main() {
    let cli = Cli::parse();
    if cli.verbose {
        println!("Running in debug mode. Cmdline input: {cli:?}");
    }

    let converted_args = match util::convert_args(&cli) {
        Ok(args) => args,
        Err(e) => {
            eprintln!("Error processing arguments: {}", e);
            std::process::exit(1);
        }
    };

    match cli.command {
        Commands::Errors {} => {
            errors::process_errors(&cli, &converted_args);
        }
    }
}
