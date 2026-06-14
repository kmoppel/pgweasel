use assert_cmd::cargo;
use assert_cmd::prelude::*;
use std::process::Command;

#[test]
fn simple_connection_aggregate() -> Result<(), Box<dyn std::error::Error>> {
    let mut cmd = Command::new(cargo::cargo_bin!("pgweasel"));

    cmd.args(["conn", "./tests/files/azure_connections.log"])
        .assert()
        .success()
        .stdout(predicates::str::contains("5  2025-05-21 11:00:00"));

    Ok(())
}

#[test]
fn connections_output_sorted_by_count() -> Result<(), Box<dyn std::error::Error>> {
    let output = Command::new(cargo::cargo_bin!("pgweasel"))
        .args(["conn", "./tests/files/connections.log"])
        .output()?;

    assert!(output.status.success());
    let stdout = String::from_utf8(output.stdout)?;

    let count_sorted_sections = [
        "Connections by host:",
        "Connections by database:",
        "Connections by user:",
        "Connections by application name:",
        "Connections by time bucket:",
    ];

    for section in count_sorted_sections {
        let section_start = stdout
            .find(section)
            .unwrap_or_else(|| panic!("Section '{section}' not found"));
        let after_header = &stdout[section_start + section.len()..];
        let section_end = after_header
            .find("\nConnections by ")
            .unwrap_or(after_header.len());
        let section_text = &after_header[..section_end];

        let counts: Vec<u16> = section_text
            .lines()
            .filter_map(|line| {
                let trimmed = line.trim();
                if trimmed.is_empty() {
                    return None;
                }
                trimmed.split_whitespace().next()?.parse::<u16>().ok()
            })
            .collect();

        assert!(
            counts.len() > 1,
            "Section '{section}' should have more than one entry for sorting to matter"
        );
        for i in 1..counts.len() {
            assert!(
                counts[i - 1] >= counts[i],
                "Section '{section}' not sorted descending: {counts:?}"
            );
        }
    }

    Ok(())
}
