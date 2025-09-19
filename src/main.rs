use clap::{Parser, Subcommand};

mod logreader;

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
            
            // Use the logreader module to read the file line by line
            match logreader::getlines(&cli.filename) {
                Ok(lines) => {
                    for (line_number, line_result) in lines.enumerate() {
                        match line_result {
                            Ok(line) => {
                                if line.contains("ERROR: ") {
                                    println!("{}", line);
                                }
                            },
                            Err(e) => eprintln!("Error reading line {}: {}", line_number + 1, e),
                        }
                    }
                }
                Err(e) => eprintln!("Error opening file '{}': {}", cli.filename, e),
            }
        }
    }
}
