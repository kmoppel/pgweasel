use crate::{filters::Filter, format::Format};

#[derive(Clone)]
pub struct FilterContainsCi {
    substring_lower: String,
}

impl FilterContainsCi {
    pub fn new(substring: String) -> Self {
        FilterContainsCi {
            substring_lower: substring.to_lowercase(),
        }
    }
}

impl Filter for FilterContainsCi {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        let record_lower = record.to_ascii_lowercase();
        memchr::memmem::find(&record_lower, self.substring_lower.as_bytes()).is_some()
    }
}
