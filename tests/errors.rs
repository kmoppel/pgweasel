use assert_cmd::cargo;
use assert_cmd::prelude::*;
use std::io::Write;
use std::process::Command;
use tempfile::Builder;

#[test]
fn simple_error_filter() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "./testdata/csvlog1.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));

    Ok(())
}

#[test]
fn simple_error_filter_with_begin_min() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    let now = chrono::Local::now();
    let now_str = now.format("%Y-%m-%d %H:%M:%S%.3f %Z").to_string();
    let content = format!("{now_str} ERROR:  error message example\n");

    let mut tmp = Builder::new().suffix(".log").tempfile()?;
    write!(tmp, "{}", content)?;
    tmp.flush()?;

    cmd.args(["-b", "1m", "err", tmp.path().to_str().unwrap()])
        .assert()
        .success()
        .stdout(predicates::str::contains(
            now.format("%Y-%m-%d %H:%M:%S").to_string(),
        ));

    Ok(())
}

#[test]
fn simple_error_filter_with_begin_end() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-b",
        "2025-05-08 12:24:37.000 EEST",
        "-e",
        "2025-05-08 12:24:37.999 EEST",
        "err",
        "./testdata/csvlog1.csv",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));

    Ok(())
}

#[test]
fn simple_error_filter_with_mask() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["-m", "2025-05-08 12:24:37", "err", "./testdata/csvlog1.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));

    Ok(())
}

#[test]
fn simple_filter_with_list_subcommand() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "list", "./testdata/csvlog1.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));

    Ok(())
}

#[test]
fn non_existing_file() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "list", "./testdata/csvlog1.cs"])
        .assert()
        .failure()
        .stderr(predicates::str::contains("FileDoesNotExist"));

    Ok(())
}
