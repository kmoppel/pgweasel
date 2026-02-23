use assert_cmd::cargo;
use assert_cmd::prelude::*;
use std::process::Command;

#[test]
fn grep_finds_matching_lines() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["grep", "dadasd", "./tests/files/debian_default.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("dasda"));

    Ok(())
}

#[test]
fn grep_is_case_insensitive() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["grep", "DADASD", "./tests/files/debian_default.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("dasda"));

    Ok(())
}

#[test]
fn grep_after_context() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-A",
        "1",
        "grep",
        "dadasd",
        "./tests/files/debian_default.log",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("terminating background worker"));

    Ok(())
}

#[test]
fn grep_before_context() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args([
        "-B",
        "1",
        "grep",
        "dadasd",
        "./tests/files/debian_default.log",
    ])
    .assert()
    .success()
    .stdout(predicates::str::contains("syntax error"));

    Ok(())
}
