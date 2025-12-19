use assert_cmd::cargo;
use assert_cmd::prelude::*; // Add methods on commands
use std::process::Command; // Run programs

#[test]
fn base_help_with_options() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.arg("--help")
        .assert()
        .success()
        .stdout(predicates::str::contains("pgweasel [OPTIONS] <COMMAND>"));

    Ok(())
}

#[test]
fn errors_command_help() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["errors", "--help"])
        .assert()
        .success()
        .stdout(predicates::str::contains(
            "pgweasel errors [OPTIONS] <PATH>...",
        ));

    Ok(())
}

#[test]
fn errors_command_with_sub_command_help() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["errors", "list", "--help"])
        .assert()
        .success()
        .stdout(predicates::str::contains(
            "pgweasel errors list [OPTIONS] <PATH>...",
        ));

    Ok(())
}
