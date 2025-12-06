use std::{
    fs::{self, File},
    io::copy,
    path::{Path, PathBuf},
};

use chrono::{DateTime, Local};
use clap::ArgMatches;
use flate2::read::GzDecoder;
use log::{debug, error};
use tempfile::TempDir;
use zip::ZipArchive;

use crate::util::time_or_interval_string_to_time;

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

pub struct FileWithPath {
    pub file: std::fs::File,
    pub path: std::path::PathBuf,
}

pub struct ConvertedArgs {
    pub matches: ArgMatches,
    pub file_list: Vec<PathBuf>,
    pub files: Vec<FileWithPath>,
    pub begin: Option<DateTime<Local>>,
    pub end: Option<DateTime<Local>>,
    pub verbose: bool,
}

impl ConvertedArgs {
    pub fn expand_dirs(mut self) -> Result<Self> {
        if let Some((_, sub_matches)) = self.matches.subcommand() {
            let paths = sub_matches
                .get_many::<PathBuf>("PATH")
                .into_iter()
                .flatten()
                .collect::<Vec<_>>();
            for p in paths {
                if p.is_file() {
                    self.file_list.push(p.to_path_buf());
                } else if p.is_dir() {
                    for entry in fs::read_dir(p)? {
                        let entry = entry?;
                        let path = entry.path();
                        self.file_list.push(path);
                    }
                }
            }
        }

        Ok(self)
    }

    pub fn expand_archives(mut self) -> Result<Self> {
        let temp_dir = TempDir::new()?;

        for path in &self.file_list {
            match path.extension().and_then(|s| s.to_str()) {
                Some("gz") => self.files.push(extract_gz(&path, temp_dir.path())?),
                Some("zip") => self.files.extend(extract_zip(&path, temp_dir.path())?),

                Some(_r) => {
                    let file_with_path = FileWithPath {
                        file: File::open(&path)?,
                        path: path.clone(),
                    };
                    self.files.push(file_with_path)
                }
                None => {}
            }
        }

        Ok(self)
    }
}

impl Into<ConvertedArgs> for ArgMatches {
    fn into(self) -> ConvertedArgs {
        // Parse begin / end flags
        let begin = if let Some(begin_str) = &self.get_one::<String>("begin") {
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

        let end = if let Some(end_str) = &self.get_one::<String>("end") {
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

        // Initialize logger based on verbose flag
        let mut verbose = false;
        env_logger::Builder::from_default_env()
            .filter_level(if self.get_flag("debug") {
                verbose = true;
                debug!("Running in debug mode.");
                log::LevelFilter::Debug
            } else {
                log::LevelFilter::Info
            })
            .init();

        ConvertedArgs {
            file_list: vec![],
            files: vec![],
            begin,
            end,
            matches: self,
            verbose,
        }
    }
}

fn extract_gz(src: &Path, temp_dir: &Path) -> Result<FileWithPath> {
    let file = fs::File::open(src)?;
    let mut decoder = GzDecoder::new(file);

    let filename = src.file_stem().unwrap().to_string_lossy().to_string();
    let out_path = temp_dir.join(filename);

    let mut out_file = fs::File::create(&out_path)?;
    copy(&mut decoder, &mut out_file)?;
    let reopened = fs::File::open(&out_path)?;

    Ok(FileWithPath {
        file: reopened,
        path: out_path,
    })
}

fn extract_zip(src: &Path, temp_dir: &Path) -> Result<Vec<FileWithPath>> {
    let file = fs::File::open(src)?;
    let mut archive = ZipArchive::new(file)?;

    let mut out_files = Vec::new();

    for i in 0..archive.len() {
        let mut zip_file = archive.by_index(i)?;
        if zip_file.is_dir() {
            continue;
        }

        let out_path = temp_dir.join(zip_file.name());

        if let Some(parent) = out_path.parent() {
            fs::create_dir_all(parent)?;
        }

        let mut out_file = fs::File::create(&out_path)?;
        copy(&mut zip_file, &mut out_file)?;

        let reopened = fs::File::open(&out_path)?;

        out_files.push(FileWithPath {
            file: reopened,
            path: out_path,
        });
    }

    Ok(out_files)
}

// #[cfg(test)]
// mod tests {
//     use super::*;

//     #[test]
//     fn test_zip_list() -> Result<()> {
//         let mut input_files: Vec<String> = vec![];
//         input_files.push("./testdata/pgbadger".to_string());
//         let cli: Cli = Cli {
//             verbose: false,
//             timestamp_mask: None,
//             begin: None,
//             end: None,
//             command: crate::Commands::Errors {
//                 min_severity: "F".to_string(),
//                 subcommand: None,
//                 input_files: vec![],
//             },
//         };
//         let convert_args: ConvertedArgs = cli.into();
//         let convert_args = convert_args.expand_dirs()?.expand_archives()?;

//         // TODO Add checks for expanded list to have appropriate file names and do not contain archive names
//         println!("File list: {:?}", convert_args.file_list);

//         Ok(())
//     }
// }
