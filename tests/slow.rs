use assert_cmd::cargo;
use assert_cmd::prelude::*;
use predicates::prelude::PredicateBooleanExt; // Add methods on commands
use std::process::Command; // Run programs

#[test]
fn simple_csv_slow_filter() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["slow", "1s", "./tests/files/csvlog_pg14.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("duration: 2722.543 ms"));
}

#[test]
fn simple_log_slow_filter() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["slow", "25ms", "./tests/files/duration.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("statement: WITH RECURSIVE"));
}

#[test]
fn aggregate_top_slow() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["slow", "top", "./tests/files/duration.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("--- 25.761ms ---"));
}

#[test]
fn aggregate_top_slow_with_filter() {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-m",
        "2025-05-21 11:00:40",
        "slow",
        "top",
        "./tests/files/duration.log",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("025-05-21 11:01:10").not());
}
