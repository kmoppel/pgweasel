use std::time::Duration;

use crate::{duration::extract_duration, filters::Filter};

#[derive(Clone)]
pub struct FilterSlow {
    treshold: Duration,
}

impl FilterSlow {
    pub fn new(treshold: Duration) -> Self {
        FilterSlow { treshold }
    }
}

impl Filter for FilterSlow {
    fn matches(&self, record: &[u8]) -> bool {
        let text = unsafe { std::str::from_utf8_unchecked(record) };
        if let Some(duration) = extract_duration(text) {
            if duration > self.treshold {
                return true;
            }
        }
        false
    }
}
