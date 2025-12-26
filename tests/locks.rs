use assert_cmd::cargo;
use assert_cmd::prelude::*;
use std::process::Command;

#[test]
fn simple_csv_slow_filter() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["locks", "./tests/files/locking.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("2025-06-03 12:46:07.925"));

    Ok(())
}
