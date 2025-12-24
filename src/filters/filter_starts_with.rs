use crate::filters::Filter;

#[derive(Clone)]
pub struct FilterStartsWith {
    substring: String,
}

impl FilterStartsWith {
    pub fn new(substring: String) -> Self {
        FilterStartsWith { substring }
    }
}

impl Filter for FilterStartsWith {
    fn matches(&self, record: &[u8]) -> bool {
        record.starts_with(self.substring.as_bytes())
    }
}
