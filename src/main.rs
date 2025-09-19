use clap::{Parser, Subcommand};

mod errors;
mod logreader;
mod util;

/// A PostgreSQL log parser
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
#[command(subcommand_precedence_over_arg = false)] // Allow also subcommands before args as well
struct Cli {
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

fn main() {
    let cli = Cli::parse();
    if cli.verbose {
        println!("Running in debug mode. Cmdline input: {cli:?}");
    }

    // Test the time parsing function if begin parameter is provided
    if cli.begin.is_some() {
        if let Some(begin_str) = &cli.begin {
            match util::time_or_interval_string_to_time(begin_str, None) {
                Ok(datetime) => {
                    println!(
                        "Parsed begin time: {}",
                        datetime.format("%Y-%m-%d %H:%M:%S %Z")
                    );
                }
                Err(e) => {
                    eprintln!("Error parsing begin time '{}': {}", begin_str, e);
                    std::process::exit(1);
                }
            }
        }
    }

    match cli.command {
        Commands::Errors {} => {
            // Pass the filename as Option<&str>
            let filename_ref = cli.filename.as_deref();
            errors::process_errors(filename_ref, cli.verbose);
        }
    }
}
