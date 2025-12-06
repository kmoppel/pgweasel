use clap::{ValueEnum, builder::PossibleValue};

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub enum Severity {
    DEBUG5,
    DEBUG4,
    DEBUG3,
    DEBUG2,
    DEBUG1,
    LOG,
    INFO,
    NOTICE,
    WARNING,
    ERROR,
    FATAL,
    PANIC,
}

impl ValueEnum for Severity {
    fn value_variants<'a>() -> &'a [Self] {
        &[
            Severity::DEBUG5,
            Severity::DEBUG4,
            Severity::DEBUG3,
            Severity::DEBUG2,
            Severity::DEBUG1,
            Severity::LOG,
            Severity::INFO,
            Severity::NOTICE,
            Severity::WARNING,
            Severity::ERROR,
            Severity::FATAL,
            Severity::PANIC,
        ]
    }

    fn to_possible_value(&self) -> Option<clap::builder::PossibleValue> {
        Some(match self {
            Severity::DEBUG5 => PossibleValue::new("DEBUG5").help(""),
            Severity::DEBUG4 => PossibleValue::new("DEBUG4").help(""),
            Severity::DEBUG3 => PossibleValue::new("DEBUG3").help(""),
            Severity::DEBUG2 => PossibleValue::new("DEBUG2").help(""),
            Severity::DEBUG1 => PossibleValue::new("DEBUG1").help(""),
            Severity::LOG => PossibleValue::new("LOG").help(""),
            Severity::INFO => PossibleValue::new("INFO").help(""),
            Severity::NOTICE => PossibleValue::new("NOTICE").help(""),
            Severity::WARNING => PossibleValue::new("WARNING").help(""),
            Severity::ERROR => PossibleValue::new("ERROR").help(""),
            Severity::FATAL => PossibleValue::new("FATAL").help(""),
            Severity::PANIC => PossibleValue::new("PANIC").help(""),
        })
    }
}

impl std::fmt::Display for Severity {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        self.to_possible_value()
            .expect("no values are skipped")
            .get_name()
            .fmt(f)
    }
}

impl std::str::FromStr for Severity {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        for variant in Self::value_variants() {
            if variant.to_possible_value().unwrap().matches(s, false) {
                return Ok(*variant);
            }
        }
        Err(format!("invalid variant: {s}"))
    }
}

impl Into<i32> for &Severity {
    fn into(self) -> i32 {
        match self {
            Severity::DEBUG5 => 0,
            Severity::DEBUG4 => 1,
            Severity::DEBUG3 => 2,
            Severity::DEBUG2 => 3,
            Severity::DEBUG1 => 4,
            Severity::LOG => 5,
            Severity::INFO => 5,
            Severity::NOTICE => 6,
            Severity::WARNING => 7,
            Severity::ERROR => 8,
            Severity::FATAL => 9,
            Severity::PANIC => 0,
        }
    }
}

// TODO: Check is it right to have backwards? and default.
impl From<String> for Severity {
    fn from(value: String) -> Self {
        match value.to_uppercase().as_str() {
            "DEBUG5" => Severity::DEBUG5,
            "DEBUG4" => Severity::DEBUG4,
            "DEBUG3" => Severity::DEBUG3,
            "DEBUG2" => Severity::DEBUG2,
            "DEBUG1" => Severity::DEBUG1,
            "LOG" => Severity::LOG,
            "INFO" => Severity::INFO,
            "NOTICE" => Severity::NOTICE,
            "WARNING" => Severity::WARNING,
            "ERROR" => Severity::ERROR,
            "FATAL" => Severity::FATAL,
            "PANIC" => Severity::PANIC,
            _ => Severity::INFO, // Default to LOG level
        }
    }
}
