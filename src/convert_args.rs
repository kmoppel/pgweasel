use chrono::{DateTime, Local};
use log::{debug, error};

use crate::{Cli, util::time_or_interval_string_to_time};

pub struct ConvertedArgs {
    pub cli: Cli,
    pub begin: Option<DateTime<Local>>,
    pub end: Option<DateTime<Local>>,
}

impl Into<ConvertedArgs> for Cli {
    fn into(self) -> ConvertedArgs {
        let begin = if let Some(begin_str) = &self.begin {
            match time_or_interval_string_to_time(begin_str, None) {
                Ok(datetime) => {
                    debug!(
                        "Parsed begin time: {}",
                        datetime.format("%Y-%m-%d %H:%M:%S %Z")
                    );
                    Some(datetime)
                }
                Err(e) => {
                    error!("Error processing arguments: {}", e);
                    std::process::exit(1);
                }
            }
        } else {
            None
        };

        let end = if let Some(end_str) = &self.end {
            match time_or_interval_string_to_time(end_str, None) {
                Ok(datetime) => {
                    debug!(
                        "Parsed end time: {}",
                        datetime.format("%Y-%m-%d %H:%M:%S %Z")
                    );
                    Some(datetime)
                }
                Err(e) => {
                    error!("Error processing arguments: {}", e);
                    std::process::exit(1);
                }
            }
        } else {
            None
        };

        ConvertedArgs {
            begin,
            end,
            cli: self,
        }
    }
}
