use assert_cmd::cargo;
use assert_cmd::prelude::*;
use predicates::prelude::PredicateBooleanExt;
use std::io::Write;
use std::process::Command;
use tempfile::Builder;

#[test]
fn simple_error_filter() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "./tests/files/csvlog1.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));
}

#[test]
fn simple_error_filter_for_log() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "./tests/files/debian_default2.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-22 15:15:09.392"));
}

#[test]
fn error_multiline_csv() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "./tests/files/multiple_lines.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-12-15 12:41:20.659"));
}

#[test]
fn simple_error_filter_with_begin_min() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    let now = chrono::Local::now();
    let now_str = now.format("%Y-%m-%d %H:%M:%S%.3f %Z").to_string();
    let content = format!("{now_str} ERROR:  error message example\n");

    let mut tmp = Builder::new().suffix(".log").tempfile()?;
    write!(tmp, "{content}")?;
    tmp.flush()?;

    cmd.args(["-b", "10m", "err", tmp.path().to_str().unwrap()])
        .assert()
        .success()
        .stdout(predicates::str::contains(
            now.format("%Y-%m-%d %H:%M:%S").to_string(),
        ));

    Ok(())
}

#[test]
fn simple_error_filter_with_begin_end() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-b",
        "2025-05-08 12:24:37.000 EEST",
        "-e",
        "2025-05-08 12:24:37.999 EEST",
        "err",
        "./tests/files/csvlog1.csv",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));
}

#[test]
fn simple_error_filter_with_mask() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-m",
        "2025-05-08 12:24:37",
        "err",
        "./tests/files/csvlog1.csv",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));
}

#[test]
fn simple_filter_with_list_subcommand() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "list", "./tests/files/csvlog1.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-05-08 12:24:37.731 EEST"));
}

#[test]
fn simple_filter_with_top_subcommand() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "top", "./tests/files/debian_default2.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains(
            "new row for relation \"pgbench_accounts\" violates check constraint \"posbal\"",
        ));
}

#[test]
fn simple_filter_with_top_max_subcommand() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "err",
        "top",
        "--max",
        "2",
        "./tests/files/debian_default2.log",
    ])
    .assert()
    .success()
    .stdout(
        predicates::str::contains("canceling statement due to user request")
            .not()
            .or(
                predicates::str::contains("terminating connection due to administrator command")
                    .not(),
            ),
    );
}

#[test]
fn simple_filter_with_top_max_subcommand_contains_same_max() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "err",
        "top",
        "--max",
        "2",
        "./tests/files/debian_default2.log",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("8  new row for relation"));
}

#[test]
fn simple_filter_with_hist_subcommand() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "err",
        "hist",
        "-b",
        "1s",
        "-l",
        "debug5",
        "./tests/files/debian_default2.log",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains(
        "[2025-05-22 15:18:10] ##################################---------------- 11",
    ));
}

#[test]
fn error_output_sorted_by_timestamp() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    // The records in debian_default2.log are deliberately out of order in the
    // file, so a correct sort must reorder them by leading timestamp.
    let output = cmd
        .args(["err", "./tests/files/debian_default2.log"])
        .assert()
        .success()
        .get_output()
        .stdout
        .clone();
    let stdout = String::from_utf8(output).unwrap();

    // Collect the leading "YYYY-MM-DD HH:MM:SS.mmm" timestamp of every record.
    let timestamps: Vec<&str> = stdout
        .lines()
        .filter(|line| line.len() >= 23 && line.as_bytes()[4] == b'-')
        .map(|line| &line[..23])
        .collect();

    assert!(
        timestamps.len() > 1,
        "expected multiple error records, got {}",
        timestamps.len()
    );

    let mut sorted = timestamps.clone();
    sorted.sort_unstable();
    assert_eq!(
        timestamps, sorted,
        "error output was not sorted by timestamp"
    );
}

#[test]
fn non_existing_file() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["err", "list", "./tests/files/csvlog1.csv_non_existing"])
        .assert()
        .failure()
        .stderr(predicates::str::contains("FileDoesNotExist"));
}

#[test]
fn panic_severity_included_in_error_output() -> Result<(), Box<dyn std::error::Error>> {
    // Regression test: PANIC is the most severe level in Postgres, but it was
    // ranked lower than DEBUG5, so `err` (min severity Error) silently dropped
    // every PANIC line.
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    let content = "2025-05-08 12:24:37.731 EEST [12345]:  5.18.4,pgweasel@localhost ERROR:  duplicate key value violates unique constraint \"pgbench_accounts_pkey\"\n2025-05-08 12:24:38.100 EEST [12345]:  5.18.4,pgweasel@localhost PANIC:  could not write to file \"base/16384/16385\"\n2025-05-08 12:24:39.100 EEST [12345]:  5.18.4,pgweasel@localhost FATAL:  terminating connection of process 12345\n";

    let mut tmp = Builder::new().suffix(".log").tempfile()?;
    write!(tmp, "{content}")?;
    tmp.flush()?;

    cmd.args(["err", tmp.path().to_str().unwrap()])
        .assert()
        .success()
        .stdout(predicates::str::contains("PANIC:  could not write to file"))
        .stdout(predicates::str::contains(
            "ERROR:  duplicate key value violates",
        ))
        .stdout(predicates::str::contains(
            "FATAL:  terminating connection of process",
        ));

    Ok(())
}
