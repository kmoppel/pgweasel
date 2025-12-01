const VALID_SEVERITIES: &[&str] = &[
    "DEBUG5", "DEBUG4", "DEBUG3", "DEBUG2", "DEBUG1", "LOG", "INFO", "NOTICE", "WARNING", "ERROR",
    "FATAL", "PANIC",
];

/// Validate that a severity level string is valid
pub fn validate_severity(severity: &str) -> Result<(), String> {
    VALID_SEVERITIES
        .contains(&severity.to_uppercase().as_str())
        .then_some(())
        .ok_or_else(|| {
            format!(
                "Invalid severity level: '{}'. Valid values are: {}",
                severity,
                VALID_SEVERITIES.join(", ")
            )
        })
}

/// Convert PostgreSQL log severity level to a numeric priority
pub fn log_entry_severity_to_num(severity: &str) -> i32 {
    match severity.to_uppercase().as_str() {
        "DEBUG5" => 0,
        "DEBUG4" => 1,
        "DEBUG3" => 2,
        "DEBUG2" => 3,
        "DEBUG1" => 4,
        "LOG" => 5,
        "INFO" => 5,
        "NOTICE" => 6,
        "WARNING" => 7,
        "ERROR" => 8,
        "FATAL" => 9,
        "PANIC" => 10,
        _ => 5, // Default to LOG level
    }
}
