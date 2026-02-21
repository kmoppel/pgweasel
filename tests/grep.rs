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
