use assert_cmd::cargo;
use assert_cmd::prelude::*;
use predicates::prelude::PredicateBooleanExt;
use std::process::Command; // Run programs

#[test]
fn simple_log_system() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["system", "./tests/files/system_test.log"])
        .assert()
        .success()
        .stdout(
            predicates::str::contains("listening").and(predicates::str::contains("was shut down")),
        );
    Ok(())
}
