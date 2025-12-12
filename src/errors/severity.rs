use clap::{ValueEnum, builder::PossibleValue};

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, Debug)]
pub enum Severity {
    Debug5,
    Debug4,
    Debug3,
    Debug2,
    Debug1,
    Log,
    Info,
    Notice,
    Warning,
    Error,
    Fatal,
    Panic,
}

impl Severity {
    pub fn from_log_string(str: &str) -> Self {
        if str.contains("LOG:") {
            return Severity::Log;
        };
        if str.contains("ERROR:") {
            return Severity::Error;
        };
        if str.contains("INFO:") {
            return Severity::Info;
        };
        if str.contains("NOTICE:") {
            return Severity::Notice;
        };
        if str.contains("WARNING:") {
            return Severity::Warning;
        };
        if str.contains("DEBUG5:") {
            return Severity::Debug5;
        };
        if str.contains("DEBUG4:") {
            return Severity::Debug4;
        };
        if str.contains("DEBUG3:") {
            return Severity::Debug3;
        };
        if str.contains("DEBUG2:") {
            return Severity::Debug2;
        };
        if str.contains("DEBUG1:") {
            return Severity::Debug1;
        };
        if str.contains("FATAL:") {
            return Severity::Fatal;
        };
        if str.contains("PANIC:") {
            return Severity::Panic;
        };
        Severity::Log
    }
}

impl ValueEnum for Severity {
    fn value_variants<'a>() -> &'a [Self] {
        &[
            Severity::Debug5,
            Severity::Debug4,
            Severity::Debug3,
            Severity::Debug2,
            Severity::Debug1,
            Severity::Log,
            Severity::Info,
            Severity::Notice,
            Severity::Warning,
            Severity::Error,
            Severity::Fatal,
            Severity::Panic,
        ]
    }

    fn to_possible_value(&self) -> Option<clap::builder::PossibleValue> {
        Some(match self {
            Severity::Debug5 => PossibleValue::new("DEBUG5").help(""),
            Severity::Debug4 => PossibleValue::new("DEBUG4").help(""),
            Severity::Debug3 => PossibleValue::new("DEBUG3").help(""),
            Severity::Debug2 => PossibleValue::new("DEBUG2").help(""),
            Severity::Debug1 => PossibleValue::new("DEBUG1").help(""),
            Severity::Log => PossibleValue::new("LOG").help(""),
            Severity::Info => PossibleValue::new("INFO").help(""),
            Severity::Notice => PossibleValue::new("NOTICE").help(""),
            Severity::Warning => PossibleValue::new("WARNING").help(""),
            Severity::Error => PossibleValue::new("ERROR").help(""),
            Severity::Fatal => PossibleValue::new("FATAL").help(""),
            Severity::Panic => PossibleValue::new("PANIC").help(""),
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
            Severity::Debug5 => 0,
            Severity::Debug4 => 1,
            Severity::Debug3 => 2,
            Severity::Debug2 => 3,
            Severity::Debug1 => 4,
            Severity::Log => 5,
            Severity::Info => 5,
            Severity::Notice => 6,
            Severity::Warning => 7,
            Severity::Error => 8,
            Severity::Fatal => 9,
            Severity::Panic => 0,
        }
    }
}

// TODO: Check is it right to have backwards? and default.
impl From<String> for Severity {
    fn from(value: String) -> Self {
        match value.to_uppercase().as_str() {
            "DEBUG5" => Severity::Debug5,
            "DEBUG4" => Severity::Debug4,
            "DEBUG3" => Severity::Debug3,
            "DEBUG2" => Severity::Debug2,
            "DEBUG1" => Severity::Debug1,
            "LOG" => Severity::Log,
            "INFO" => Severity::Info,
            "NOTICE" => Severity::Notice,
            "WARNING" => Severity::Warning,
            "ERROR" => Severity::Error,
            "FATAL" => Severity::Fatal,
            "PANIC" => Severity::Panic,
            _ => Severity::Info, // Default to LOG level
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn from_log_string() {
        let sev1 = Severity::from_log_string("string :ERROR string");
        assert_eq!(Severity::Error, sev1);

        let sev2 = Severity::from_log_string(
            "2025-05-21 10:57:10.100 UTC [596]: [1-1] db=postgres,user=postgres,host=91.129.106.131 ERROR:  syntax error at or near \"sdaasdasda\" at character 12025-05-21 10:57:10.100 UTC [596]: [2-1] db=postgres,user=postgres,host=91.129.106.131 STATEMENT:  sdaasdasda",
        );
        assert_eq!(Severity::Error, sev2);
    }
}
