use assert_cmd::cargo;
use assert_cmd::prelude::*; // Add methods on commands
use std::process::Command; // Run programs

#[test]
fn simple_error_filter() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["slow", "100", "./testdata/csvlog_pg14.csv"])
        .assert()
        .success()
        .stdout(predicates::str::contains("duration: 339.050 ms"));

    Ok(())
}
