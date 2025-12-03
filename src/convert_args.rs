use std::{
    fs,
    io::copy,
    path::{Path, PathBuf},
};

use chrono::{DateTime, Local};
use flate2::read::GzDecoder;
use log::{debug, error};
use tempfile::TempDir;
use zip::ZipArchive;

use crate::{Cli, util::time_or_interval_string_to_time};

pub type Result<T> = core::result::Result<T, Error>;
pub type Error = Box<dyn std::error::Error>;

pub struct ConvertedArgs {
    pub cli: Cli,
    pub file_list: Vec<PathBuf>,
    pub begin: Option<DateTime<Local>>,
    pub end: Option<DateTime<Local>>,
}

impl ConvertedArgs {
    pub fn expand_dirs(mut self) -> Result<Self> {
        for p_str in &self.cli.input_files {
            let p = Path::new(p_str);
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

        Ok(self)
    }

    pub fn expand_archives(mut self) -> Result<Self> {
        let temp_dir = TempDir::new()?;
        let mut out_files: Vec<PathBuf> = vec![];

        for f in &self.file_list {
            match f.extension().and_then(|s| s.to_str()) {
                Some("gz") => out_files.extend(extract_gz(&f, temp_dir.path())?),
                Some("zip") => out_files.extend(extract_zip(&f, temp_dir.path())?),
                Some(_r) => out_files.push(f.clone()),
                None => {}
            }
        }
        self.file_list = out_files;

        Ok(self)
    }
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
            file_list: vec![],
            begin,
            end,
            cli: self,
        }
    }
}

fn extract_gz(src: &Path, temp_dir: &Path) -> Result<Vec<PathBuf>> {
    let file = fs::File::open(src)?;
    let mut decoder = GzDecoder::new(file);

    let filename = src.file_stem().unwrap().to_string_lossy().to_string();
    let out_path = temp_dir.join(filename);

    let mut out_file = fs::File::create(&out_path)?;
    copy(&mut decoder, &mut out_file)?;

    Ok(vec![out_path])
}

fn extract_zip(src: &Path, temp_dir: &Path) -> Result<Vec<PathBuf>> {
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

        out_files.push(out_path);
    }

    Ok(out_files)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_zip_list() -> Result<()> {
        // TODO Add checks for expanded list to have appropriate file names and do not contain archive names
        let cli: Cli = Cli {
            input_files: vec![],
            verbose: false,
            timestamp_mask: None,
            begin: None,
            end: None,
            command: crate::Commands::Errors {
                min_severity: "F".to_string(),
                subcommand: None,
            },
        };
        let mut convert_args: ConvertedArgs = cli.into();
        convert_args
            .cli
            .input_files
            .push("./testdata/pgbadger".to_string());
        let convert_args = convert_args.expand_dirs()?.expand_archives()?;

        println!("File list: {:?}", convert_args.file_list);

        Ok(())
    }
}
