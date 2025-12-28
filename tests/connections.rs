use assert_cmd::cargo;
use assert_cmd::prelude::*;
use std::process::Command;

#[test]
fn simple_connection_aggregate() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["conn", "./tests/files/azure_connections.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("5  2025-05-21 11:00:00 +03:00"));

    Ok(())
}
