use std::{
    fs::{self, File},
    io::copy,
    path::{Path, PathBuf},
};

use chrono::{DateTime, Local};
use clap::ArgMatches;
use flate2::read::GzDecoder;
use log::debug;
use tempfile::TempDir;
use zip::ZipArchive;

use crate::{Error, util::time_or_interval_string_to_time};

use crate::Result;

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
    pub mask: Option<String>,
    pub verbose: bool,
    pub print_details: bool,
}

impl ConvertedArgs {
    pub fn parse_from_matches(val: ArgMatches) -> Result<Self> {
        // Parse begin / end flags
        let begin = if let Some(begin_str) = val.get_one::<String>("begin") {
            Some(time_or_interval_string_to_time(begin_str, None)?)
        } else {
            None
        };

        let end = if let Some(end_str) = val.get_one::<String>("end") {
            Some(time_or_interval_string_to_time(end_str, None)?)
        } else {
            None
        };

        let mask = val.get_one::<String>("mask").map(|s| s.to_owned());

        // Initialize logger based on verbose flag
        let mut verbose = false;
        env_logger::Builder::from_default_env()
            .filter_level(if val.get_flag("debug") {
                verbose = true;
                debug!("Running in debug mode.");
                log::LevelFilter::Debug
            } else {
                log::LevelFilter::Info
            })
            .init();

        Ok(ConvertedArgs {
            file_list: vec![],
            files: vec![],
            begin,
            end,
            mask,
            matches: val,
            verbose,
            print_details: true,
        })
    }

    pub fn expand_dirs(mut self) -> Result<Self> {
        if let Some((_, sub_matches)) = self.matches.subcommand() {
            let paths = sub_matches
                .get_many::<PathBuf>("PATH")
                .into_iter()
                .flatten()
                .collect::<Vec<_>>();
            let file_list_to_add = self.expand_paths_helper(paths)?;
            self.file_list.extend(file_list_to_add);
            if let Some((_, sub_sub_matches)) = sub_matches.subcommand() {
                let paths = sub_sub_matches
                    .get_many::<PathBuf>("PATH")
                    .into_iter()
                    .flatten()
                    .collect::<Vec<_>>();
                let file_list_to_add = self.expand_paths_helper(paths)?;
                self.file_list.extend(file_list_to_add);
            }
        }

        Ok(self)
    }

    fn expand_paths_helper(&self, path: Vec<&PathBuf>) -> Result<Vec<PathBuf>> {
        let mut result: Vec<PathBuf> = vec![];
        for p in path {
            if p.is_file() {
                result.push(p.to_path_buf());
            } else if p.is_dir() {
                debug!("Expanding directory: {:?}", p);
                for entry in fs::read_dir(p)? {
                    let entry = entry?;
                    let path = entry.path();
                    result.push(path);
                }
            } else {
                return Err(Error::FileDoesNotExist { path: p.clone() });
            }
        }
        Ok(result)
    }

    pub fn expand_archives(mut self) -> Result<Self> {
        let temp_dir = TempDir::new()?;

        for path in &self.file_list {
            match path.extension().and_then(|s| s.to_str()) {
                Some("gz") => self.files.push(extract_gz(path, temp_dir.path())?),
                Some("zip") => self.files.extend(extract_zip(path, temp_dir.path())?),

                Some(_r) => {
                    let file_with_path = FileWithPath {
                        file: File::open(path)?,
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

fn extract_gz(src: &Path, temp_dir: &Path) -> Result<FileWithPath> {
    let file = fs::File::open(src)?;
    let mut decoder = GzDecoder::new(file);

    let filename = src
        .file_stem()
        .ok_or(Error::FailedToExtractStemFromPath)?
        .to_string_lossy()
        .to_string();
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
