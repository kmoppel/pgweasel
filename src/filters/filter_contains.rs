use crate::filters::Filter;

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
    fn matches(&self, record: &[u8]) -> bool {
        memchr::memmem::find(record, self.substring.as_bytes()).is_some()
    }
}
