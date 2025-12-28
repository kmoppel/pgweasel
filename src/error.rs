use std::{num::ParseIntError, path::PathBuf};

use derive_more::From;

pub type Result<T> = core::result::Result<T, Error>;

#[derive(Debug, From)]
pub enum Error {
    // Enable this if you want custom errors (also see region below)
    #[from]
    Custom(String),

    // -- Aggregators
    NotAbleToExtractMessage {
        record: String,
    },

    // -- ConvertArgs
    FileDoesNotExist {
        path: PathBuf,
    },
    FiledToParseBeginEnd {
        input: String,
        source: crate::util::TimeParseError,
    },

    // -- Parsers
    JsonNotYetImplemented,
    FileHasNoExtension {
        file: PathBuf,
    },
    FileExtensionIsNotSupported {
        file: PathBuf,
    },

    // -- LogParser
    FailedToRead {
        error: std::io::Error,
    },
    FailedToParseCsv {
        error: csv::Error,
    },
    NotAbleToReadRecordFromCsvLine,
    FailedToReadFirstRecord,
    FailedToParseCsvToPostgres {
        error: csv::Error,
    },

    // -- Externals
    #[from]
    Io(std::io::Error),

    #[from]
    ParseInt(ParseIntError),

    #[from]
    Zip(zip::result::ZipError),

    #[from]
    TimeParseError(crate::util::TimeParseError),

    #[from]
    DurationError(humantime::DurationError),
}

// region:    --- Custom --- Uncomment if want custom errors

impl Error {
    pub fn custom(val: impl std::fmt::Display) -> Self {
        Self::Custom(val.to_string())
    }
}

impl From<&str> for Error {
    fn from(val: &str) -> Self {
        Self::Custom(val.to_string())
    }
}

// endregion: --- Custom

// region:    --- Error Boilerplate

impl core::fmt::Display for Error {
    fn fmt(&self, fmt: &mut core::fmt::Formatter) -> core::result::Result<(), core::fmt::Error> {
        write!(fmt, "{self:?}")
    }
}

impl std::error::Error for Error {}

// endregion: --- Error Boilerplate
