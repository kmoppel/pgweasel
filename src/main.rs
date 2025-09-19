use clap::{Parser, Subcommand};

/// A PostgreSQL log parser
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
#[command(subcommand_precedence_over_arg = false)]  // Allow also subcommands before args as well
struct Cli {
    /// Input logfile path
    #[arg(global = true, required = false)]
    filename: String,

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
    println!("{cli:?}");

    match cli.command {
        Commands::Errors {} => {
            println!("Parsing file: {}", cli.filename);
        }
    }
}
