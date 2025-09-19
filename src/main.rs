use clap::{Parser, Subcommand};

mod logreader;

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
        println!("{cli:?}");
    }

    match cli.command {
        Commands::Errors {} => {
            // Check if filename was provided
            let filename = match &cli.filename {
                Some(f) => f,
                None => {
                    eprintln!("Error: filename is required for the errors command");
                    std::process::exit(1);
                }
            };

            if cli.verbose {
                println!("Parsing file: {}", filename);
            }

            // Use the logreader module to read the file line by line
            match logreader::getlines(filename) {
                Ok(lines) => {
                    for (line_number, line_result) in lines.enumerate() {
                        match line_result {
                            Ok(line) => {
                                if line.contains("ERROR: ") {
                                    println!("{}", line);
                                }
                            }
                            Err(e) => eprintln!("Error reading line {}: {}", line_number + 1, e),
                        }
                    }
                }
                Err(e) => eprintln!("Error opening file '{}': {}", filename, e),
            }
        }
    }
}
