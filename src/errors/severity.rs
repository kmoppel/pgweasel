use clap::{ValueEnum, builder::PossibleValue};

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, Debug)]
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

impl Severity {
    pub fn from_log_string(str: &str) -> Self {
        if str.contains("LOG:") {
            return Severity::LOG;
        };
        if str.contains("ERROR:") {
            return Severity::ERROR;
        };
        if str.contains("INFO:") {
            return Severity::INFO;
        };
        if str.contains("NOTICE:") {
            return Severity::NOTICE;
        };
        if str.contains("WARNING:") {
            return Severity::WARNING;
        };
        if str.contains("DEBUG5:") {
            return Severity::DEBUG5;
        };
        if str.contains("DEBUG4:") {
            return Severity::DEBUG4;
        };
        if str.contains("DEBUG3:") {
            return Severity::DEBUG3;
        };
        if str.contains("DEBUG2:") {
            return Severity::DEBUG2;
        };
        if str.contains("DEBUG1:") {
            return Severity::DEBUG1;
        };
        if str.contains("FATAL:") {
            return Severity::FATAL;
        };
        if str.contains("PANIC:") {
            return Severity::PANIC;
        };
        Severity::LOG
    }
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

impl From<&Severity> for i32 {
    fn from(val: &Severity) -> Self {
        match val {
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn from_log_string() {
        let sev1 = Severity::from_log_string("string :ERROR string");
        assert_eq!(Severity::ERROR, sev1);

        let sev2 = Severity::from_log_string("2025-05-21 10:57:10.100 UTC [596]: [1-1] db=postgres,user=postgres,host=91.129.106.131 ERROR:  syntax error at or near \"sdaasdasda\" at character 12025-05-21 10:57:10.100 UTC [596]: [2-1] db=postgres,user=postgres,host=91.129.106.131 STATEMENT:  sdaasdasda");
        assert_eq!(Severity::ERROR, sev2);
    }
}
