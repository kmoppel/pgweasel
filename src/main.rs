use log::error;

use crate::{convert_args::ConvertedArgs, errors::{Severity, process_errors}};

// use crate::convert_args::ConvertedArgs;

mod cli;
mod convert_args;
mod parsers;
mod errors;
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

    let mut converted_args: ConvertedArgs = matches.clone().into();
    converted_args = converted_args.expand_dirs()?.expand_archives()?;

    match matches.subcommand() {
        Some(("error", sub_matches)) => {
            let error_command = sub_matches.subcommand().unwrap_or(("list", sub_matches));
            match error_command {
                ("list", list_subcommand) => {
                    process_errors(
                        &converted_args,
                        list_subcommand
                            .get_one::<Severity>("level")
                            .unwrap_or(&Severity::ERROR),
                    );
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
