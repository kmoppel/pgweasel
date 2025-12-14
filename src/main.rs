use log::error;

use crate::{convert_args::ConvertedArgs, errors::process_errors, severity::Severity};

mod cli;
mod convert_args;
mod error;
mod errors;
mod parsers;
mod severity;
mod util;

pub use self::error::{Error, Result};

fn main() -> Result<()> {
    let cli = cli::cli();
    let matches = cli.clone().get_matches();

    let mut converted_args: ConvertedArgs = matches.clone().into();
    converted_args = converted_args.expand_dirs()?.expand_archives()?;

    match matches.subcommand() {
        Some(("error", sub_matches)) => {
            let error_command = sub_matches.subcommand().unwrap_or(("list", sub_matches));
            match error_command {
                ("list", list_subcommand) => {
                    process_errors(
                        converted_args,
                        list_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::Error),
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
        Some(("slow", _)) => {
            error!("Not implemented")
        }
        Some(("stats", _)) => {
            error!("Not implemented")
        }
        _ => error!("command not found"),
    }

    Ok(())
}
