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
}
