use crate::{filters::Filter, format::Format};

#[derive(Clone)]
pub struct FilterContains {
    substring: String,
}

impl FilterContains {
    pub fn new(substring: String) -> Self {
        FilterContains { substring }
    }
}

impl Filter for FilterContains {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        memchr::memmem::find(record, self.substring.as_bytes()).is_some()
    }
}
