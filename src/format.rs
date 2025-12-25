use crate::severity::Severity;

pub enum Format {
    Csv,
    Plain,
}

impl Format {
    pub fn from_file_extension(file_name: &str) -> Self {
        if file_name.ends_with(".csv") {
            Format::Csv
        } else {
            Format::Plain
        }
    }

    pub fn severity_from_string(&self, text: &str) -> Severity {
        match self {
            Format::Csv => Severity::from_csv_string(text),
            Format::Plain => Severity::from_log_string(text),
        }
    }
}
